//go:build windows

package main

import "github.com/AzureIvory/winui/core"

// layout 按当前窗口尺寸摆放演示控件。
func layout(w, h int32) {
	if ui.app == nil {
		return
	}

	margin := ui.app.DP(24)
	columnGap := ui.app.DP(28)
	leftW := (w - margin*2 - columnGap) / 2
	rightX := margin + leftW + columnGap
	fieldH := ui.app.DP(42)
	rowGap := ui.app.DP(16)

	ui.title.SetBounds(core.Rect{X: margin, Y: ui.app.DP(20), W: w - margin*2, H: ui.app.DP(36)})
	ui.renderer.SetBounds(core.Rect{X: margin, Y: ui.app.DP(56), W: w - margin*2, H: ui.app.DP(22)})
	ui.status.SetBounds(core.Rect{X: margin, Y: ui.app.DP(84), W: w - margin*2, H: ui.app.DP(24)})

	ui.progressPct.SetBounds(core.Rect{X: margin, Y: ui.app.DP(118), W: leftW, H: ui.app.DP(20)})
	ui.progress.SetBounds(core.Rect{X: margin, Y: ui.app.DP(144), W: leftW, H: ui.app.DP(16)})
	ui.check.SetBounds(core.Rect{X: margin, Y: ui.app.DP(178), W: leftW, H: ui.app.DP(30)})
	ui.radioA.SetBounds(core.Rect{X: margin, Y: ui.app.DP(216), W: leftW, H: ui.app.DP(30)})
	ui.radioB.SetBounds(core.Rect{X: margin, Y: ui.app.DP(254), W: leftW, H: ui.app.DP(30)})
	ui.edit.SetBounds(core.Rect{X: margin, Y: ui.app.DP(302), W: leftW, H: fieldH})

	ui.list.SetBounds(core.Rect{X: rightX, Y: ui.app.DP(122), W: leftW, H: ui.app.DP(176)})
	ui.combo.SetBounds(core.Rect{X: rightX, Y: ui.app.DP(314), W: leftW, H: fieldH})

	buttonY := h - margin - ui.app.DP(48)
	buttonW := ui.app.DP(132)
	ui.btnStep.SetBounds(core.Rect{X: margin, Y: buttonY, W: buttonW, H: ui.app.DP(40)})
	ui.btnReset.SetBounds(core.Rect{X: margin + buttonW + rowGap, Y: buttonY, W: buttonW, H: ui.app.DP(40)})
}
