package modal

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

func assertEqual(t *testing.T, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func Test_OddWidthModalPosition_Displays_AsExpected(t *testing.T) {
	// Arrange
	background := "0000000000\n0000000000\n0000000000\n0000000000\n0000000000"
	expectedModalDisplay := "0000000000\n0000000000\n0001234000\n0000000000\n0000000000"

	foreground := func() tea.View { return tea.View{Content: "1234"} }

	// Act
	dialog := New(lipgloss.Center, lipgloss.Center, foreground, nil, nil)
	dialog.Open(tea.View{Content: background})

	// Assert
	assertEqual(t, dialog.View().Content, expectedModalDisplay)
}

func Test_EvenWidthModalPosition_Displays_AsExpected(t *testing.T) {
	// Arrange
	background := "000000000\n000000000\n000000000\n000000000\n000000000"
	expectedModalDisplay := "000000000\n000000000\n001234000\n000000000\n000000000"

	foreground := func() tea.View { return tea.View{Content: "1234"} }

	// Act
	dialog := New(lipgloss.Center, lipgloss.Center, foreground, nil, nil)
	dialog.Open(tea.View{Content: background})

	// Assert
	assertEqual(t, dialog.View().Content, expectedModalDisplay)
}

func Test_StackedModal_Displays_AsExpected(t *testing.T) {
	// Arrange
	background := "0000000000\n0000000000\n0000000000\n0000000000\n0000000000"
	expectedModalDisplay := "0000000000\n0000000000\n0001234000\n0000000000\n0000000000"
	expectedStackedModalDisplay := "0000000000\n0000000000\n0001994000\n0000000000\n0000000000"

	foreground := func() tea.View { return tea.View{Content: "1234"} }
	stackedForeground := func() tea.View { return tea.View{Content: "99"} }

	// Act
	dialog := New(lipgloss.Center, lipgloss.Center, foreground, nil, nil)
	stackedDialog := New(lipgloss.Center, lipgloss.Center, stackedForeground, nil, nil)

	dialog.Open(tea.View{Content: background})
	stackedDialog.Open(dialog.View())

	// Assert
	assertEqual(t, dialog.View().Content, expectedModalDisplay)
	assertEqual(t, stackedDialog.View().Content, expectedStackedModalDisplay)
}

func Test_Modal_RespondsTo_ConfirmHandler(t *testing.T) {
	// Arrange
	onConfirmHandler := func() tea.Msg {
		return ConfirmMsg{}
	}

	modal := New(lipgloss.Center, lipgloss.Center, func() tea.View { return tea.View{Content: "1"} }, onConfirmHandler, nil)

	// Act
	modal.Open(tea.View{Content: "000"})

	updated, cmd := modal.Update(tea.KeyPressMsg{Code: 'Y'})
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

	modal := New(lipgloss.Center, lipgloss.Center, func() tea.View { return tea.View{Content: "1"} }, nil, onCloseHandler)

	// Act
	modal.Open(tea.View{Content: "000"})
	updated, cmd := modal.Update(tea.KeyPressMsg{Code: 'N'})
	msg := cmd()
	_, ok := msg.(CloseMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(Model).Opened(), false)
}
