package main

import (
	"image"
	"image/color"
)

// Convert to binary image
func Transform(src image.Image) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	ret := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			_, g, _, _ := src.At(x, y).RGBA()
			ret.Set(x, y, color.RGBA{0, uint8(g / 256), 0, 0})
		}
	}
	return ret
}
