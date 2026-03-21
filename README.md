# bubbletea-modal

A modal component that can be used to create dialogs and popups in your terminal applications. Targets v1 of Lipgloss and Bubble Tea (before the introduction of the compositing API).

## Features
- **Satisfies `tea.Model`**: familiar API for initialisation and updates.
- **ANSI-aware**: preserves colour and text style passed in from Lipgloss or other style libraries.
- **Positionable**: supports `lipgloss.Position` positioning on both axes.
- **Stackable**: `modal.Model` accepts the View() of other `modal.Model` as a background.
- **OSC-8 support**: "non-typical" sequences such as OSC-8 URLs are supported.

## Usage
```go
func dialogContent() string {
	message := lipgloss.NewStyle().Margin(2).PaddingBottom(2).Render("Are you sure you want to quit?")
	confirm := lipgloss.NewStyle().MarginRight(2).Render("Confirm Y")
	cancel := lipgloss.NewStyle().Render("Cancel N")

	joined := lipgloss.JoinVertical(
		lipgloss.Center,
		message,
		lipgloss.JoinHorizontal(lipgloss.Center, confirm, cancel),
	)

	return lipgloss.NewStyle.Border(lipgloss.RoundedBorder()).Render(joined)
}

// Create model (e.g. during parent initialisation), as you would with a text input or viewport
modal := modal.New(
	lipgloss.Center, // Horizontal and vertical positioning.
	lipgloss.Center, 
	dialogContent, // Function handle to dialog content.
	func() tea.Msg { return modal.ConfirmMsg },
	nil,  // Extra behaviour if the user decides not to quit!
)

// ... then respond to a "Y" keypress
m.modal.Open(parentModel.View())
```

## Usage Notes
- `modal` operates on a snapshot basis: its `Open()` method expects a string background (usually the `View()` of the parent component).
- `onConfirm()`/`onClose()` behaviour is injectable at creation time.
- The default `onConfirm()` behaviour is to return nil. The default `onClose()` behaviour is to close (stop displaying the modal) and return nil (this is often enough for most uses, but is left open to more complex behaviour).

## Roadmap
- v2 using Lipgloss Layers API.