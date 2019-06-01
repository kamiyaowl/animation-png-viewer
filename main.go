package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"reflect"
)

// TODO: PNGをちゃんと型にしてあげる
func parsePng(f *os.File) (png []uint8, err error) {
	// png header check
	validSignature := []uint8{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	signature := make([]uint8, len(validSignature))
	n, err := f.Read(signature)
	if n == 0 {
		return nil, errors.New("png headerが読み取れなかった")
	}
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(validSignature, signature) {
		return nil, errors.New("pngファイルではない")
	}
	// read chunks
	for {
		// length(4uint8) + chunk-type
		headersBuf := make([]uint8, 8)
		n, err := f.Read(headersBuf)
		if n < len(headersBuf) {
			// EOF
			break
		}
		if err != nil {
			return nil, errors.New("Chunkヘッダの読み込みエラー")
		}
		length := binary.BigEndian.Uint32(headersBuf[0:4])
		chunkType := string(headersBuf[4:4])
		fmt.Println(length)
		fmt.Println(chunkType)
		// Test
		break
	}
	return nil, nil
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
	png, err := parsePng(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: show png image
	fmt.Println(png)

	fmt.Println("done.")
}
