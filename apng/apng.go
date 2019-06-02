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

// scanlineごとにある
type FilterType int

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

func (self *Apng) BitPerPixel() (uint8, error) {
	switch self.Ihdr.ColorType {
	case 0:
		// grayscale
		return self.Ihdr.BitDepth, nil
	case 2:
		// true color
		return self.Ihdr.BitDepth * 3, nil
	case 3:
		// index color
		return self.Ihdr.BitDepth, nil
	case 4:
		// grayscale(with alpha)
		return self.Ihdr.BitDepth * 2, nil
	case 6:
		// true color(with alpha)
		return self.Ihdr.BitDepth * 4, nil
	default:
		return 0, errors.New("colorTypeが正しくない")
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
	// IDATは
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
	bitPerPixel, err := self.BitPerPixel()
	if err != nil {
		return nil, err
	}
	lineBytes := int(bitPerPixel)*self.Ihdr.Width + 1
	// filter処理をもとに戻す。scanlineごとのfilter-typeで分岐
	dst := image.NewRGBA(image.Rect(0, 0, self.Ihdr.Width, self.Ihdr.Height))
	// TODO: remove image testcode
	for i := 0; i < self.Ihdr.Width; i++ {
		for j := 0; j < self.Ihdr.Height; j++ {
			dst.Set(i, j, color.RGBA{uint8(i % 255), uint8(j % 255), uint8((i + j) % 255), uint8(255)})
		}
	}
	for j := 0; j < self.Ihdr.Height; j++ {
		basePtr := j * lineBytes // TODO fix
		filterType := extracted[basePtr]
		switch filterType {
		}
		break
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
