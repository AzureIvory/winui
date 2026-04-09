# Markup 绑定指南

这份文档面向直接使用 `widgets/markup` 的业务代码，重点说明如何在 markup 里声明绑定，并在修改一份数据后自动刷新 UI。

## 适用场景

适合下面这类需求：

- 改窗口标题
- 改标签、按钮、单选框、多选框文案
- 改输入框内容
- 改进度条数值
- 改组件可见性、禁用状态
- 改绝对布局的位置和尺寸
- 改 `select` / `listbox` 的数据和选中项

如果只是一次性构建静态界面，不需要 `State`。

## 1. 最小用法

Go:

```go
state := markup.NewState(map[string]any{
	"page": map[string]any{
		"title":   "搜索",
		"visible": true,
	},
	"form": map[string]any{
		"query": "初始关键字",
	},
})

doc, err := markup.LoadIntoScene(scene, htmlText, "", markup.LoadOptions{
	State: state,
})
if err != nil {
	panic(err)
}

state.Set("page.title", "搜索结果")
state.Set("form.query", "下一次查询")
_ = doc
```

Markup:

```html
<window bind-title="page.title">
  <body>
    <label bind-text="page.title" bind-visible="page.visible"></label>
    <input bind-value="form.query" />
  </body>
</window>
```

效果：

- `page.title` 改了，窗口标题和标签文字都会刷新
- `page.visible` 改了，标签显隐会刷新
- `form.query` 改了，输入框内容会刷新

## 2. State 的更新方式

`markup.State` 提供三种常用更新方式：

- `state.Set(path, value)`: 更新一个路径，例如 `state.Set("page.title", "新标题")`
- `state.Patch(map[string]any{...})`: 一次更新多个路径
- `state.Replace(snapshot)`: 用一整份新快照替换旧数据

建议：

- 你的数据本来就是 `map[string]any` 时，优先用 `Set` / `Patch`
- 你的数据更适合用 struct 重新组装时，优先用 `Replace`

注意：

- `Set` 设计目标是 map 风格的根数据，最稳妥的根类型是 `map[string]any`
- 路径分隔符是 `.`

## 3. 支持的绑定属性

### 3.1 窗口

- `bind-title`

示例：

```html
<window bind-title="page.title">
  <body></body>
</window>
```

### 3.2 文本类

- `bind-text`

适用：

- `label`
- `button`
- `checkbox`
- `radio`

示例：

```html
<label bind-text="profile.name"></label>
<button bind-text="actions.submitText"></button>
```

### 3.3 值类

- `bind-value`

适用：

- `input`
- `textarea`
- `progress`

示例：

```html
<input bind-value="form.keyword" />
<textarea bind-value="editor.content"></textarea>
<progress bind-value="task.percent"></progress>
```

### 3.4 状态类

- `bind-visible`
- `bind-enabled`
- `bind-checked`

示例：

```html
<label bind-visible="panel.showHint"></label>
<button bind-enabled="page.canSubmit"></button>
<checkbox bind-checked="filters.onlyMine"></checkbox>
```

### 3.5 尺寸与绝对布局

通用尺寸：

- `bind-width`
- `bind-height`

绝对布局：

- `bind-left`
- `bind-top`
- `bind-right`
- `bind-bottom`
- `bind-x`
- `bind-y`

示例：

```html
<body style="display:absolute">
  <label
    bind-text="card.title"
    bind-left="card.x"
    bind-top="card.y"
    bind-width="card.width"
    bind-height="card.height"></label>
</body>
```

说明：

- `bind-x` 等价于 `bind-left`
- `bind-y` 等价于 `bind-top`
- `bind-width` / `bind-height` 在绝对布局里也会参与布局约束

## 4. 列表数据绑定

`select` 和 `listbox` 支持：

- `bind-items`
- `bind-selected`

### 4.1 直接绑定字符串数组

```go
state.Set("cities", []string{"Shanghai", "Beijing", "Shenzhen"})
state.Set("selected_city", "Beijing")
```

```html
<listbox bind-items="cities" bind-selected="selected_city"></listbox>
```

### 4.2 绑定结构化数据

Go:

```go
state.Set("users", []map[string]any{
	{"id": "u1", "name": "Alice"},
	{"id": "u2", "name": "Bob"},
})
state.Set("selected_user", "u2")
```

Markup:

```html
<listbox
  bind-items="users"
  bind-selected="selected_user"
  item-text-field="name"
  item-value-field="id"></listbox>
```

可选字段：

- `item-text-field`
- `item-value-field`
- `item-disabled-field`

如果不写这些字段，绑定层会优先尝试常见名字，例如 `text` / `label` / `name` / `title` 和 `value` / `id` / `key`。

## 5. 缺省值与回退规则

绑定存在时：

- 路径有值，使用绑定值
- 路径不存在，回退到 markup 里原本的静态值
- `bind-items` 路径显式为 `nil` 时，会清空列表

这意味着你可以先写一份静态 UI，再逐步把其中一部分替换成绑定。

## 6. 推荐写法

推荐：

- 用清晰的路径名，例如 `page.title`、`form.query`、`dialog.open.visible`
- 把窗口级、页面级、表单级数据分层放
- 尺寸需要跟数据联动时，显式绑定 `width` / `height`
- 列表项是对象数组时，明确写上 `item-text-field` 和 `item-value-field`

不推荐：

- 把很多无关字段全堆在根级
- 依赖文本变化自动推导控件自然尺寸
- 让多个业务来源同时改同一路径

## 7. 一个更完整的例子

```go
state := markup.NewState(map[string]any{
	"page": map[string]any{
		"title":     "用户列表",
		"show_hint": true,
	},
	"hint": "双击可打开详情",
	"users": []map[string]any{
		{"id": "u1", "name": "Alice"},
		{"id": "u2", "name": "Bob"},
	},
	"selected_user": "u1",
})

_, err := markup.LoadIntoScene(scene, htmlText, "", markup.LoadOptions{
	State: state,
})
if err != nil {
	panic(err)
}

state.Patch(map[string]any{
	"page.title":     "用户列表（已刷新）",
	"page.show_hint": false,
	"selected_user":  "u2",
})
```

```html
<window bind-title="page.title">
  <body>
    <label bind-text="page.title"></label>
    <label bind-text="hint" bind-visible="page.show_hint"></label>
    <listbox
      bind-items="users"
      bind-selected="selected_user"
      item-text-field="name"
      item-value-field="id"></listbox>
  </body>
</window>
```

## 8. 当前边界

当前绑定系统有意保持简单：

- 不支持表达式求值，例如 `a + b`
- 不支持在 markup 里直接写循环模板
- 不自动监听普通 Go struct 字段赋值

如果你要自动刷新 UI，请通过 `markup.State` 更新数据，而不是直接修改某个 struct 字段后期待 UI 自己感知。
