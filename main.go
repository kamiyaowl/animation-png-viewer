package main

import (
	"fmt"
	"os"
)
import "./apng"

func main() {
	// TODO: set path from cmd argument
	path := "sample_data/PNG_transparency_demonstration_1.png"

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	img := apng.Image{}
	err = img.Parse(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: show png image
	fmt.Println(img)

	fmt.Println("done.")
}
