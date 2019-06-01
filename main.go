package main

import (
	"flag"
	"fmt"
)
import "./apng"

func main() {
	var (
		src = flag.String("src", "", "png filepath")
	)
	flag.Parse()

	if *src == "" {
		fmt.Println("srcオプションで読み込むファイルを指定してください。 例: -src <filepath>")
		return
	}
	img := apng.Image{}
	err := img.Parse(*src)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: show png image
	fmt.Println(img)

	fmt.Println("done.")
}
