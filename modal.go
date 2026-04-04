package modal

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type Model struct {
	hPos       lipgloss.Position
	vPos       lipgloss.Position
	foreground func() string

	confirmKey string
	cancelKey  string

	onOpen    tea.Cmd
	onConfirm tea.Cmd
	onCancel  tea.Cmd

	background string
	isOpen     bool

	dimBackground bool
	clickToClose  bool
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

type TerminalCell struct {
	Rune       rune
	Style      lipgloss.Style
	HasContent bool
}

const (
	newlineCharacter  = 10
	escapeCharacter   = 27
	ansiCsiIntroducer = '['
	sgrTerminator     = 'm'
	osc8Introducer    = ']'
	osc8Terminator    = '\a'
)

// Creates a new modal.
func New(opts ...Option) Model {
	m := Model{
		hPos:       lipgloss.Center,
		vPos:       lipgloss.Center,
		foreground: func() string { return "" },
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
		m.hPos = HPos
		m.vPos = VPos
	}
}

// Sets the foreground (the contents) of the modal. This is
// a function pointer, so your parent component can define
// its appearance and content in any way you want.
func WithForeground(fg func() string) Option {
	return func(m *Model) {
		m.foreground = fg
	}
}

// Sets the tea.Cmd to return when the dialog is
// confirmed (when confirmKey) is pressed. By
// default, the confirm command is nil.
func WithConfirmCmd(cmd tea.Cmd) Option {
	return func(m *Model) {
		m.onConfirm = cmd
	}
}

// Sets the tea.Cmd to return when the dialog is
// canceled (when cancelKey) is pressed. The dialog
// will close itself automatically, but all other
// behaviour is left to the user. By default,
// the cancel command is nil.
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
//
// This is expected to be used with time.Sleep() to make
// the modal close automatically after a delay, but can
// be used for anything you want to happen when the modal
// is opened.
func WithOpenCmd(cmd tea.Cmd) Option {
	return func(m *Model) {
		m.onOpen = cmd
	}
}

// Allows the dialog to be closed by clicking outside it.
func WithClickToClose(clickToClose bool) Option {
	return func(m *Model) {
		m.clickToClose = clickToClose
	}
}

func (m Model) Opened() bool {
	return m.isOpen
}

func (m Model) Open(background string) (Model, tea.Cmd) {
	m.isOpen = true
	m.background = background

	return m, tea.Batch(
		func() tea.Msg { return OpenedMsg{} },
		m.onOpen,
	)
}

func (m Model) Close() Model {
	m.isOpen = false
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) foregroundBounds() (x, y, w, h int) {
	foreground := m.foreground()
	containerWidth := lipgloss.Width(m.background)
	containerHeight := lipgloss.Height(m.background)
	fgWidth := lipgloss.Width(foreground)
	fgHeight := lipgloss.Height(foreground)

	bgGrid := toTerminalCellGrid(m.background, containerWidth, containerHeight)
	bgWidth := len(bgGrid[0])
	bgHeight := len(bgGrid)

	return applyPosition(m.hPos, bgWidth, fgWidth),
		applyPosition(m.vPos, bgHeight, fgHeight),
		fgWidth,
		fgHeight
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if !m.clickToClose {
			break
		}

		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			fgX, fgY, fgWidth, fgHeight := m.foregroundBounds()
			clickedForeground := msg.X >= fgX && msg.X < fgX+fgWidth &&
				msg.Y >= fgY && msg.Y < fgY+fgHeight

			if !clickedForeground {
				m = m.Close()
				return m, m.onCancel
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case m.confirmKey:
			return m, m.onConfirm

		case m.cancelKey:
			m = m.Close()
			return m, m.onCancel
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.isOpen {
		return ""
	}

	return m.Composite()
}

func (m Model) Composite() string {
	foreground := m.foreground()

	containerWidth := lipgloss.Width(m.background)
	containerHeight := lipgloss.Height(m.background)

	fgWidth := lipgloss.Width(foreground)
	fgHeight := lipgloss.Height(foreground)

	bgGrid := toTerminalCellGrid(m.background, containerWidth, containerHeight)
	fgGrid := toTerminalCellGrid(foreground, fgWidth, fgHeight)

	bgWidth := len(bgGrid[0])
	bgHeight := len(bgGrid)

	yOffset := applyPosition(m.vPos, bgHeight, fgHeight)
	xOffset := applyPosition(m.hPos, bgWidth, fgWidth)

	if m.dimBackground {
		for rowIdx := range bgGrid {
			for i := range bgGrid[rowIdx] {
				isBehindForeground := rowIdx >= yOffset &&
					rowIdx < yOffset+fgHeight &&
					i >= xOffset &&
					i < xOffset+fgWidth

				if !isBehindForeground {
					bgGrid[rowIdx][i].Style = bgGrid[rowIdx][i].Style.Faint(true)
				}
			}
		}
	}

	for rowIdx := range fgHeight {
		if rowIdx+yOffset >= bgHeight {
			break
		}
		for i := xOffset; i < min(xOffset+fgWidth, bgWidth); i++ {
			fgCell := fgGrid[rowIdx][i-xOffset]
			if fgCell.HasContent {
				bgGrid[rowIdx+yOffset][i] = fgCell
			}
		}
	}

	var finalView strings.Builder

	// Frequent calls to style.Render() are expensive - try to
	// accumulate cells with the same style into a buffer before
	// applying the style to relieve this.
	var bufferStyle lipgloss.Style
	var buffer strings.Builder
	flushBuffer := func() {
		finalView.WriteString(bufferStyle.Render(buffer.String()))
		buffer.Reset()
	}

	for i, row := range bgGrid {
		for _, cell := range row {
			if cell.Style.Value() != bufferStyle.Value() {
				flushBuffer()
				bufferStyle = cell.Style
			}

			buffer.WriteRune(cell.Rune)
		}

		if buffer.Len() > 0 {
			flushBuffer()
		}

		if i < len(bgGrid)-1 {
			finalView.WriteRune('\n')
		}
	}

	return finalView.String()
}

func applyPosition(pos lipgloss.Position, bgDimension, fgDimension int) int {
	var offset int
	if pos == lipgloss.Left || pos == lipgloss.Top {
		offset = 0
	} else if pos > 0 && pos < 1 {
		bgOffset := float64(bgDimension) * float64(pos)
		fgOffset := float64(fgDimension) * float64(pos)
		offset = int(bgOffset - fgOffset)
	} else {
		offset = bgDimension - fgDimension
	}
	return max(0, offset)
}

func toTerminalCellGrid(input string, width int, height int) [][]TerminalCell {
	grid := make([][]TerminalCell, height)
	for i := range grid {
		grid[i] = make([]TerminalCell, width)
		for j := range grid[i] {
			grid[i][j] = TerminalCell{HasContent: false}
		}
	}

	currentX, currentY := 0, 0
	currentStyle := lipgloss.NewStyle()

	inputAsRunes := []rune(input)

	for i := 0; i < len(inputAsRunes); i++ {
		currentRune := inputAsRunes[i]

		if currentRune == newlineCharacter {
			currentX = 0
			currentY++
			continue
		}

		// We only need to consider doing something
		// if we encounter the start of an escape sequence. Otherwise,
		// it's fine to just add the rune as-is.
		if currentRune == escapeCharacter {
			isAnsiCode := false
			isOsc8Code := false

			// Handle malformed sequences, etc
			if i+1 >= len(inputAsRunes) {
				continue
			}

			switch inputAsRunes[i+1] {
			case ansiCsiIntroducer:
				isAnsiCode = true
			case osc8Introducer:
				isOsc8Code = true
			}

			if isAnsiCode || isOsc8Code {
				sequenceParamStart := i + 2
				sequenceEndIndex := -1

				// Consume all the parameters for this sequence
				for j := sequenceParamStart; j < len(inputAsRunes); j++ {
					currentParam := inputAsRunes[j]

					if isAnsiCode && currentParam == sgrTerminator {
						sequenceEndIndex = j
						break
					}

					if isOsc8Code && currentParam == osc8Terminator {
						sequenceEndIndex = j
						break
					}

					// Almost everything from Lipgloss will be SGR, but handle
					// other ANSI function handles as a precaution
					if isAnsiCode && currentParam >= 0x40 && currentParam <= 0x7E {
						sequenceEndIndex = j
						break
					}
				}

				// If we've managed to parse a valid ANSI control sequence,
				// set that to the current style so it can be stored for each
				// terminal cell
				if sequenceEndIndex != -1 {
					parametersString := string(inputAsRunes[sequenceParamStart:sequenceEndIndex])
					currentStyle = parseStyleState(currentStyle, parametersString)

					i = sequenceEndIndex
					continue
				}
			}
		}

		currentRuneWidth := runewidth.RuneWidth(currentRune)
		if currentX+currentRuneWidth > width {
			currentX = 0
			currentY++
		}

		if currentY >= height {
			break
		}

		if currentX < width {
			grid[currentY][currentX] = TerminalCell{Rune: currentRune, Style: currentStyle, HasContent: true}

			// Wide rune (e.g. CJK) - blank the adjacent spot
			if currentRuneWidth == 2 && currentX+1 < width {
				grid[currentY][currentX+1] = TerminalCell{Rune: 0, Style: currentStyle, HasContent: true}
			}
			currentX += currentRuneWidth
		}
	}

	return grid
}

func parseStyleState(style lipgloss.Style, params string) lipgloss.Style {
	if params == "" || params == "0" {
		return lipgloss.NewStyle()
	}

	paramValues := []int{}
	for param := range strings.SplitSeq(params, ";") {
		val, _ := strconv.Atoi(param)
		paramValues = append(paramValues, val)
	}

	// When interpolating the ANSI sequences to make coloured text,
	// Lipgloss retains the specific "mode" the text colour was initialised
	// in. For example the internal representation of lipgloss.Color("9")
	// is different from lipgloss.Color("#FF0000") even if they're equivalent.
	//
	// This means we need to detect possible colour types based on the code ranges
	// rather than assume it's been converted to RGB or some standard.
	for i := 0; i < len(paramValues); i++ {
		switch {
		case paramValues[i] == 0:
			style = lipgloss.NewStyle()
		case paramValues[i] == 1:
			style = style.Bold(true)
		case paramValues[i] == 4:
			style = style.Underline(true)

		// Foreground colour types
		case paramValues[i] >= 30 && paramValues[i] <= 37:
			colour := lipgloss.Color(strconv.Itoa(paramValues[i]))
			style = style.Foreground(colour)

		case paramValues[i] == 38 && i+2 < len(paramValues) && paramValues[i+1] == 5:
			colour := lipgloss.Color(strconv.Itoa(paramValues[i+2]))
			style = style.Foreground(colour)
			i += 2

		case paramValues[i] == 38 && i+4 < len(paramValues) && paramValues[i+1] == 2:
			colour := fmt.Sprintf(
				"#%02x%02x%02x",
				paramValues[i+2],
				paramValues[i+3],
				paramValues[i+4],
			)
			style = style.Foreground(lipgloss.Color(colour))
			i += 4

		// Background colour types
		case paramValues[i] >= 40 && paramValues[i] <= 47:
			colour := lipgloss.Color(strconv.Itoa(paramValues[i]))
			style = style.Background(colour)

		case paramValues[i] == 48 && i+2 < len(paramValues) && paramValues[i+1] == 5:
			colour := lipgloss.Color(strconv.Itoa(paramValues[i+2]))
			style = style.Background(colour)
			i += 2

		case paramValues[i] == 48 && i+4 < len(paramValues) && paramValues[i+1] == 2:
			colour := fmt.Sprintf(
				"#%02x%02x%02x",
				paramValues[i+2],
				paramValues[i+3],
				paramValues[i+4],
			)
			style = style.Background(lipgloss.Color(colour))
			i += 4
		}
	}

	return style
}
