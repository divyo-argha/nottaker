package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func visibleLen(s string) int {
	inEscape := false
	count := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}

func main() {
	width := 80
	usableWidth := width - 2
	parts := []string{
		"^N new", "^X/F4 close", "^O/F3 open", "^S/F2 save", "^T share",
		"^R receive", "^→/← switch", "Tab cycle", "^C quit",
	}
	for i := range parts {
		parts[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(parts[i])
	}
	
	saveStatus := "✓ saved 21:12:28"
	
	var lines []string
	currentLine := ""
	for i, part := range parts {
		partLen := visibleLen(part)
		if currentLine == "" {
			currentLine = part
		} else {
			if visibleLen(currentLine)+2+partLen > usableWidth {
				lines = append(lines, currentLine)
				currentLine = part
			} else {
				currentLine += "  " + part
			}
		}
		if i == len(parts)-1 {
			saveLen := visibleLen(saveStatus)
			if visibleLen(currentLine)+1+saveLen > usableWidth {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			gap := usableWidth - visibleLen(currentLine) - saveLen
			if gap < 0 {
				gap = 0
			}
			currentLine += strings.Repeat(" ", gap) + saveStatus
			lines = append(lines, currentLine)
		}
	}
	
	styleLegend := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("248")).
			Padding(0, 1)

	res := styleLegend.Width(width).Render(strings.Join(lines, "\n"))
	fmt.Println("Result:")
	fmt.Println(res)
}
