//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// ColorChannels 把颜色拆分红、绿、蓝三个 8 位通道。
// winui 的 Color 以 0x00BBGGRR 形式存储，因此红色位于低位。
func ColorChannels(color core.Color) (r, g, b byte) {
	return byte(color), byte(color >> 8), byte(color >> 16)
}

// BlendChannel 把前景通道 fg 按强度 alpha 混合到背景通道 bg 上。
// alpha 为 0 时完全取背景，为 255 时完全取前景。
func BlendChannel(bg, fg, alpha byte) byte {
	const scale = 255
	value := int(bg)*(scale-int(alpha)) + int(fg)*int(alpha)
	return byte((value + scale/2) / scale)
}

// BlendColor 把前景色 fg 按强度 alpha 叠加到背景色 bg 上，返回不透明结果。
// 由于 Color 不携带 alpha 通道，半透明叠加需要调用方显式给出混合强度。
func BlendColor(bg, fg core.Color, alpha byte) core.Color {
	bgR, bgG, bgB := ColorChannels(bg)
	fgR, fgG, fgB := ColorChannels(fg)
	return core.RGB(
		BlendChannel(bgR, fgR, alpha),
		BlendChannel(bgG, fgG, alpha),
		BlendChannel(bgB, fgB, alpha),
	)
}

// IsColorNearWhite 判断颜色是否接近白色，常用于自动挑选对比文字色。
func IsColorNearWhite(color core.Color) bool {
	r, g, b := ColorChannels(color)
	return r >= 240 && g >= 240 && b >= 240
}
