package modal

import (
	"fmt"
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

func assertStyleEqual(t *testing.T, got, want lipgloss.Style) {
	t.Helper()

	// Escape the colour and formatting sequences
	// to show raw output
	gotRaw := fmt.Sprintf("%q", got.Render("A"))
	wantRaw := fmt.Sprintf("%q", want.Render("A"))

	if gotRaw != wantRaw {
		t.Errorf("got %v, want %v", gotRaw, wantRaw)
	}
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

	blankStyle := lipgloss.NewStyle()
	expectedStyleState := blankStyle.
		Underline(true).
		Foreground(lipgloss.Color("32")).
		Background(lipgloss.Color("42"))

	// Act
	parsedStyleState := parseStyleState(blankStyle, sgrParams)

	// Assert
	assertStyleEqual(t, parsedStyleState, expectedStyleState)
}

func Test_Parse256ColourAnsiParameters_Creates_ExpectedStyle(t *testing.T) {
	// Arrange
	sgrParams := "0;38;5;99;48;5;99"

	blankStyle := lipgloss.NewStyle()
	expectedStyleState := blankStyle.
		Foreground(lipgloss.Color("99")).
		Background(lipgloss.Color("99"))

	// Act
	parsedStyleState := parseStyleState(blankStyle, sgrParams)

	// Assert
	assertStyleEqual(t, parsedStyleState, expectedStyleState)
}

func Test_ParseTrueColourAnsiParameters_Creates_ExpectedStyle(t *testing.T) {
	// Arrange
	// Reset, bold, then apply the colour
	sgrParams := "0;1;38;2;255;0;255;48;2;255;0;255"
	blankStyle := lipgloss.NewStyle()

	expectedStyleState := blankStyle.
		Bold(true).
		Foreground(lipgloss.Color("#FF00FF")).
		Background(lipgloss.Color("#FF00FF"))

	// Act
	parsedStyleState := parseStyleState(blankStyle, sgrParams)

	// Assert
	assertStyleEqual(t, parsedStyleState, expectedStyleState)
}

func Test_AnsiParse_Skips_MalformedSequences(t *testing.T) {
	// Arrange
	malformedSequence := "A\x1b["
	expectedStyleState := lipgloss.NewStyle()

	// Act
	grid := ToTerminalCellGrid(malformedSequence, 5, 1)

	// Assert
	assertEqual(t, len(grid), len(grid))
	assertEqual(t, grid[0][0].Rune, 'A')
	assertEqual(t, grid[0][0].Style.Render("A"), expectedStyleState.Render("A"))
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
	assertEqual(t, grid[0][0].Style.Render("A"), lipgloss.NewStyle().Render("A"))
	assertEqual(t, stringResult.String(), "Here!")
}

func Test_Modal_Displays_AsExpected(t *testing.T) {
	// Arrange
	background := "0000000000\n0000000000\n0000000000\n0000000000\n0000000000"
	expectedModalDisplay := "0000000000\n0000000000\n0001234000\n0000000000\n0000000000"

	foreground := func() string { return "1234" }

	// Act
	dialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(foreground),
	)

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
	dialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(foreground),
	)

	stackedDialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(stackedForeground),
	)

	dialog.Open(background)
	stackedDialog.Open(dialog.View())

	// Assert
	assertEqual(t, dialog.View(), expectedModalDisplay)
	assertEqual(t, stackedDialog.View(), expectedStackedModalDisplay)
}

func Test_Modal_RespondsTo_DefaultConfirmKey(t *testing.T) {
	// Arrange
	onConfirmHandler := func() tea.Msg {
		return ConfirmMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() string { return "1" }),
		WithConfirmCmd(onConfirmHandler),
	)

	// Act
	modal.Open("000")
	updated, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")})

	var msg tea.Msg
	if cmd != nil {
		msg = cmd()
	}
	_, ok := msg.(ConfirmMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(*Model).Opened(), true)
}

func Test_Modal_RespondsTo_CustomConfirmKey(t *testing.T) {
	// Arrange
	onConfirmHandler := func() tea.Msg {
		return ConfirmMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() string { return "1" }),
		WithKeyMap("enter", "N"), // Keep using "N" (default) for close
		WithConfirmCmd(onConfirmHandler),
	)

	// Act
	modal.Open("000")
	updated, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")})

	var msg tea.Msg
	if cmd != nil {
		msg = cmd()
	}
	_, ok := msg.(ConfirmMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(*Model).Opened(), true)
}

func Test_Modal_RespondsTo_DefaultCancelKey(t *testing.T) {
	// Arrange
	onCloseHandler := func() tea.Msg {
		return CloseMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() string { return "1" }),
		WithCancelCmd(onCloseHandler),
	)

	// Act
	modal.Open("000")
	updated, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("N")})

	var msg tea.Msg
	if cmd != nil {
		msg = cmd()
	}
	_, ok := msg.(CloseMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(*Model).Opened(), false)
}

func Test_Modal_RespondsTo_CustomCancelKey(t *testing.T) {
	// Arrange
	onCloseHandler := func() tea.Msg {
		return CloseMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() string { return "1" }),
		WithKeyMap("Y", "esc"), // keep Y default
		WithCancelCmd(onCloseHandler),
	)

	// Act
	modal.Open("000")
	updated, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("esc")})

	var msg tea.Msg
	if cmd != nil {
		msg = cmd()
	}
	_, ok := msg.(CloseMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(*Model).Opened(), false)
}

func Test_Modal_SafelyConsumes_UnrelatedKeyPress(t *testing.T) {
	// Arrange
	onCloseHandler := func() tea.Msg {
		return CloseMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() string { return "1" }),
		WithKeyMap("enter", "esc"),
		WithCancelCmd(onCloseHandler),
	)

	// Act
	modal.Open("000")
	_, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("ctrl+q")})

	// Assert
	assertEqual(t, cmd == nil, true)
}

func Test_Modal_DimBackground_DimsBackground(t *testing.T) {
	// Arrange
	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() string { return "1" }),
		WithDimmedBackground(true),
	)

	expectedBgStyle := lipgloss.NewStyle().Faint(true)

	// Act
	modal.Open("000")

	// Assert
	assertEqual(
		t,
		fmt.Sprintf(
			"%s%s%s",
			expectedBgStyle.Render("0"),
			"1",
			expectedBgStyle.Render("0")),
		modal.View(),
	)
}

func Test_Modal_Autocloses(t *testing.T) {
	// Arrange
	modal := New(
		WithForeground(func() string { return "1" }),
		WithOpenCmd(func() tea.Msg {
			return AutocloseMsg{}
		}),
	)

	// Act
	cmd := modal.Open("000")
	msg := cmd()

	// Because Open() returns a tea.Batch, we need to
	// execute all of them and check that the autoclose
	// occurred at some point, although the order doesn't
	// matter.
	batch, _ := msg.(tea.BatchMsg)
	hasAutoClose := false
	for _, cmd := range batch {
		_, ok := cmd().(AutocloseMsg)
		if ok {
			hasAutoClose = true
		}
	}

	// Assert
	assertEqual(t, hasAutoClose, true)
}

func Benchmark_Modal_Composite(b *testing.B) {
	background := strings.Repeat(strings.Repeat("0", 120)+"\n", 39) + strings.Repeat("0", 120)
	foreground := func() string { return strings.Repeat("1", 40) }

	dialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(foreground),
		WithDimmedBackground(true),
	)
	dialog.Open(background)

	for b.Loop() {
		dialog.View()
	}
}
