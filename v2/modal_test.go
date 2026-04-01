package modal

import (
	"fmt"
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
	dialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(foreground),
	)

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
	dialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(foreground),
	)

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
	dialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(foreground),
	)

	stackedDialog := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(stackedForeground),
	)

	dialog.Open(tea.View{Content: background})
	stackedDialog.Open(dialog.View())

	// Assert
	assertEqual(t, dialog.View().Content, expectedModalDisplay)
	assertEqual(t, stackedDialog.View().Content, expectedStackedModalDisplay)
}

func Test_Modal_RespondsTo_DefaultConfirmKey(t *testing.T) {
	// Arrange
	onConfirmHandler := func() tea.Msg {
		return ConfirmMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() tea.View { return tea.View{Content: "1"} }),
		WithConfirmCmd(onConfirmHandler),
	)

	// Act
	modal.Open(tea.View{Content: "000"})

	updated, cmd := modal.Update(tea.KeyPressMsg{Code: 'Y'})

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
		WithForeground(func() tea.View { return tea.View{Content: "1"} }),
		WithKeyMap("enter", "N"), // Keep N default
		WithConfirmCmd(onConfirmHandler),
	)

	// Act
	modal.Open(tea.View{Content: "000"})

	updated, cmd := modal.Update(tea.KeyPressMsg{Text: "enter"})

	var msg tea.Msg
	if cmd != nil {
		msg = cmd()
	}

	_, ok := msg.(ConfirmMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(*Model).Opened(), true)
}

func Test_Modal_RespondsTo_DefaultCloseKey(t *testing.T) {
	// Arrange
	onCloseHandler := func() tea.Msg {
		return CloseMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() tea.View { return tea.View{Content: "1"} }),
		WithCancelCmd(onCloseHandler),
	)

	// Act
	modal.Open(tea.View{Content: "000"})
	updated, cmd := modal.Update(tea.KeyPressMsg{Code: 'N'})

	var msg tea.Msg
	if cmd != nil {
		msg = cmd()
	}
	_, ok := msg.(CloseMsg)

	// Assert
	assertEqual(t, ok, true)
	assertEqual(t, updated.(*Model).Opened(), false)
}

func Test_Modal_RespondsTo_CustomCloseKey(t *testing.T) {
	// Arrange
	onCloseHandler := func() tea.Msg {
		return CloseMsg{}
	}

	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() tea.View { return tea.View{Content: "1"} }),
		WithKeyMap("Y", "esc"), // Keep Y default
		WithCancelCmd(onCloseHandler),
	)

	// Act
	modal.Open(tea.View{Content: "000"})
	updated, cmd := modal.Update(tea.KeyPressMsg{Text: "esc"})

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
		WithForeground(func() tea.View { return tea.View{Content: "1"} }),
		WithKeyMap("enter", "esc"),
		WithCancelCmd(onCloseHandler),
	)

	// Act
	modal.Open(tea.View{Content: "000"})
	_, cmd := modal.Update(tea.KeyPressMsg{Text: "ctrl+q"})

	// Assert
	assertEqual(t, cmd == nil, true)
}

func Test_Modal_DimBackground_DimsBackground(t *testing.T) {
	// Arrange
	modal := New(
		WithPosition(lipgloss.Center, lipgloss.Center),
		WithForeground(func() tea.View { return tea.View{Content: "1"} }),
		WithDimmedBackground(true),
	)

	expectedBgStyle := lipgloss.NewStyle().Faint(true)

	// Act
	modal.Open(tea.View{Content: "000"})

	// Assert
	assertEqual(
		t,
		fmt.Sprintf(
			"%s%s%s",
			expectedBgStyle.Render("0"),
			"1",
			expectedBgStyle.Render("0")),
		modal.View().Content,
	)
}

func Test_Modal_Autocloses(t *testing.T) {
	// Arrange
	modal := New(
		WithForeground(func() tea.View { return tea.View{Content: "1"} }),
		WithOpenCmd(func() tea.Msg {
			return AutocloseMsg{}
		}),
	)

	// Act
	cmd := modal.Open(tea.View{Content: "000"})
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
