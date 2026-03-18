# 开发 winui

## 目标

这个模块的目标是为 Go 项目提供可复用的 Windows 原生 UI 工具包。它应始终保持：

- 仅支持 Windows
- 自包含
- 不包含应用特定的业务逻辑
- 易于嵌入小型工具

## 仓库结构

- `core/`：底层 Win32 封装与绘制原语
- `widgets/`：构建在 `core` 之上的控件抽象
- `cmd/demo/`：自包含的手工测试应用
- `scripts/`：仓库维护辅助脚本

## 前置条件

- Windows
- Go 1.22 或更高版本

## 常用命令

在模块根目录执行：

```powershell
go test ./...
go run ./cmd/demo
go vet ./...
```

## 模块路径变更

如果仓库路径变化，更新模块路径和内部导入：

```powershell
.\scripts\Set-ModulePath.ps1 -ModulePath github.com/<your-org>/<your-repo>
```

## 设计准则

- 保持 `core` 只关注 Win32 原语，不承担控件语义。
- 保持 `widgets` 不依赖任何应用逻辑。
- 所有平台相关文件都应带构建标签。
- 优先清晰的导出 API，而不是炫技式抽象。
- 以 `Scene` 作为控件事件与渲染的主要协调者。

## 文档准则

- 在 `doc.go` 中添加包注释。
- 为导出符号补全文档。
- 函数、自定义类型、结构体字段与常量统一使用中文注释，便于直接从源码阅读。
- 当辅助函数的行为无法从名称直接看出时，也应补注释。
- 公共 API 有明显变化时同步更新 README。


