package main

import (
	"log"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF75B7"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#FF75B7")).
			Foreground(lipgloss.Color("#FFFFFF"))

	DefaultConfig = Config{
		GridColumns: 4,
		GridRows:    3,
		MaxResults:  12,
	}
)

type EmojiLoadedMsg struct {
	Emojis map[string][]string
	Err    error
}

type State int

const (
	StateLoading State = iota
	StateReady
	StateError
)

type Config struct {
	GridColumns int
	GridRows    int
	MaxResults  int
}

type model struct {
	cfg        Config
	state      State
	emojis     map[string][]string
	filtered   []string
	input      textinput.Model
	selected   int
	focusInput bool
	err        error
	page       int
	keys       keyMap
	help       help.Model
}

type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Select      key.Binding
	ToggleFocus key.Binding
	Quit        key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ToggleFocus, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right}, // first column
		{k.ToggleFocus, k.Quit},         // second column
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k", "ctrl+p"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j", "ctrl+n"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	ToggleFocus: key.NewBinding(
		key.WithKeys("tab", "/"),
		key.WithHelp("tab", "switch focus"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c", "q"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

func loadEmojis() tea.Msg {
	emojis, err := GetEmojis()
	return EmojiLoadedMsg{Emojis: emojis, Err: err}
}

func initialModel(cfg Config) model {
	ti := textinput.New()
	ti.Placeholder = "Type to search emojis..."
	ti.Focus()

	return model{
		cfg:        cfg,
		state:      StateLoading,
		input:      ti,
		focusInput: true,
		keys:       keys,
		help:       help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return loadEmojis
}

func (m model) filterEmojis() []string {
	if m.input.Value() == "" {
		result := make([]string, 0, len(m.emojis))
		for emoji := range m.emojis {
			result = append(result, emoji)
		}
		return result
	}

	result := make([]string, 0)
	search := strings.ToLower(m.input.Value())

	for emoji, keywords := range m.emojis {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(keyword), search) {
				result = append(result, emoji)
				break
			}
		}
	}
	return result
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EmojiLoadedMsg:
		if msg.Err != nil {
			m.state = StateError
			m.err = msg.Err
			return m, nil
		}
		m.state = StateReady
		m.emojis = msg.Emojis
		m.filtered = m.filterEmojis()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			if msg.String() == "q" && m.focusInput {
				// do nothing
			} else {
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.ToggleFocus):
			m.focusInput = !m.focusInput
			var cmd tea.Cmd
			return m, cmd
		case key.Matches(msg, m.keys.Select):
			if !m.focusInput && len(m.filtered) > 0 {
				clipboard.WriteAll(m.filtered[m.selected])
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Up):
			if !m.focusInput {
				newPos := max(0, m.selected-m.cfg.GridColumns)
				m.selected = newPos
			}
		case key.Matches(msg, m.keys.Down):
			if !m.focusInput {
				newPos := min(len(m.filtered)-1, m.selected+m.cfg.GridColumns)
				m.selected = newPos
			}
		case key.Matches(msg, m.keys.Left):
			if !m.focusInput {
				m.selected = max(0, m.selected-1)
			}
		case key.Matches(msg, m.keys.Right):
			if !m.focusInput {
				m.selected = min(len(m.filtered)-1, m.selected+1)
			}
		}
	}

	var cmd tea.Cmd
	if m.focusInput {
		oldValue := m.input.Value()
		m.input, cmd = m.input.Update(msg)

		if oldValue != m.input.Value() {
			m.filtered = m.filterEmojis()
		}
	}

	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(m.input.View() + "\n")

	switch m.state {
	case StateLoading:
		s.WriteString("Loading emojis...")
	case StateError:
		return errorStyle.Render("Error: " + m.err.Error())
	default:
		start := m.page * m.cfg.MaxResults
		end := min(start+m.cfg.MaxResults, len(m.filtered))

		for i := start; i < end; i += m.cfg.GridColumns {
			for j := 0; j < m.cfg.GridColumns && i+j < end; j++ {
				emoji := m.filtered[i+j]
				if i+j == m.selected && !m.focusInput {
					s.WriteString(selectedStyle.Render(" " + emoji + " "))
				} else {
					s.WriteString(" " + emoji + " ")
				}
			}
			s.WriteString("\n")
		}
	}

	helpView := m.help.View(m.keys)
	height := 8 - strings.Count(s.String(), "\n") - strings.Count(helpView, "\n")

	s.WriteString(strings.Repeat("\n", height))
	s.WriteString(helpView)

	return s.String()
}

func main() {
	p := tea.NewProgram(initialModel(DefaultConfig))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
