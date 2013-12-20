package main

import (
	"code.google.com/p/go-tour/pic"
	"image"
	"image/color"
	//"math"
)

type Image struct{}

func (img Image) ColorModel() color.Model {
	return color.RGBAModel
}

func (img Image) Bounds() image.Rectangle {
	return image.Rect(0, 0, 100, 100)
}

func (img Image) At(x, y int) color.Color {
	//v := uint8(math.Pow(float64(x), float64(y)))
	v := uint8(x) * uint8(y)
	return color.RGBA{v, v, 255, 255}
}

func main() {
	m := Image{}
	pic.ShowImage(m)
}
