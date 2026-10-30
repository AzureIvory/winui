//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// ActionVisualState is presentation only. The host owns operation state,
// availability, transitions and any lifetime of success/error feedback.
type ActionVisualState uint8

const (
	// ActionVisualDefault is the ordinary available-action appearance.
	ActionVisualDefault ActionVisualState = iota
	// ActionVisualWorking indicates an operation in progress without disabling it.
	ActionVisualWorking
	// ActionVisualSuccess indicates host-reported completion.
	ActionVisualSuccess
	// ActionVisualError indicates a host-reported failure.
	ActionVisualError
)

// ItemActionPresentation describes one row's action without storing business state.
type ItemActionPresentation struct {
	Text     string
	State    ActionVisualState
	Disabled bool
}

// ActionButtonStyle returns a shared, high-contrast action appearance for custom buttons
// and ListBox item actions. It never changes availability or schedules transitions.
func ActionButtonStyle(state ActionVisualState) ButtonStyle {
	background := core.RGB(29, 78, 216)
	switch state {
	case ActionVisualWorking:
		background = core.RGB(3, 105, 161)
	case ActionVisualSuccess:
		background = core.RGB(21, 128, 61)
	case ActionVisualError:
		background = core.RGB(185, 28, 28)
	default:
		state = ActionVisualDefault
	}
	white := core.RGB(255, 255, 255)
	style := ButtonStyle{
		Font:      FontSpec{SizeDP: 12, Weight: 600},
		TextColor: white, DownText: white, DisabledText: white,
		Background: background, Hover: BlendColor(background, white, 12),
		Pressed:  BlendColor(background, core.RGB(0, 0, 0), 24),
		Disabled: background, Border: background, Shape: ButtonShapeDefault,
		CornerRadius: 5, ImageSizeDP: 16, GapDP: 4, PadDP: 4,
	}
	if state == ActionVisualDefault {
		style.Disabled = core.RGB(226, 232, 240)
		style.DisabledText = core.RGB(100, 116, 139)
	}
	return style
}
