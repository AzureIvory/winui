//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AzureIvory/winui/core"
	"github.com/AzureIvory/winui/widgets"
)

type backdropSpec struct {
	Color          string          `json:"color"`
	Opacity        json.RawMessage `json:"opacity"`
	Blur           json.RawMessage `json:"blur"`
	DismissOnClick json.RawMessage `json:"dismissOnClick"`
}

func (b *builder) buildModal(window *Window, spec nodeSpec) (widgets.Widget, error) {
	modal := widgets.NewModal(nodeID(spec))
	layout, layoutKind, err := buildLayout(window, spec.Layout)
	if err != nil {
		return nil, err
	}
	modal.SetLayout(layout)
	if style, err := parsePanelStyle(spec.Style); err != nil {
		return nil, err
	} else {
		modal.SetStyle(style)
	}
	if err := b.applyModalBackdrop(window, modal, spec); err != nil {
		return nil, err
	}
	b.applyCommonState(window, modal, spec)

	regular := make([]widgets.Widget, 0, len(spec.Children))
	modals := make([]widgets.Widget, 0, len(spec.Children))
	for _, child := range spec.Children {
		built, err := b.buildNode(window, child, layoutKind)
		if err != nil {
			return nil, err
		}
		if _, ok := built.(*widgets.Modal); ok {
			modals = append(modals, built)
			continue
		}
		regular = append(regular, built)
	}
	for _, child := range regular {
		modal.Add(child)
	}
	for _, child := range modals {
		modal.Add(child)
	}
	if actionName := strings.TrimSpace(spec.OnDismiss); actionName != "" {
		modal.SetOnDismiss(func() {
			b.dispatchAction(actionName, b.baseActionContext(window, actionName, modal))
		})
	}
	return modal, nil
}

func (b *builder) applyModalBackdrop(window *Window, modal *widgets.Modal, spec nodeSpec) error {
	if modal == nil || len(spec.Backdrop) == 0 {
		return nil
	}
	var backdrop backdropSpec
	if err := json.Unmarshal(spec.Backdrop, &backdrop); err != nil {
		return err
	}
	if backdrop.Color != "" {
		color, ok, err := parseColorValue(backdrop.Color)
		if err != nil {
			return err
		}
		if ok {
			modal.SetBackdropColor(color)
		}
	}
	opacitySource, err := parseIntSource(backdrop.Opacity)
	if err != nil {
		return err
	}
	blurSource, err := parseIntSource(backdrop.Blur)
	if err != nil {
		return err
	}
	dismissSource, err := parseBoolSource(backdrop.DismissOnClick)
	if err != nil {
		return err
	}
	modal.SetBackdropOpacity(byte(clampInt32(resolveIntSourceOrDefault(opacitySource, b.opts.Data, int32(modal.BackdropOpacity())), 0, 255)))
	modal.SetBlurRadiusDP(maxInt32(0, resolveIntSourceOrDefault(blurSource, b.opts.Data, modal.BlurRadiusDP())))
	modal.SetDismissOnBackdrop(resolveBoolSourceOrDefault(dismissSource, b.opts.Data, modal.DismissOnBackdrop()))
	if opacitySource.Binding != "" {
		b.addBinding(window, []string{opacitySource.Binding}, func(ctx *bindingContext) {
			modal.SetBackdropOpacity(byte(clampInt32(resolveIntSourceOrDefault(opacitySource, ctx.data, int32(modal.BackdropOpacity())), 0, 255)))
		})
	}
	if blurSource.Binding != "" {
		b.addBinding(window, []string{blurSource.Binding}, func(ctx *bindingContext) {
			modal.SetBlurRadiusDP(maxInt32(0, resolveIntSourceOrDefault(blurSource, ctx.data, modal.BlurRadiusDP())))
		})
	}
	if dismissSource.Binding != "" {
		b.addBinding(window, []string{dismissSource.Binding}, func(ctx *bindingContext) {
			modal.SetDismissOnBackdrop(resolveBoolSourceOrDefault(dismissSource, ctx.data, modal.DismissOnBackdrop()))
		})
	}
	return nil
}

func setLabelPreferredSize(label *widgets.Label) {
	if label == nil {
		return
	}
	if label.Multiline() {
		widgets.SetPreferredSize(label, core.Size{Width: 180})
		return
	}
	widgets.SetPreferredSize(label, core.Size{Width: 180, Height: 28})
}

func (b *builder) applyLabelOptions(window *Window, label *widgets.Label, spec nodeSpec) error {
	multilineSource, err := parseBoolSource(spec.Multiline)
	if err != nil {
		return err
	}
	wordWrapSource, err := parseBoolSource(spec.WordWrap)
	if err != nil {
		return err
	}
	label.SetMultiline(resolveBoolSourceOrDefault(multilineSource, b.opts.Data, false))
	label.SetWordWrap(resolveBoolSourceOrDefault(wordWrapSource, b.opts.Data, false))
	setLabelPreferredSize(label)
	if multilineSource.Binding != "" {
		b.addBinding(window, []string{multilineSource.Binding}, func(ctx *bindingContext) {
			label.SetMultiline(resolveBoolSourceOrDefault(multilineSource, ctx.data, false))
			setLabelPreferredSize(label)
		})
	}
	if wordWrapSource.Binding != "" {
		b.addBinding(window, []string{wordWrapSource.Binding}, func(ctx *bindingContext) {
			label.SetWordWrap(resolveBoolSourceOrDefault(wordWrapSource, ctx.data, false))
		})
	}
	return nil
}

func (b *builder) configureWindowImage(window *Window, spec windowSpec) error {
	path, sizeDP, err := b.windowImageConfig(spec)
	if err != nil || path == "" {
		return err
	}
	img, err := b.loadImage(path)
	if err != nil {
		return fmt.Errorf("window %q image: %w", spec.ID, err)
	}

	window.Meta.Image = img
	window.Meta.ImagePath = path
	window.Meta.ImageSizeDP = sizeDP

	window.addResourceReloader(func(ctx resourceReloadContext) error {
		switch ctx.Reason {
		case ReloadReasonDPIChanged:
			if ctx.App != nil && window.Meta.Image != nil {
				ctx.App.SetWindowImage(window.Meta.Image)
			}
			return nil
		default:
			reloaded, err := b.loadImage(path)
			if err != nil {
				return fmt.Errorf("window %q image: %w", window.ID, err)
			}
			old := window.Meta.Image
			window.Meta.Image = reloaded
			window.Meta.ImagePath = path
			window.Meta.ImageSizeDP = sizeDP
			if ctx.App != nil {
				ctx.App.SetWindowImage(reloaded)
			}
			if old != nil && old != reloaded {
				_ = old.Close()
			}
			return nil
		}
	})
	return nil
}

func (b *builder) configureButtonImage(window *Window, button *widgets.Button, spec nodeSpec) error {
	path, sizeDP, err := b.nodeImageConfig(spec)
	if err != nil || path == "" {
		return err
	}
	img, err := b.loadImage(path)
	if err != nil {
		return err
	}

	button.SetImage(img)
	if sizeDP > 0 {
		style := button.Style
		style.ImageSizeDP = sizeDP
		button.SetStyle(style)
	}

	window.addResourceReloader(func(ctx resourceReloadContext) error {
		switch ctx.Reason {
		case ReloadReasonDPIChanged:
			if scene := window.Scene(); scene != nil {
				scene.Invalidate(button)
			}
			return nil
		default:
			reloaded, err := b.loadImage(path)
			if err != nil {
				return err
			}
			old := button.Image
			button.SetImage(reloaded)
			if sizeDP > 0 {
				style := button.Style
				style.ImageSizeDP = sizeDP
				button.SetStyle(style)
			}
			if old != nil && old != reloaded {
				_ = old.Close()
			}
			return nil
		}
	})
	return nil
}

func (b *builder) windowImageConfig(spec windowSpec) (string, int32, error) {
	if strings.TrimSpace(spec.Image) == "" {
		return "", 0, nil
	}
	sizeDP, err := b.resolveImageSize(spec.ImageSizeDP)
	if err != nil {
		return "", 0, err
	}
	return spec.Image, sizeDP, nil
}

func (b *builder) nodeImageConfig(spec nodeSpec) (string, int32, error) {
	if strings.TrimSpace(spec.Image) == "" {
		return "", 0, nil
	}
	sizeDP, err := b.resolveImageSize(spec.ImageSizeDP)
	if err != nil {
		return "", 0, err
	}
	return spec.Image, sizeDP, nil
}

func (b *builder) resolveImageSize(raw json.RawMessage) (int32, error) {
	source, err := parseIntSource(raw)
	if err != nil {
		return 0, err
	}
	if source.Has {
		return resolveIntSource(source, b.opts.Data), nil
	}
	return b.opts.ImageSizeDP, nil
}

func (b *builder) loadImage(src string) (*core.Image, error) {
	data, err := os.ReadFile(b.resolveAssetPath(src))
	if err != nil {
		return nil, err
	}
	return core.LoadImageBytes(data)
}

func modalAbsoluteLayoutData(window *Window) absoluteLayoutData {
	return absoluteLayoutData{
		window: window,
		frame: layoutFrame{
			X: literalExprSource(0),
			Y: literalExprSource(0),
			R: literalExprSource(0),
			B: literalExprSource(0),
		},
	}
}

func literalExprSource(value int32) exprSource {
	return exprSource{Has: true, Literal: literalScalarExpr(value)}
}

func clampInt32(value, min, max int32) int32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
