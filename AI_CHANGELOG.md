# AI Changelog

## 2026-04-19

### Composable DPI policy system

- Added widget-level DPI policies in `widgets.ScalePolicy` with `Mode`, `Layout`, `Font`, `Image`, `Padding`, `Gap`, and `Radius`
- Added `widgets.SetScalePolicy(...)`, `widgets.ScalePolicyOf(...)`, `widgets.LayoutScaleModeOf(...)`, and `widgets.ScaleLayoutValue(...)`
- Preserved backward compatibility: widgets and JSON nodes without an explicit policy still use the legacy all-DP behavior
- Layout sizing now distinguishes between:
  - a widget's own layout metrics such as preferred size and absolute `frame`
  - a container's own layout metrics such as row/col/grid/form `gap`, `padding`, `itemSize`, and `labelWidth`
- JSON UI now accepts a node-level `scale` field
- JSON absolute-layout `frame` expressions remain logical DP by default, but switch to physical-pixel evaluation when `scale.layout = "px"`
- `Label` and `Button` now resolve font / image / padding / gap / radius through the new policy slots
- Added DPI regression coverage for `100%`, `125%`, `150%`, and `175%` across widgets and JSON UI

## 2026-04-20

### Staticcheck in CI

- Added `scripts/staticcheck.ps1` to run `staticcheck` with `GOOS=windows` and an explicit `CGO_ENABLED=0/1` variant
- Added `staticcheck` to `.github/workflows/ci.yml` so both `cgo=0` and `cgo=1` CI jobs run the same lint path
- Cleaned the current repo `staticcheck` findings so the library passes under both build variants
