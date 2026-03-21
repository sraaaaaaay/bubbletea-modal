package modal

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func assertEqual(t *testing.T, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func Test_ReconstructedTrueColourTerminalCell_IdenticalTo_LipglossStyle(t *testing.T) {
	// Arrange
	tc := TerminalCell{
		Rune: 'A',
		Style: StyleState{
			FgColourMode: ColourTruecolour,
			BgColourMode: ColourTruecolour,

			FgColour: nil,
			// #FF00FF - hot pink
			FgR: 255,
			FgG: 0,
			FgB: 255,

			BgColour: nil,
			// #00FF00 - lime
			BgR: 0,
			BgG: 255,
			BgB: 0,

			Bold:       true,
			Underlined: false,
		},
	}

	foregroundColour := lipgloss.Color("#FF00FF")
	backgroundColour := lipgloss.Color("#00FF00")

	correspondingLipglossStyle := lipgloss.
		NewStyle().
		Foreground(foregroundColour).
		Background(backgroundColour).
		Bold(true)

	// Act
	rebuiltOutput := tc.Rebuild()
	lipglossOutput := correspondingLipglossStyle.Render("A")

	// Assert
	assertEqual(t, rebuiltOutput, lipglossOutput)
}

func Test_TerminalCellGrid_Has_ExpectedWidth(t *testing.T) {
	// Arrange
	stringToRender := "Hello World\n\nBye Now!"
	expectedWidth := 18 // Ignoring newlines
	expectedHeight := 4 // Two lines + two newlines

	// Act
	grid := ToTerminalCellGrid(stringToRender, expectedWidth, expectedHeight)

	// Assert
	assertEqual(t, len(grid[0]), expectedWidth)
}

func Test_TerminalCellGrid_Has_ExpectedHeight(t *testing.T) {
	// Arrange
	stringToRender := "HelloWorld\n\nBye Now!"
	expectedWidth := 18
	expectedHeight := 4

	// Act
	grid := ToTerminalCellGrid(stringToRender, expectedWidth, expectedHeight)

	// Assert
	assertEqual(t, len(grid), expectedHeight)
}

func Test_ParseBasicColourAnsiParameters_Creates_ExpectedStyle(t *testing.T) {
	// Arrange
	sgrParams := "0;4;32;42"
	fgColour := 32 // Green
	bgColour := 42 // Green

	blankStyle := StyleState{}
	expectedStyleState := StyleState{
		FgColourMode: ColourBasic,
		FgColour:     &fgColour,
		BgColourMode: ColourBasic,
		BgColour:     &bgColour,
		Underlined:   true,
	}

	// Act
	parsedStyleState := parseStyleState(blankStyle, sgrParams)

	// Assert
	assertEqual(t, parsedStyleState.FgColourMode, expectedStyleState.FgColourMode)
	assertEqual(t, *parsedStyleState.FgColour, *expectedStyleState.FgColour)
	assertEqual(t, parsedStyleState.BgColourMode, expectedStyleState.BgColourMode)
	assertEqual(t, *parsedStyleState.BgColour, *expectedStyleState.BgColour)
	assertEqual(t, parsedStyleState.Underlined, expectedStyleState.Underlined)
}

func Test_Parse256ColourAnsiParameters_Creates_ExpectedStyle(t *testing.T) {
	// Arrange
	sgrParams := "0;38;5;99;48;5;99"
	colour := 99 // Lavender

	blankStyle := StyleState{}
	expectedStyleState := StyleState{
		FgColourMode: Colour256,
		FgColour:     &colour,
		BgColourMode: Colour256,
		BgColour:     &colour,
	}

	// Act
	parsedStyleState := parseStyleState(blankStyle, sgrParams)

	// Assert
	assertEqual(t, parsedStyleState.FgColourMode, expectedStyleState.FgColourMode)
	assertEqual(t, *parsedStyleState.FgColour, *expectedStyleState.FgColour)
	assertEqual(t, parsedStyleState.BgColourMode, expectedStyleState.BgColourMode)
	assertEqual(t, *parsedStyleState.BgColour, *expectedStyleState.BgColour)
}

func Test_ParseTrueColourAnsiParameters_Creates_ExpectedStyle(t *testing.T) {
	// Arrange
	// Reset, bold, then apply the colour
	sgrParams := "0;1;38;2;255;0;255;48;2;255;0;255"
	blankStyle := StyleState{}
	expectedStyleState := StyleState{
		FgColourMode: ColourTruecolour,
		// #FF00FF - hot pink
		FgR: 255,
		FgG: 0,
		FgB: 255,

		BgColourMode: ColourTruecolour,
		BgR:          255,
		BgG:          0,
		BgB:          255,
		Bold:         true,
	}

	// Act
	parsedStyleState := parseStyleState(blankStyle, sgrParams)

	// Assert
	assertEqual(t, parsedStyleState.FgColourMode, expectedStyleState.FgColourMode)
	assertEqual(t, parsedStyleState.BgColourMode, expectedStyleState.BgColourMode)
	assertEqual(t, parsedStyleState.FgR, expectedStyleState.FgR)
	assertEqual(t, parsedStyleState.FgG, expectedStyleState.FgG)
	assertEqual(t, parsedStyleState.FgB, expectedStyleState.FgB)
	assertEqual(t, parsedStyleState.BgR, expectedStyleState.BgR)
	assertEqual(t, parsedStyleState.BgG, expectedStyleState.BgG)
	assertEqual(t, parsedStyleState.BgB, expectedStyleState.BgB)
	assertEqual(t, parsedStyleState.Bold, expectedStyleState.Bold)
}

func Test_AnsiParse_Skips_MalformedSequences(t *testing.T) {
	// Arrange
	malformedSequence := "A\x1b["
	expectedStyleState := StyleState{}

	// Act
	grid := ToTerminalCellGrid(malformedSequence, 5, 1)

	// Assert
	assertEqual(t, len(grid), len(grid))
	assertEqual(t, grid[0][0].Rune, 'A')
	assertEqual(t, grid[0][0].Style, expectedStyleState)
}

func Test_AnsiParse_Handles_Osc8Url(t *testing.T) {
	// Arrange
	url := "https://www.google.com"
	text := "Here!"

	introducer := "\u001B]8"
	bell := "\u0007"
	noParams := ";;"

	osc8String := introducer + noParams + url + bell + text + introducer + noParams + bell

	// Act
	grid := ToTerminalCellGrid(osc8String, 5, 1)

	var stringResult strings.Builder
	for _, cell := range grid[0] {
		stringResult.WriteRune(cell.Rune)
	}

	// Assert
	assertEqual(t, len(grid), 1)
	assertEqual(t, len(grid[0]), 5)
	assertEqual(t, grid[0][0].Rune, 'H')
	assertEqual(t, grid[0][4].Rune, '!')
	assertEqual(t, grid[0][0].Style, StyleState{})
	assertEqual(t, stringResult.String(), "Here!")
}

func Test_Modal_Displays_AsExpected(t *testing.T) {
	// Arrange
	background := "0000000000\n0000000000\n0000000000\n0000000000\n0000000000"
	expectedModalDisplay := "0000000000\n0000000000\n0001234000\n0000000000\n0000000000"

	foreground := func() string { return "1234" }

	// Act
	dialog := New(lipgloss.Center, lipgloss.Center, foreground, nil, nil)
	dialog.Open(background)

	// Assert
	assertEqual(t, dialog.View(), expectedModalDisplay)
}

func Test_StackedModal_Displays_AsExpected(t *testing.T) {
	// Arrange
	background := "0000000000\n0000000000\n0000000000\n0000000000\n0000000000"
	expectedModalDisplay := "0000000000\n0000000000\n0001234000\n0000000000\n0000000000"
	expectedStackedModalDisplay := "0000000000\n0000000000\n0001994000\n0000000000\n0000000000"

	foreground := func() string { return "1234" }
	stackedForeground := func() string { return "99" }

	// Act
	dialog := New(lipgloss.Center, lipgloss.Center, foreground, nil, nil)
	stackedDialog := New(lipgloss.Center, lipgloss.Center, stackedForeground, nil, nil)

	dialog.Open(background)
	stackedDialog.Open(dialog.View())

	// Assert
	assertEqual(t, dialog.View(), expectedModalDisplay)
	assertEqual(t, stackedDialog.View(), expectedStackedModalDisplay)
}

func Test_Modal_RespondsTo_ConfirmHandler(t *testing.T) {
	// Arrange
	onConfirmHandler := func() tea.Msg {
		return ConfirmMsg{}
	}

	modal := New(lipgloss.Center, lipgloss.Center, func() string { return "1" }, onConfirmHandler, nil)

	// Act
	modal.Open("000")
	updated, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")})
	msg := cmd()
	_, ok := msg.(ConfirmMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(Model).Opened(), true)
}

func Test_Modal_RespondsTo_CloseHandler(t *testing.T) {
	// Arrange
	onCloseHandler := func() tea.Msg {
		return CloseMsg{}
	}

	modal := New(lipgloss.Center, lipgloss.Center, func() string { return "1" }, nil, onCloseHandler)

	// Act
	modal.Open("000")
	updated, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})
	msg := cmd()
	_, ok := msg.(CloseMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(Model).Opened(), false)
}
