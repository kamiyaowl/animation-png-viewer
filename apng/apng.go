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
	Ihdr Ihdr
	Idat Idat
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

func (self *Apng) BytePerPixel() (uint8, error) {
	switch ColorType(self.Ihdr.ColorType) {
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
		// TODO: Implement here
		return targetValue, nil
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
func (self *Apng) ToImage() (img image.Image, err error) {
	// deflateめんどいしライブラリで許して
	readBuf := bytes.NewBuffer(self.Idat)
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
	bytePerPixel, err := self.BytePerPixel()
	if err != nil {
		return nil, err
	}
	lineBytes := int(bytePerPixel)*self.Ihdr.Width + 1
	dstBufSize := self.Ihdr.Width * self.Ihdr.Height * int(bytePerPixel)
	dstBuf := make([]byte, dstBufSize) // ColorTypeに応じて格納してくれればいい
	fmt.Printf("len(extracted):%v\tbytePerPixel:%v\tlineBytes:%v\tdstBufSize:%v\t(lineBytes*height):%v\n", len(extracted), bytePerPixel, lineBytes, dstBufSize, lineBytes*self.Ihdr.Height)

	for j := 0; j < self.Ihdr.Height; j++ {
		currentLinePtr := j * lineBytes
		prevLinePtr := (j - 1) * lineBytes
		filterType := FilterType(extracted[currentLinePtr])
		// 水平方向のpixel単位でループ
		for i := 0; i < self.Ihdr.Width; i++ {
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
				dstPtr := (j*self.Ihdr.Width+i)*int(bytePerPixel) + c
				data, err := cancelFilter(targetValue, filterType, topPixelValue, leftPixelValue)
				if err != nil {
					return nil, err
				}
				dstBuf[dstPtr] = data
				// fmt.Printf("j:%v\ti:%v\tc:%v\tdstPtr:%v\t", j, i, c, dstPtr)
				// fmt.Printf("currentPixelPtr:%v\tprevPixelPtr:%v\tprevLinePixelPtr:%v\n", currentPixelPtr, prevPixelPtr, prevLinePixelPtr)
			}
		}
	}
	// できたデータをとりあえず画像にするね
	// dstBuf->dst
	dst := image.NewRGBA(image.Rect(0, 0, self.Ihdr.Width, self.Ihdr.Height))
	for j := 0; j < self.Ihdr.Height; j++ {
		for i := 0; i < self.Ihdr.Width; i++ {
			ptr := (j*self.Ihdr.Width + i) * int(bytePerPixel)
			// fmt.Printf("j:%v\ti:%v\tptr:%v\n", j, i, ptr)

			switch ColorType(self.Ihdr.ColorType) {
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
