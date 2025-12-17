package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap 定义全局快捷键
type KeyMap struct {
	Quit       key.Binding
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Help       key.Binding
	Cancel     key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Home       key.Binding
	End        key.Binding
	LogPrev    key.Binding
	LogNext    key.Binding
	LogAll     key.Binding
	LogResume  key.Binding
}

// DefaultKeyMap 默认快捷键配置
var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "退出"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "上移"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "下移"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "确认"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "帮助"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "取消"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "向上滚动"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "向下滚动"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup"),
		key.WithHelp("pgup", "上一页"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown"),
		key.WithHelp("pgdown", "下一页"),
	),
	Home: key.NewBinding(
		key.WithKeys("home"),
		key.WithHelp("home", "顶部"),
	),
	End: key.NewBinding(
		key.WithKeys("end"),
		key.WithHelp("end", "底部"),
	),
	LogPrev: key.NewBinding(
		key.WithKeys("[", "shift+tab"),
		key.WithHelp("[/⇧tab", "上一个日志页"),
	),
	LogNext: key.NewBinding(
		key.WithKeys("]", "tab"),
		key.WithHelp("]/tab", "下一个日志页"),
	),
	LogAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "全部日志并置底"),
	),
	LogResume: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "恢复自动滚动"),
	),
}

// ShortHelp 返回简短帮助信息
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp 返回完整帮助信息
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.ScrollUp, k.ScrollDown},
		{k.PageUp, k.PageDown, k.Home, k.End},
		{k.LogPrev, k.LogNext, k.LogAll, k.LogResume},
		{k.Help, k.Quit},
	}
}
