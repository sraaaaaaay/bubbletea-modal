package modal

import (
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

type Model struct {
	HPos lipgloss.Position
	VPos lipgloss.Position

	containerWidth  int
	containerHeight int
	Foreground      func() tea.View
	onConfirm       func() tea.Msg
	onCancel        func() tea.Msg
	background      tea.View
	open            bool
}

type OpenedMsg struct{}
type ConfirmMsg struct{}
type CloseMsg struct{}

const (
	NewlineCharacter  = 10
	EscapeCharacter   = 27
	AnsiCsiIntroducer = '['
	SgrTerminator     = 'm'
	Osc8Introducer    = ']'
	Osc8Terminator    = '\a'
)

func defaultCmd() tea.Msg {
	return nil
}

func New(hPos, vPos lipgloss.Position, fg func() tea.View, onConfirm func() tea.Msg, onCancel func() tea.Msg) (m Model) {
	m.Foreground = fg
	m.HPos = hPos
	m.VPos = vPos

	if onConfirm != nil {
		m.onConfirm = onConfirm
	} else {
		m.onConfirm = defaultCmd
	}

	if onCancel != nil {
		m.onCancel = onCancel
	} else {
		m.onCancel = defaultCmd
	}

	return m
}

func (m Model) Opened() bool {
	return m.open
}

func (m *Model) Open(background tea.View) tea.Msg {
	m.open = true
	m.background = background

	m.containerWidth = lipgloss.Width(background.Content)
	m.containerHeight = lipgloss.Height(background.Content)
	return OpenedMsg{}
}

func (m *Model) Close() {
	m.open = false
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "Y", "enter":
			return m, m.onConfirm

		case "N", "esc":
			m.Close()
			return m, m.onCancel
		}
	}

	return m, nil
}

func (m Model) View() tea.View {
	if !m.open {
		return tea.View{Content: ""}
	}

	return tea.View{Content: m.Composite()}
}

func (m *Model) Composite() string {
	foreground := m.Foreground()

	bg := lipgloss.NewLayer(m.background.Content)
	fg := lipgloss.
		NewLayer(foreground.Content).
		X(applyPosition(m.HPos, m.containerWidth, lipgloss.Width(foreground.Content))).
		Y(applyPosition(m.VPos, m.containerHeight, lipgloss.Height(foreground.Content))).
		Z(1)

	compositor := lipgloss.NewCompositor(bg, fg)
	return compositor.Render()
}

func applyPosition(pos lipgloss.Position, bgDimension, fgDimension int) int {
	if pos == lipgloss.Left || pos == lipgloss.Top {
		return 0
	} else if pos > 0 && pos < 1 {
		bgOffset := float64(bgDimension) * float64(pos)
		fgOffset := float64(fgDimension) * float64(pos)
		return int(bgOffset - fgOffset)
	} else {
		return bgDimension - fgDimension
	}
}
