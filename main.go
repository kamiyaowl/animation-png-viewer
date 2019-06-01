package main

import (
	"flag"
	"fmt"
	"os"
)
import "./apng"

func main() {
	var (
		src = flag.String("src", "", "png filepath")
	)
	flag.Parse()

	f, err := os.Open(*src)
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
