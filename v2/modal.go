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

	confirmKey string
	cancelKey  string

	onOpen    tea.Cmd
	onConfirm tea.Cmd
	onCancel  tea.Cmd

	background tea.View
	isOpen     bool

	dimBackground bool
}

type Option func(*Model)

// A convenience message returned when the modal is opened.
// This can be used if the parent only has a single modal,
// but consumers should define their own message types if
// they expect to have multiple Models in their component
type OpenedMsg struct{}

// A convenience message returned when the modal is confirmed.
// This can be used if the parent only has a single modal,
// but consumers should define their own message types if
// they expect to have multiple Models in their component
type ConfirmMsg struct{}

// A convenience message returned when the modal is closed.
// This can be used if the parent only has a single modal,
// but consumers should define their own message types if
// they expect to have multiple Models in their component
type CloseMsg struct{}

// The standard use case for the callback is to auto-close the dialog,
// so this is provided as a convenience. This can be used if the parent
// only has a single modal, but consumers should define their own message
// types if they expect to have multiple Models in their component
type AutocloseMsg struct{}

const (
	NewlineCharacter  = 10
	EscapeCharacter   = 27
	AnsiCsiIntroducer = '['
	SgrTerminator     = 'm'
	Osc8Introducer    = ']'
	Osc8Terminator    = '\a'
)

// Creates a new modal.
func New(opts ...Option) Model {
	m := Model{
		HPos:       lipgloss.Center,
		VPos:       lipgloss.Center,
		Foreground: func() tea.View { return tea.View{Content: ""} },
		onConfirm:  func() tea.Msg { return nil },
		onCancel:   func() tea.Msg { return nil },
		confirmKey: "Y",
		cancelKey:  "N",
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Sets the horizontal and vertical position of the modal
// within the background container.
func WithPosition(HPos, VPos lipgloss.Position) Option {
	return func(m *Model) {
		m.HPos = HPos
		m.VPos = VPos
	}
}

// Sets the foreground (the contents) of the modal. This is
// a function pointer, so your parent component can define
// its appearance and content in any way you want.
func WithForeground(fg func() tea.View) Option {
	return func(m *Model) {
		m.Foreground = fg
	}
}

// Sets the tea.Cmd to return when the dialog is
// confirmed (when confirmKey) is pressed.
func WithConfirmCmd(cmd tea.Cmd) Option {
	return func(m *Model) {
		m.onConfirm = cmd
	}
}

// Sets the tea.Cmd to return when the dialog is
// canceled (when cancelKey) is pressed. The dialog
// will close itself automatically, but all other
// behaviour is left to the user.
func WithCancelCmd(cmd tea.Cmd) Option {
	return func(m *Model) {
		m.onCancel = cmd
	}
}

// Sets the keymap for confirm/cancel behaviour.
func WithKeyMap(confirm string, cancel string) Option {
	return func(m *Model) {
		m.confirmKey = confirm
		m.cancelKey = cancel
	}
}

// Automatically dims the background when the dialog is open.
func WithDimmedBackground(dim bool) Option {
	return func(m *Model) {
		m.dimBackground = dim
	}
}

// Defines a tea.Cmd to return from Open().
// This is expected to be used with time.Sleep() for a toast-style
// display, but it's flexible enough for other uses.
func WithOpenCmd(cmd tea.Cmd) Option {
	return func(m *Model) {
		m.onOpen = cmd
	}
}

func (m Model) Opened() bool {
	return m.isOpen
}

func (m *Model) Open(background tea.View) tea.Cmd {
	m.isOpen = true
	m.background = background

	m.containerWidth = lipgloss.Width(background.Content)
	m.containerHeight = lipgloss.Height(background.Content)

	return tea.Batch(
		func() tea.Msg { return OpenedMsg{} },
		m.onOpen,
	)
}

func (m *Model) Close() {
	m.isOpen = false
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case m.confirmKey:
			return m, m.onConfirm

		case m.cancelKey:
			m.Close()
			return m, m.onCancel
		}
	}

	return m, nil
}

func (m Model) View() tea.View {
	if !m.isOpen {
		return tea.View{Content: ""}
	}

	return tea.View{Content: m.Composite()}
}

func (m *Model) Composite() string {
	foreground := m.Foreground()
	background := lipgloss.
		NewStyle().
		Faint(m.dimBackground).
		Render(m.background.Content)

	bg := lipgloss.NewLayer(background)
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
