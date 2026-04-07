//go:build windows

package widgets

import "github.com/AzureIvory/winui/core"

// FontSpec 描述控件文本使用的字体规格。
type FontSpec struct {
	// Face 指定字体名称。
	Face string
	// SizeDP 指定按 DPI 缩放前的字号。
	SizeDP int32
	// Weight 指定字体粗细，400 为常规，700 为粗体。
	Weight int32
}

// TextStyle 描述通用文本绘制样式。
type TextStyle struct {
	// Font 指定文本使用的字体规格。
	Font FontSpec
	// Color 指定文本颜色。
	Color core.Color
	// Format 指定 DrawText 使用的排版标志。
	Format uint32
}

// ButtonStyle 描述按钮控件的外观样式。
type ButtonStyle struct {
	// Font 指定按钮文本使用的字体规格。
	Font FontSpec
	// TextAlign specifies how button text is aligned within its content area.
	TextAlign Alignment
	// TextColor 指定默认状态下的文字颜色。
	TextColor core.Color
	// DownText 指定按下状态下的文字颜色。
	DownText core.Color
	// DisabledText 指定禁用状态下的文字颜色。
	DisabledText core.Color
	// Background 指定默认背景色。
	Background core.Color
	// Hover 指定悬停状态背景色。
	Hover core.Color
	// Pressed 指定按下状态背景色。
	Pressed core.Color
	// Disabled 指定禁用状态背景色。
	Disabled core.Color
	// Border 指定边框颜色。
	Border core.Color
	// CornerRadius 指定圆角半径。
	CornerRadius int32
	// IconSizeDP 指定图标尺寸。
	IconSizeDP int32
	// TextInsetDP 指定文本区域参考高度。
	TextInsetDP int32
	// GapDP 指定图标和文字之间的间距。
	GapDP int32
	// PadDP 指定按钮内容内边距。
	PadDP int32
}

// ProgressStyle 描述进度条控件的外观样式。
type ProgressStyle struct {
	// Font 指定百分比文本使用的字体规格。
	Font FontSpec
	// TextColor 指定百分比文本颜色。
	TextColor core.Color
	// TrackColor 指定轨道颜色。
	TrackColor core.Color
	// FillColor 指定进度填充颜色。
	FillColor core.Color
	// BubbleColor 指定气泡颜色。
	BubbleColor core.Color
	// CornerRadius 指定轨道圆角半径。
	CornerRadius int32
	// ShowPercent 控制是否显示百分比气泡。
	ShowPercent bool
}

// ChoiceIndicatorStyle 表示单选框或多选框选中标记的绘制样式。
type ChoiceIndicatorStyle uint8

const (
	// ChoiceIndicatorAuto 表示沿用主题或控件默认的选中标记样式。
	ChoiceIndicatorAuto ChoiceIndicatorStyle = iota
	// ChoiceIndicatorDot 表示使用圆点样式绘制选中标记。
	ChoiceIndicatorDot
	// ChoiceIndicatorCheck 表示使用打钩样式绘制选中标记。
	ChoiceIndicatorCheck
)

// ChoiceStyle 描述复选框和单选框的外观样式。
type ChoiceStyle struct {
	// Font 指定标签文本使用的字体规格。
	Font FontSpec
	// TextColor 指定默认文本颜色。
	TextColor core.Color
	// DisabledText 指定禁用文本颜色。
	DisabledText core.Color
	// Background 指定指示器背景色。
	Background core.Color
	// BorderColor 指定默认边框颜色。
	BorderColor core.Color
	// HoverBorder 指定悬停边框颜色。
	HoverBorder core.Color
	// FocusBorder 指定焦点边框颜色。
	FocusBorder core.Color
	// IndicatorColor 指定选中指示器颜色。
	IndicatorColor core.Color
	// CheckColor 指定打钩图形颜色。
	CheckColor core.Color
	// IndicatorStyle 指定选中标记的绘制样式。
	IndicatorStyle ChoiceIndicatorStyle
	// HoverBackground 指定悬停时包裹区域背景色。
	HoverBackground core.Color
	// DisabledBg 指定禁用状态背景色。
	DisabledBg core.Color
	// DisabledBorder 指定禁用状态边框颜色。
	DisabledBorder core.Color
	// CornerRadius 指定指示器圆角半径。
	CornerRadius int32
	// IndicatorSizeDP 指定指示器尺寸。
	IndicatorSizeDP int32
	// IndicatorGapDP 指定指示器和文本之间的间距。
	IndicatorGapDP int32
}

// ListStyle 描述列表框控件的外观样式。
type ListStyle struct {
	// Font 指定列表文本使用的字体规格。
	Font FontSpec
	// TextColor 指定默认文本颜色。
	TextColor core.Color
	// DisabledText 指定禁用项文本颜色。
	DisabledText core.Color
	// Background 指定列表背景色。
	Background core.Color
	// BorderColor 指定默认边框颜色。
	BorderColor core.Color
	// HoverBorder 指定悬停边框颜色。
	HoverBorder core.Color
	// FocusBorder 指定焦点边框颜色。
	FocusBorder core.Color
	// ItemHoverColor 指定悬停行背景色。
	ItemHoverColor core.Color
	// ItemSelectedColor 指定选中行背景色。
	ItemSelectedColor core.Color
	// ItemTextColor 指定选中行文本颜色。
	ItemTextColor core.Color
	// ItemHeightDP 指定每行高度。
	ItemHeightDP int32
	// PaddingDP 指定列表内边距。
	PaddingDP int32
	// CornerRadius 指定列表圆角半径。
	CornerRadius int32
}

// ComboStyle 描述组合框控件的外观样式。
type ComboStyle struct {
	// Font 指定组合框文本使用的字体规格。
	Font FontSpec
	// TextColor 指定已选文本颜色。
	TextColor core.Color
	// PlaceholderColor 指定占位文本颜色。
	PlaceholderColor core.Color
	// Background 指定输入框背景色。
	Background core.Color
	// BorderColor 指定默认边框颜色。
	BorderColor core.Color
	// HoverBorder 指定悬停边框颜色。
	HoverBorder core.Color
	// FocusBorder 指定焦点边框颜色。
	FocusBorder core.Color
	// ArrowColor 指定箭头颜色。
	ArrowColor core.Color
	// PopupBackground 指定弹出层背景色。
	PopupBackground core.Color
	// ItemHoverColor 指定弹出项悬停背景色。
	ItemHoverColor core.Color
	// ItemSelectedColor 指定弹出项选中背景色。
	ItemSelectedColor core.Color
	// ItemTextColor 指定弹出项选中文本颜色。
	ItemTextColor core.Color
	// ItemHeightDP 指定弹出项高度。
	ItemHeightDP int32
	// PaddingDP 指定控件和弹出层内边距。
	PaddingDP int32
	// CornerRadius 指定圆角半径。
	CornerRadius int32
	// MaxVisibleItems 指定弹出层最多可见条目数。
	MaxVisibleItems int32
}

// EditStyle 描述编辑框控件的外观样式。
type EditStyle struct {
	// Font 指定编辑框文本使用的字体规格。
	Font FontSpec
	// TextAlign specifies how edit text is aligned within the content area.
	TextAlign Alignment
	// TextColor 指定默认文本颜色。
	TextColor core.Color
	// PlaceholderColor 指定占位文本颜色。
	PlaceholderColor core.Color
	// Background 指定背景色。
	Background core.Color
	// BorderColor 指定默认边框颜色。
	BorderColor core.Color
	// HoverBorder 指定悬停边框颜色。
	HoverBorder core.Color
	// FocusBorder 指定焦点边框颜色。
	FocusBorder core.Color
	// DisabledText 指定禁用文本颜色。
	DisabledText core.Color
	// DisabledBg 指定禁用背景色。
	DisabledBg core.Color
	// CaretColor 指定光标颜色。
	CaretColor core.Color
	// PaddingDP 指定编辑区内边距。
	PaddingDP int32
	// CornerRadius 指定圆角半径。
	CornerRadius int32
}

// Theme 聚合场景中各类控件的默认样式。
type Theme struct {
	// BackgroundColor 指定场景默认背景色。
	BackgroundColor core.Color
	// Text 指定常规文本样式。
	Text TextStyle
	// Title 指定标题文本样式。
	Title TextStyle
	// Button 指定按钮默认样式。
	Button ButtonStyle
	// Progress 指定进度条默认样式。
	Progress ProgressStyle
	// CheckBox 指定复选框默认样式。
	CheckBox ChoiceStyle
	// RadioButton 指定单选按钮默认样式。
	RadioButton ChoiceStyle
	// ListBox 指定列表框默认样式。
	ListBox ListStyle
	// ComboBox 指定组合框默认样式。
	ComboBox ComboStyle
	// Edit 指定编辑框默认样式。
	Edit EditStyle
}

// DefaultTheme 返回控件未覆写时使用的默认主题。
func DefaultTheme() *Theme {
	return &Theme{
		BackgroundColor: core.RGB(255, 255, 255),
		Text: TextStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 16,
			},
			Color:  core.RGB(16, 16, 16),
			Format: core.DTCenter | core.DTVCenter | core.DTSingleLine,
		},
		Title: TextStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 20,
			},
			Color:  core.RGB(16, 16, 16),
			Format: core.DTCenter | core.DTVCenter | core.DTSingleLine,
		},
		Button: ButtonStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 16,
			},
			TextAlign:    AlignCenter,
			TextColor:    core.RGB(15, 23, 42),
			DownText:     core.RGB(255, 255, 255),
			DisabledText: core.RGB(148, 163, 184),
			Background:   core.RGB(255, 255, 255),
			Hover:        core.RGB(239, 246, 255),
			Pressed:      core.RGB(37, 99, 235),
			Disabled:     core.RGB(241, 245, 249),
			Border:       core.RGB(191, 219, 254),
			CornerRadius: 10,
			TextInsetDP:  18,
			GapDP:        8,
			PadDP:        12,
		},
		Progress: ProgressStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 14,
				Weight: 700,
			},
			TextColor:    core.RGB(255, 255, 255),
			TrackColor:   core.RGB(243, 244, 246),
			FillColor:    core.RGB(16, 185, 129),
			BubbleColor:  core.RGB(5, 150, 105),
			CornerRadius: 12,
			ShowPercent:  true,
		},
		CheckBox: ChoiceStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 15,
			},
			TextColor:       core.RGB(31, 41, 55),
			DisabledText:    core.RGB(156, 163, 175),
			Background:      core.RGB(255, 255, 255),
			BorderColor:     core.RGB(203, 213, 225),
			HoverBorder:     core.RGB(56, 189, 248),
			FocusBorder:     core.RGB(14, 165, 233),
			IndicatorColor:  core.RGB(14, 165, 233),
			CheckColor:      core.RGB(255, 255, 255),
			IndicatorStyle:  ChoiceIndicatorDot,
			HoverBackground: core.RGB(240, 249, 255),
			DisabledBg:      core.RGB(243, 244, 246),
			DisabledBorder:  core.RGB(209, 213, 219),
			CornerRadius:    6,
			IndicatorSizeDP: 18,
			IndicatorGapDP:  10,
		},
		RadioButton: ChoiceStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 15,
			},
			TextColor:       core.RGB(31, 41, 55),
			DisabledText:    core.RGB(156, 163, 175),
			Background:      core.RGB(255, 255, 255),
			BorderColor:     core.RGB(203, 213, 225),
			HoverBorder:     core.RGB(56, 189, 248),
			FocusBorder:     core.RGB(14, 165, 233),
			IndicatorColor:  core.RGB(14, 165, 233),
			CheckColor:      core.RGB(255, 255, 255),
			IndicatorStyle:  ChoiceIndicatorDot,
			HoverBackground: core.RGB(240, 249, 255),
			DisabledBg:      core.RGB(243, 244, 246),
			DisabledBorder:  core.RGB(209, 213, 219),
			CornerRadius:    9,
			IndicatorSizeDP: 18,
			IndicatorGapDP:  10,
		},
		ListBox: ListStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 15,
			},
			TextColor:         core.RGB(31, 41, 55),
			DisabledText:      core.RGB(156, 163, 175),
			Background:        core.RGB(255, 255, 255),
			BorderColor:       core.RGB(203, 213, 225),
			HoverBorder:       core.RGB(96, 165, 250),
			FocusBorder:       core.RGB(37, 99, 235),
			ItemHoverColor:    core.RGB(239, 246, 255),
			ItemSelectedColor: core.RGB(37, 99, 235),
			ItemTextColor:     core.RGB(255, 255, 255),
			ItemHeightDP:      34,
			PaddingDP:         8,
			CornerRadius:      10,
		},
		ComboBox: ComboStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 15,
			},
			TextColor:         core.RGB(31, 41, 55),
			PlaceholderColor:  core.RGB(156, 163, 175),
			Background:        core.RGB(255, 255, 255),
			BorderColor:       core.RGB(203, 213, 225),
			HoverBorder:       core.RGB(96, 165, 250),
			FocusBorder:       core.RGB(37, 99, 235),
			ArrowColor:        core.RGB(37, 99, 235),
			PopupBackground:   core.RGB(255, 255, 255),
			ItemHoverColor:    core.RGB(239, 246, 255),
			ItemSelectedColor: core.RGB(37, 99, 235),
			ItemTextColor:     core.RGB(255, 255, 255),
			ItemHeightDP:      34,
			PaddingDP:         10,
			CornerRadius:      10,
			MaxVisibleItems:   6,
		},
		Edit: EditStyle{
			Font: FontSpec{
				Face:   "Microsoft YaHei UI",
				SizeDP: 15,
			},
			TextAlign:        AlignStart,
			TextColor:        core.RGB(31, 41, 55),
			PlaceholderColor: core.RGB(156, 163, 175),
			Background:       core.RGB(255, 255, 255),
			BorderColor:      core.RGB(203, 213, 225),
			HoverBorder:      core.RGB(96, 165, 250),
			FocusBorder:      core.RGB(37, 99, 235),
			DisabledText:     core.RGB(156, 163, 175),
			DisabledBg:       core.RGB(243, 244, 246),
			CaretColor:       core.RGB(37, 99, 235),
			PaddingDP:        10,
			CornerRadius:     10,
		},
	}
}
