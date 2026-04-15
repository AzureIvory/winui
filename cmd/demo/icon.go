//go:build windows

package main

import (
	"image"
	"image/color"

	"github.com/AzureIvory/winui/core"
)

// demoBadge 生成一个简单的演示图片，用于左图右字按钮样式。
func demoBadge(fill, accent color.RGBA) *core.Image {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	fillCircle(img, 16, 16, 14, fill)
	fillCircle(img, 11, 11, 5, accent)
	fillCircle(img, 20, 21, 4, accent)

	imageResource, err := core.NewImageFromDecoded(img)
	if err != nil {
		return nil
	}
	return imageResource
}

// fillCircle 在图像上填充一个纯色圆形。
func fillCircle(img *image.RGBA, cx, cy, radius int, clr color.RGBA) {
	if img == nil || radius <= 0 {
		return
	}
	r2 := radius * radius
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if !image.Pt(x, y).In(img.Bounds()) {
				continue
			}
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= r2 {
				img.SetRGBA(x, y, clr)
			}
		}
	}
}
