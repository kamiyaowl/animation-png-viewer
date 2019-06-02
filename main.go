package main

import (
	"flag"
	"fmt"

	"./apng"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "png preview",
		Bounds: pixel.R(0, 0, 1024, 768),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	for !win.Closed() {
		win.Update()
	}
}
func main() {
	var (
		src = flag.String("src", "", "png filepath")
	)
	flag.Parse()

	if *src == "" {
		fmt.Println("srcオプションで読み込むファイルを指定してください。 例: -src <filepath>")
		return
	}
	data := apng.Apng{}
	err := data.Parse(*src)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: show png image
	// img, err := data.ToImage()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	pixelgl.Run(run)
}
