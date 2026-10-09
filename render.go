package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const barWidth = 28

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F5F5F5"))
	yearStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8A8A8A"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8A8A8A"))
	labelStyle  = lipgloss.NewStyle().Width(16).Foreground(lipgloss.Color("#C0C0C0"))
	timeStyle   = lipgloss.NewStyle().Bold(true).Width(8).Align(lipgloss.Right)
	linkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#5C9DFF")).Underline(true)
	cardStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#555555")).Padding(0, 2)
	emptyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))
	errorStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C")).Italic(true)
)

// timeRow describes one playtime category shown as a bar.
type timeRow struct {
	label   string
	seconds int
	polls   int
	color   lipgloss.Color
}

func renderResults(w io.Writer, res *SearchResult, query string) {
	if len(res.Data) == 0 {
		fmt.Fprintln(w, emptyStyle.Render(fmt.Sprintf("No results for %q", query)))
		return
	}

	for _, g := range res.Data {
		fmt.Fprintln(w, cardStyle.Render(renderGame(g)))
	}

	footer := fmt.Sprintf("%d result(s) · page %d/%d", res.Count, res.PageCurrent, res.PageTotal)
	fmt.Fprintln(w, footerStyle.Render(footer))
}

func renderGame(g Game) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(g.Name))
	if g.ReleaseYear > 0 {
		b.WriteString(" " + yearStyle.Render(fmt.Sprintf("(%d)", g.ReleaseYear)))
	}
	b.WriteString("\n")

	if g.Platforms != "" {
		b.WriteString(dimStyle.Render(g.Platforms) + "\n")
	}
	b.WriteString("\n")

	rows := timeRows(g)
	if len(rows) == 0 {
		b.WriteString(dimStyle.Render("No playtime data yet") + "\n")
	}

	maxSeconds := 0
	for _, r := range rows {
		maxSeconds = max(maxSeconds, r.seconds)
	}
	for _, r := range rows {
		b.WriteString(renderRow(r, maxSeconds) + "\n")
	}

	b.WriteString("\n" + renderStats(g) + "\n")
	b.WriteString(linkStyle.Render(g.URL()))
	return b.String()
}

func timeRows(g Game) []timeRow {
	candidates := []timeRow{
		{"Main Story", g.Main, g.MainCount, lipgloss.Color("#50FA7B")},
		{"Main + Extras", g.MainExtra, g.ExtraCount, lipgloss.Color("#8BE9FD")},
		{"Completionist", g.Completion, g.CompCount, lipgloss.Color("#BD93F9")},
		{"All Styles", g.AllStyles, g.AllCount, lipgloss.Color("#FFB86C")},
		{"Co-Op", g.CoOp, 0, lipgloss.Color("#FF79C6")},
		{"Versus", g.Multi, 0, lipgloss.Color("#FF5555")},
	}

	var rows []timeRow
	for _, r := range candidates {
		if r.seconds > 0 {
			rows = append(rows, r)
		}
	}
	return rows
}

func renderRow(r timeRow, maxSeconds int) string {
	filled := 0
	if maxSeconds > 0 {
		filled = max(1, r.seconds*barWidth/maxSeconds)
	}
	style := lipgloss.NewStyle().Foreground(r.color)
	bar := style.Render(strings.Repeat("█", filled)) + dimStyle.Render(strings.Repeat("░", barWidth-filled))

	line := labelStyle.Render(r.label) + bar + " " + timeStyle.Foreground(r.color).Render(formatDuration(r.seconds))
	if r.polls > 0 {
		line += " " + dimStyle.Render(fmt.Sprintf("(%s)", compactCount(r.polls)))
	}
	return line
}

func renderStats(g Game) string {
	var parts []string
	if g.ReviewScore > 0 {
		parts = append(parts, fmt.Sprintf("★ %d", g.ReviewScore))
	}
	parts = append(parts,
		fmt.Sprintf("%s playing", compactCount(g.Playing)),
		fmt.Sprintf("%s backlogs", compactCount(g.Backlog)),
		fmt.Sprintf("%s beaten", compactCount(g.Completed)),
	)
	return dimStyle.Render(strings.Join(parts, "  ·  "))
}

// formatDuration renders seconds as "Xh Ym", rounding to the nearest minute.
func formatDuration(seconds int) string {
	minutes := (seconds + 30) / 60
	hours, rem := minutes/60, minutes%60
	switch {
	case hours == 0:
		return fmt.Sprintf("%dm", rem)
	case rem == 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dh %dm", hours, rem)
	}
}

func compactCount(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
