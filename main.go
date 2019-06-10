package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"time"

	"./apng"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

// imageをjpegで保存します
func saveJpg(img image.Image, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	jpeg.Encode(f, img, nil)

	return nil
}
func imageToSprite(img *image.Image) *pixel.Sprite {
	picture := pixel.PictureDataFromImage(*img)
	sprite := pixel.NewSprite(picture, picture.Bounds())
	return sprite
}

// アニメーションを出したい
func showAnimation(data *apng.Apng) {
	frames, err := data.GenerateAnimation()
	if err != nil {
		panic(err)
	}
	cfg := pixelgl.WindowConfig{
		Title:  "Preview image",
		Bounds: pixel.R(0, 0, float64(data.Ihdr.Width), float64(data.Ihdr.Height)),
		VSync:  true,
	}
	// image to sprite
	sprites := []pixel.Sprite{}
	for _, f := range frames {
		sp := imageToSprite(&f.Image)
		sprites = append(sprites, *sp)
	}

	// 適当に順番に表示する
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
	for currentIndex := 0; !win.Closed(); currentIndex = (currentIndex + 1) % len(sprites) {
		win.Clear(colornames.Skyblue)
		sprites[currentIndex].Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
		duration := time.Millisecond * time.Duration(frames[currentIndex].DelaySeconds)
		time.Sleep(duration)
	}
}

// アニメーションじゃないとき
func showImage(data *apng.Apng) {
	// decode idat
	img, err := data.ToImage()
	if err != nil {
		panic(err)
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
		showImage(&data)
	}
}
func main() {
	pixelgl.Run(run)
}
