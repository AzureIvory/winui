//go:build windows

package main

import (
	"encoding/binary"
	"image"
	"image/color"

	"github.com/AzureIvory/winui/core"
)

// demoBadge 生成一个简单的演示图标，用于左图标按钮样式。
func demoBadge(fill, accent color.RGBA) *core.Icon {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	fillCircle(img, 16, 16, 14, fill)
	fillCircle(img, 11, 11, 5, accent)
	fillCircle(img, 20, 21, 4, accent)

	icon, err := core.LoadIconFromICO(buildICO(img), 32)
	if err != nil {
		return nil
	}
	return icon
}

// buildICO 把 RGBA 图像打包成单图标 ICO 数据。
func buildICO(img *image.RGBA) []byte {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	maskStride := ((w + 31) / 32) * 4
	maskSize := maskStride * h
	bmpSize := 40 + w*h*4 + maskSize

	data := make([]byte, 6+16+bmpSize)
	binary.LittleEndian.PutUint16(data[2:], 1)
	binary.LittleEndian.PutUint16(data[4:], 1)

	entry := data[6:22]
	entry[0] = byte(w)
	entry[1] = byte(h)
	binary.LittleEndian.PutUint16(entry[4:], 1)
	binary.LittleEndian.PutUint16(entry[6:], 32)
	binary.LittleEndian.PutUint32(entry[8:], uint32(bmpSize))
	binary.LittleEndian.PutUint32(entry[12:], 22)

	bmp := data[22:]
	binary.LittleEndian.PutUint32(bmp[0:], 40)
	binary.LittleEndian.PutUint32(bmp[4:], uint32(w))
	binary.LittleEndian.PutUint32(bmp[8:], uint32(h*2))
	binary.LittleEndian.PutUint16(bmp[12:], 1)
	binary.LittleEndian.PutUint16(bmp[14:], 32)
	binary.LittleEndian.PutUint32(bmp[20:], uint32(w*h*4))

	pixelOffset := 40
	index := 0
	for y := h - 1; y >= 0; y-- {
		row := img.Pix[y*img.Stride:]
		for x := 0; x < w; x++ {
			src := x * 4
			dst := pixelOffset + index*4
			data[22+dst] = row[src+2]
			data[22+dst+1] = row[src+1]
			data[22+dst+2] = row[src]
			data[22+dst+3] = row[src+3]
			index++
		}
	}

	return data
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
