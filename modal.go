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
	HPos       lipgloss.Position
	VPos       lipgloss.Position
	Foreground func() string

	confirmKey string
	cancelKey  string

	onConfirm tea.Cmd
	onCancel  tea.Cmd

	background      string
	isOpen          bool
	containerWidth  int
	containerHeight int
}

type Option func(*Model)

type OpenedMsg struct{}
type ConfirmMsg struct{}
type CloseMsg struct{}

type TerminalCell struct {
	Rune       rune
	Style      StyleState
	HasContent bool
}

type ColourMode int

type StyleState struct {
	FgColourMode ColourMode
	BgColourMode ColourMode

	FgColour      *int
	FgR, FgG, FgB int

	BgColour      *int
	BgR, BgG, BgB int

	Bold       bool
	Underlined bool
}

const (
	ColourNone       ColourMode = iota // No ANSI escape sequence is applied
	Colour16                           // 8 colours and 8 bright colours
	Colour256                          // 8-bit colours
	ColourTruecolour                   // 24-bit (hex colours)
)

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
		Foreground: func() string { return "" },
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
func WithForeground(fg func() string) Option {
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

func (m Model) Opened() bool {
	return m.isOpen
}

func (m *Model) Open(background string) tea.Msg {
	m.isOpen = true
	m.background = background

	m.containerWidth = lipgloss.Width(background)
	m.containerHeight = lipgloss.Height(background)
	return OpenedMsg{}
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

func (m Model) View() string {
	if !m.isOpen {
		return ""
	}

	return m.Composite()
}

func (m *Model) Composite() string {
	foreground := m.Foreground()

	bgGrid := ToTerminalCellGrid(m.background, m.containerWidth, m.containerHeight)
	fgGrid := ToTerminalCellGrid(foreground, m.containerWidth, m.containerHeight)

	bgWidth := len(bgGrid[0])
	bgHeight := len(bgGrid)

	// The foreground grid gets extended to match the background scale,
	// so we need to get the "original" dimensions
	fgWidth := lipgloss.Width(foreground)
	fgHeight := lipgloss.Height(foreground)

	yOffset := applyPosition(m.VPos, bgHeight, fgHeight)
	xOffset := applyPosition(m.HPos, bgWidth, fgWidth)

	for rowIdx := range fgHeight {
		for i := xOffset; i < xOffset+lipgloss.Width(foreground); i++ {
			fgCell := fgGrid[rowIdx][i-xOffset]
			if fgCell.HasContent {
				bgGrid[rowIdx+yOffset][i] = fgCell
			}
		}
	}

	var builder strings.Builder
	for i, row := range bgGrid {
		for _, cell := range row {
			builder.WriteString(cell.Rebuild())
		}
		if i < len(bgGrid)-1 {
			builder.WriteRune('\n')
		}
	}

	return builder.String()
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

func (tc *TerminalCell) Rebuild() string {
	style := lipgloss.NewStyle().
		Bold(tc.Style.Bold).
		Underline(tc.Style.Underlined)

	switch tc.Style.FgColourMode {
	case Colour16, Colour256:
		if tc.Style.FgColour != nil {
			style = style.Foreground(lipgloss.Color(strconv.Itoa(*tc.Style.FgColour)))
		}

	case ColourTruecolour:
		style = style.Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", tc.Style.FgR, tc.Style.FgG, tc.Style.FgB)))
	}

	switch tc.Style.BgColourMode {
	case Colour16, Colour256:
		if tc.Style.BgColour != nil {
			style = style.Background(lipgloss.Color(strconv.Itoa(*tc.Style.BgColour)))
		}

	case ColourTruecolour:
		style = style.Background(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", tc.Style.BgR, tc.Style.BgG, tc.Style.BgB)))
	}

	return style.Render(string(tc.Rune))
}

func ToTerminalCellGrid(input string, width int, height int) [][]TerminalCell {
	grid := make([][]TerminalCell, height)
	for i := range grid {
		grid[i] = make([]TerminalCell, width)
		for j := range grid[i] {
			grid[i][j] = TerminalCell{HasContent: false}
		}
	}

	currentX, currentY := 0, 0
	currentStyle := StyleState{}

	inputAsRunes := []rune(input)

	for i := 0; i < len(inputAsRunes); i++ {
		currentRune := inputAsRunes[i]

		if currentRune == NewlineCharacter {
			currentX = 0
			currentY++
			continue
		}

		// We only need to consider doing something
		// if we encounter the start of an escape sequence. Otherwise,
		// it's fine to just add the rune as-is.
		if currentRune == EscapeCharacter {
			isAnsiCode := false
			isOsc8Code := false

			// Handle malformed sequences, etc
			if i+1 >= len(inputAsRunes) {
				continue
			}

			switch inputAsRunes[i+1] {
			case AnsiCsiIntroducer:
				isAnsiCode = true
			case Osc8Introducer:
				isOsc8Code = true
			}

			if isAnsiCode || isOsc8Code {
				sequenceParamStart := i + 2
				sequenceEndIndex := -1

				// Consume all the parameters for this sequence
				for j := sequenceParamStart; j < len(inputAsRunes); j++ {
					currentParam := inputAsRunes[j]

					if isAnsiCode && currentParam == SgrTerminator {
						sequenceEndIndex = j
						break
					}

					if isOsc8Code && currentParam == Osc8Terminator {
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

func parseStyleState(style StyleState, params string) StyleState {
	if params == "" || params == "0" {
		return StyleState{}
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
			style = StyleState{}
		case paramValues[i] == 1:
			style.Bold = true
		case paramValues[i] == 4:
			style.Underlined = true

		// Foreground colour types
		case paramValues[i] >= 30 && paramValues[i] <= 37:
			style.FgColourMode = Colour16
			c := paramValues[i]
			style.FgColour = &c

		case paramValues[i] == 38 && i+2 < len(paramValues) && paramValues[i+1] == 5:
			style.FgColourMode = Colour256
			c := paramValues[i+2]
			style.FgColour = &c
			i += 2

		case paramValues[i] == 38 && i+4 < len(paramValues) && paramValues[i+1] == 2:
			style.FgColourMode = ColourTruecolour
			style.FgR = paramValues[i+2]
			style.FgG = paramValues[i+3]
			style.FgB = paramValues[i+4]
			i += 4

		// Background colour types
		case paramValues[i] >= 40 && paramValues[i] <= 47:
			style.BgColourMode = Colour16
			c := paramValues[i]
			style.BgColour = &c

		case paramValues[i] == 48 && i+2 < len(paramValues) && paramValues[i+1] == 5:
			style.BgColourMode = Colour256
			c := paramValues[i+2]
			style.BgColour = &c
			i += 2

		case paramValues[i] == 48 && i+4 < len(paramValues) && paramValues[i+1] == 2:
			style.BgColourMode = ColourTruecolour
			style.BgR = paramValues[i+2]
			style.BgG = paramValues[i+3]
			style.BgB = paramValues[i+4]
			i += 4
		}
	}

	return style
}
