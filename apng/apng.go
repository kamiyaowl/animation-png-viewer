package apng

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"math"
	"os"
	"reflect"
)

type ColorType uint8

const (
	GrayScale          ColorType = 0
	TrueColor          ColorType = 2
	IndexColor         ColorType = 3
	GrayScaleWithAlpha ColorType = 4
	TrueColorWithAlpha ColorType = 6
)

// scanlineごとにある
type FilterType uint8

const (
	None FilterType = iota
	Sub
	Up
	Average
	Paeth
)

type Apng struct {
	Ihdr   Ihdr
	Idat   Idat
	Fctl   []Fctl
	Fdat   []Fdat
	Actl   Actl
	IsApng bool
}
type Ihdr struct {
	Width     int
	Height    int
	BitDepth  uint8
	ColorType uint8
	Compress  uint8
	Filter    uint8
	Interlace uint8
}
type Idat []uint8

// Animation Control
type Actl struct {
	NumFrames uint32
	NumPlays  uint32 // 0が指定されたら無限ループ
}

type DisposeOp uint8

const (
	OpNone       DisposeOp = iota
	OpBackground           // 透明な黒で上書き
	OpPrevious             // 次のフレームに映るときに、前のフレームの状態に戻す
)

type BlendOp uint8

const (
	OpSource BlendOp = iota // バッファに上書き
	OpOver                  // アルファブレンディング合成する
)

// Frame Control
type Fctl struct {
	SequenceNumber uint32
	Width          uint32
	Height         uint32
	OffsetX        uint32
	OffsetY        uint32
	DelayNum       uint16 // Frame Delayの分子
	DelayDen       uint16 // Frame Delayの分母
	DisposeOp      uint8
	BlendOp        uint8
}

// FrameData, sequence numberは配列で管理した時のインデックスにする
type Fdat struct {
	SequenceNumber uint32
	FrameData      Idat
}

// アニメーション生成用
type AnimationData struct {
	Image        image.Image
	DelaySeconds float32
}

func BytePerPixel(colorType ColorType) (uint8, error) {
	switch colorType {
	case GrayScale:
		return 1, nil
	case TrueColor:
		return 3, nil
	case IndexColor:
		return 1, nil
	case GrayScaleWithAlpha:
		return 2, nil
	case TrueColorWithAlpha:
		return 4, nil
	default:
		return 0, errors.New("ColorTypeが正しくない")
	}
}
func (self *Apng) BytePerPixel() (uint8, error) {
	return BytePerPixel(ColorType(self.Ihdr.ColorType))
}

// 一番親しい色を加算する
func paethPredictor(target byte, top byte, left byte) byte {
	p := float64(int(target) + int(top) - int(left))
	pTarget := math.Abs(p - float64(target))
	pTop := math.Abs(p - float64(top))
	pLeft := math.Abs(p - float64(left))
	if pTarget <= pTop && pTarget <= pLeft {
		return target
	} else if pTop <= pLeft {
		return top
	} else {
		return left
	}
}

// pngの圧縮用フィルタを解除します
func cancelFilter(targetValue byte, filterType FilterType, topPixelValue byte, leftPixelValue byte) (byte, error) {
	switch filterType {
	case None:
		return targetValue, nil
	case Sub:
		data := byte((int(targetValue) + int(leftPixelValue)) % 256)
		return data, nil
	case Up:
		data := byte((int(targetValue) + int(topPixelValue)) % 256)
		return data, nil
	case Average:
		avg := (int(leftPixelValue) + int(topPixelValue)) / 2
		data := byte((int(targetValue) + avg) % 256)
		return data, nil
	case Paeth:
		paeth := paethPredictor(targetValue, topPixelValue, leftPixelValue)
		data := byte((int(targetValue) + int(paeth)) % 256)
		return data, nil
	default:
		return 0, errors.New("FilterTypeが正しくない")
	}
}

func (self *Apng) parseIHDR(data []uint8) (err error) {
	if len(data) != 13 {
		return errors.New("IHDRのヘッダサイズは13でなければならない")
	}
	self.Ihdr.Width = int(binary.BigEndian.Uint32(data[0:4]))
	self.Ihdr.Height = int(binary.BigEndian.Uint32(data[4:8]))
	self.Ihdr.BitDepth = data[8]
	self.Ihdr.ColorType = data[9]
	self.Ihdr.Compress = data[10]
	self.Ihdr.Filter = data[11]
	self.Ihdr.Interlace = data[12]

	return nil
}
func (self *Apng) parseIDAT(data []uint8) (err error) {
	self.Idat = append(self.Idat, data...)
	return nil
}
func (self *Apng) parseACTL(data []uint8) (err error) {
	if len(self.Idat) > 0 {
		return errors.New("acTL chunkはIDATより前になければいけない")
	}
	if len(data) != 8 {
		return errors.New("acTLのヘッダサイズは8でなければならない")
	}
	self.IsApng = true
	self.Actl.NumFrames = binary.BigEndian.Uint32(data[0:4])
	self.Actl.NumPlays = binary.BigEndian.Uint32(data[4:8])

	return nil
}
func (self *Apng) parseFCTL(data []uint8) (err error) {
	if len(data) != 26 {
		return errors.New("fcTLのヘッダサイズは26でなければならない")
	}
	var fctl Fctl
	fctl.SequenceNumber = binary.BigEndian.Uint32(data[0:4])
	fctl.Width = binary.BigEndian.Uint32(data[4:8])
	fctl.Height = binary.BigEndian.Uint32(data[8:12])
	fctl.OffsetX = binary.BigEndian.Uint32(data[12:16])
	fctl.OffsetY = binary.BigEndian.Uint32(data[16:20])
	fctl.DelayNum = binary.BigEndian.Uint16(data[20:22])
	fctl.DelayDen = binary.BigEndian.Uint16(data[22:24])
	fctl.DisposeOp = data[24]
	fctl.BlendOp = data[25]

	self.Fctl = append(self.Fctl, fctl)

	return nil
}
func (self *Apng) parseFDAT(data []uint8) (err error) {
	seqNumber := binary.BigEndian.Uint32(data[0:4])
	// すでにある場合は追記する
	for _, v := range self.Fdat {
		if v.SequenceNumber == seqNumber {
			v.FrameData = append(v.FrameData, data[4:]...)
			return nil
		}
	}
	// なければ作って付け足そう
	var fdat Fdat
	fdat.SequenceNumber = seqNumber
	fdat.FrameData = data[4:]

	self.Fdat = append(self.Fdat, fdat)

	return nil
}

// Animation PNGとして定義されているすべての画像を生成します
// acTL chunkがない場合は、IDATの画像一枚を返します
func (self *Apng) GenerateAnimation() ([]AnimationData, error) {
	return nil, nil
}
func (idat *Idat) ToImage(width int, height int, colorType ColorType) (image.Image, error) {
	// deflateめんどいしライブラリで許して
	readBuf := bytes.NewBuffer(*idat)
	zr, err := zlib.NewReader(readBuf)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, err
	}
	// byte列に展開
	extracted := buf.Bytes()
	if err != nil {
		return nil, err
	}

	// filter処理をもとに戻す。scanlineごとのfilter-typeで分岐
	// extracted -> dstBuf
	bytePerPixel, err := BytePerPixel(colorType)
	if err != nil {
		return nil, err
	}
	lineBytes := int(bytePerPixel)*width + 1
	dstBufSize := width * height * int(bytePerPixel)
	dstBuf := make([]byte, dstBufSize) // ColorTypeに応じて格納してくれればいい

	for j := 0; j < height; j++ {
		currentLinePtr := j * lineBytes
		prevLinePtr := (j - 1) * lineBytes
		filterType := FilterType(extracted[currentLinePtr])
		// 水平方向のpixel単位でループ
		for i := 0; i < width; i++ {
			// +1はfilterTypeを考慮
			currentPixelPtr := currentLinePtr + 1 + (i * int(bytePerPixel))
			prevPixelPtr := currentLinePtr + 1 + ((i - 1) * int(bytePerPixel))
			prevLinePixelPtr := prevLinePtr + 1 + (i * int(bytePerPixel))
			// pixelごとの色情報ごとにループ
			for c := 0; c < int(bytePerPixel); c++ {
				targetValue := extracted[currentPixelPtr+c]
				topPixelValue := byte(0)
				leftPixelValue := byte(0)
				if i > 0 {
					leftPixelValue = extracted[prevPixelPtr+c]
				}
				if j > 0 {
					topPixelValue = extracted[prevLinePixelPtr+c]
				}
				// すべては出揃った、あとはよしなにやってくれ
				dstPtr := (j*width+i)*int(bytePerPixel) + c
				data, err := cancelFilter(targetValue, filterType, topPixelValue, leftPixelValue)
				if err != nil {
					return nil, err
				}
				dstBuf[dstPtr] = data
			}
		}
	}
	// できたデータをとりあえず画像にするね
	// dstBuf->dst
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			ptr := (j*width + i) * int(bytePerPixel)
			// fmt.Printf("j:%v\ti:%v\tptr:%v\n", j, i, ptr)

			switch colorType {
			case GrayScale:
				dst.SetRGBA(i, j, color.RGBA{dstBuf[ptr], dstBuf[ptr], dstBuf[ptr], uint8(255)})
			case TrueColor:
				dst.SetRGBA(i, j, color.RGBA{dstBuf[ptr], dstBuf[ptr+1], dstBuf[ptr+2], uint8(255)})
			case IndexColor:
				return nil, errors.New("IndexColorにはまだ非対応")
			case GrayScaleWithAlpha:
				dst.SetRGBA(i, j, color.RGBA{dstBuf[ptr], dstBuf[ptr], dstBuf[ptr], dstBuf[ptr+1]})
			case TrueColorWithAlpha:
				dst.SetRGBA(i, j, color.RGBA{dstBuf[ptr], dstBuf[ptr+1], dstBuf[ptr+2], dstBuf[ptr+3]})
			default:
				return nil, errors.New("IndexColorにはまだ非対応")
			}
		}
	}
	return dst, nil
}

func (self *Apng) ToImage() (img image.Image, err error) {
	return self.Idat.ToImage(self.Ihdr.Width, self.Ihdr.Height, ColorType(self.Ihdr.ColorType))
}

func (self *Apng) Parse(src string) (err error) {
	// read file
	f, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	// png header check
	validSignature := []uint8{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	signature := make([]uint8, len(validSignature))
	n, err := f.Read(signature)
	if n == 0 {
		return errors.New("png headerが読み取れなかった")
	}
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(validSignature, signature) {
		return errors.New("pngファイルではない")
	}
	// apngかどうかはacTLがIDATより前に来るかで決まる
	self.IsApng = false
	// read chunks
	isReadIhdr := false
	isReadIdat := false
	isReadIend := false
	for {
		// read header: length(4uint8) + chunk-type
		headersBuf := make([]uint8, 8)
		n, err := f.Read(headersBuf)
		if n < len(headersBuf) {
			// EOF
			break
		}
		if err != nil {
			return errors.New("Chunkヘッダの読み込みエラー")
		}
		length := binary.BigEndian.Uint32(headersBuf[0:4])
		chunkType := string(headersBuf[4:8])
		// read data
		dataBuf := make([]uint8, length)
		n, err = f.Read(dataBuf)
		if err != nil {
			return errors.New("Chunkデータの読み込みエラー")
		}
		// read crc
		crcBuf := make([]uint8, 4)
		n, err = f.Read(crcBuf)
		if err != nil {
			return errors.New("Chunk CRCの読み込みエラー")
		}
		// check crc
		crc := binary.BigEndian.Uint32(crcBuf)
		crcSrc := append(headersBuf[4:8], dataBuf...)
		dataCrc := crc32.ChecksumIEEE(crcSrc)
		if crc != dataCrc {
			fmt.Printf("crc mismatch chunkType:%s crc:%d dataCrc:%d", chunkType, crc, dataCrc)
			continue
		}
		// chunktypeで分岐
		fmt.Printf("chunkType:%s length:%d\n", chunkType, length)
		switch chunkType {
		case "IHDR":
			isReadIhdr = true
			err = self.parseIHDR(dataBuf)
		case "IDAT":
			if !isReadIhdr {
				return errors.New("IHDRが定義される前にIDATが読み出されました")
			}
			isReadIdat = true
			err = self.parseIDAT(dataBuf)
		case "IEND":
			isReadIend = true
			break
		case "acTL":
			err = self.parseACTL(dataBuf)
		case "fcTL":
			err = self.parseFCTL(dataBuf)
		case "fdAT":
			err = self.parseFDAT(dataBuf)
		default:
			fmt.Printf("%sは未実装ヘッダです\n", chunkType)
			err = nil
			continue
		}
		// error check
		if err != nil {
			return err
		}
	}
	// 必須チャンクは来ましたか
	if !isReadIhdr {
		return errors.New("IHDRが記述されていません")
	}
	if !isReadIdat {
		return errors.New("IDATが記述されていません")
	}
	if !isReadIend {
		return errors.New("IENDが記述されていません")
	}
	// 完璧やん
	return nil
}
