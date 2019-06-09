package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"os"

	"./apng"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

// アニメーションを出したい
func showAnimation(data *apng.Apng) {
	results, err := data.GenerateAnimation()
	if err != nil {
		panic(err)
	}
	if results == nil {
		fmt.Println("まだ未実装")
	}
}

// アニメーションじゃないとき
func showImage(data *apng.Apng, outPath string) {
	// decode idat
	img, err := data.ToImage()
	if err != nil {
		panic(err)
	}
	// save image
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		jpeg.Encode(f, img, nil)
	}
	// initialize window
	cfg := pixelgl.WindowConfig{
		Title:  "Preview image",
		Bounds: pixel.R(0, 0, float64(data.Ihdr.Width), float64(data.Ihdr.Height)),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	// image to sprite
	picture := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(picture, picture.Bounds())
	win.Clear(colornames.Skyblue)
	sprite.Draw(win, pixel.IM.Moved(win.Bounds().Center()))

	for !win.Closed() {
		win.Update()
	}
}
func run() {
	// parse args
	var (
		srcPath = flag.String("src", "", "png filepath")
		outPath = flag.String("out", "", "jpeg output, アニメーションのときは無効です")
	)
	flag.Parse()
	if *srcPath == "" {
		fmt.Println("srcオプションで読み込むファイルを指定してください。 例: -src <filepath>")
		return
	}
	// load image
	data := apng.Apng{}
	err := data.Parse(*srcPath)
	if err != nil {
		panic(err)
	}
	fmt.Println(data.Ihdr)

	switch data.IsApng {
	case true:
		showAnimation(&data)
	case false:
		showImage(&data, *outPath)
	}
}
func main() {
	pixelgl.Run(run)
}
