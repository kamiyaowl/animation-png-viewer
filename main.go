package main

import (
	//"hash/crc32"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"reflect"
)

type Image struct {
	width  uint32
	height uint32
}

// TODO: PNGをちゃんと型にしてあげる
func (self *Image) parsePng(f *os.File) (err error) {
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
		// TODO: CRCの実装
		// crc := binary.BigEndian.Uint32(crcBuf)
		// TODO: check crc
		// dataCrc := crc32.ChecksumIEEE(dataBuf) // ChunkTypeからやるべき
		// if crc != dataCrc {
		// 	fmt.Printf("crc mismatch chunkType:%s crc:%d dataCrc:%d", chunkType, crc, dataCrc)
		// 	continue
		// }
		fmt.Printf("chunkType:%s length:%d\n", chunkType, length)
		// chunk typeで分岐
		switch chunkType {
		case "IHDR":
		case "IEND":
		default:
			fmt.Printf("%sは未実装ヘッダです\n", chunkType)
			continue
		}
		// Test
		break
	}
	return nil
}
func main() {
	// TODO: set path from cmd argument
	path := "sample_data/PNG_transparency_demonstration_1.png"
	fmt.Println("image file opening...")
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	// read and parse png
	fmt.Println("png parsing...")
	img := Image{}
	err = img.parsePng(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: show png image
	fmt.Println(img)

	fmt.Println("done.")
}
