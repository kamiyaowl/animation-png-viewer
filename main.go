package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
)

// TODO: PNGをちゃんと型にしてあげる
func parsePng(f *os.File) (png []byte, err error) {
	// png header check
	validSignature := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	signature := make([]byte, len(validSignature))
	n, err := f.Read(signature)
	if n == 0 {
		return nil, errors.New("png header read error")
	}
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(validSignature, signature) {
		return nil, errors.New("This file is not png format.")
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
