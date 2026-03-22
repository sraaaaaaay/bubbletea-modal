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
	HPos lipgloss.Position
	VPos lipgloss.Position

	containerWidth  int
	containerHeight int
	Foreground      func() string
	onConfirm       func() tea.Msg
	onCancel        func() tea.Msg
	background      string
	open            bool
}

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
	ColourNone ColourMode = iota
	ColourBasic
	Colour256
	ColourTruecolour
)

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

func New(hPos, vPos lipgloss.Position, fg func() string, onConfirm func() tea.Msg, onCancel func() tea.Msg) (m Model) {
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

func (m *Model) Open(background string) tea.Msg {
	m.open = true
	m.background = background

	m.containerWidth = lipgloss.Width(background)
	m.containerHeight = lipgloss.Height(background)
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

func (m Model) View() string {
	if !m.open {
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
	case ColourBasic, Colour256:
		if tc.Style.FgColour != nil {
			style = style.Foreground(lipgloss.Color(strconv.Itoa(*tc.Style.FgColour)))
		}

	case ColourTruecolour:
		style = style.Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", tc.Style.FgR, tc.Style.FgG, tc.Style.FgB)))
	}

	switch tc.Style.BgColourMode {
	case ColourBasic, Colour256:
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

			if inputAsRunes[i+1] == AnsiCsiIntroducer {
				isAnsiCode = true
			} else if inputAsRunes[i+1] == Osc8Introducer {
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

	for i := 0; i < len(paramValues); i++ {
		switch {
		case paramValues[i] == 0:
			style = StyleState{}
		case paramValues[i] == 1:
			style.Bold = true
		case paramValues[i] == 4:
			style.Underlined = true

		case paramValues[i] >= 30 && paramValues[i] <= 37:
			style.FgColourMode = ColourBasic
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

		// Background colours (40-47, 48;5 and 48;2)
		case paramValues[i] >= 40 && paramValues[i] <= 47:
			style.BgColourMode = ColourBasic
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
