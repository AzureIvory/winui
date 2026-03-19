//go:build windows

package widgets

// mergeFontSpec 按字段把字体覆盖值合并到基础字体配置上。
func mergeFontSpec(base, override FontSpec) FontSpec {
	if override.Face != "" {
		base.Face = override.Face
	}
	if override.SizeDP != 0 {
		base.SizeDP = override.SizeDP
	}
	if override.Weight != 0 {
		base.Weight = override.Weight
	}
	return base
}

// mergeTextStyle 按字段把文本样式覆盖值合并到基础样式上。
func mergeTextStyle(base, override TextStyle) TextStyle {
	base.Font = mergeFontSpec(base.Font, override.Font)
	if override.Color != 0 {
		base.Color = override.Color
	}
	if override.Format != 0 {
		base.Format = override.Format
	}
	return base
}
