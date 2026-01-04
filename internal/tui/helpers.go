package tui

import "strings"

// renderSpinnerAction renders a spinner with an action message
// Format: [spinner] [action] [name (optional)]...
func renderSpinnerAction(spinnerView, action, name string) string {
	b := renderSimpleHeader("")
	b.WriteString(SpinnerStyle.Render(spinnerView))
	b.WriteString(" ")
	b.WriteString(action)
	if name != "" {
		b.WriteString(" ")
		b.WriteString(SuccessStyle.Render(name))
	}
	b.WriteString("...")
	return b.String()
}

// renderSpinnerWithHint renders a spinner action with an additional hint line
// Format: [spinner] [action] [name (optional)]...
//
//	[hint]
func renderSpinnerWithHint(spinnerView, action, name, hint string) string {
	var b strings.Builder
	b.WriteString(renderSpinnerAction(spinnerView, action, name))
	if hint != "" {
		b.WriteString("\n\n")
		b.WriteString(DimmedStyle.Render(hint))
	}
	return b.String()
}
