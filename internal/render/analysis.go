package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kujtimiihoxha/recap/internal/agent"
)

func wrapStyle(w int) lipgloss.Style {
	return lipgloss.NewStyle().Width(w)
}

func renderAnalysis(a agent.AnalysisResult, width int) string {
	width = max(width, 80)
	// border (2) + border padding (2*2) = 6
	contentWidth := max(width-6, 40)

	var sections []string

	// Summary.
	if a.Summary != "" {
		body := wrapStyle(contentWidth).Render(a.Summary)
		sections = append(sections, renderSection("Summary", body))
	}

	// Documents.
	if len(a.Documents) > 0 {
		var rows []string
		for _, d := range a.Documents {
			header := fmt.Sprintf("  %s  %s", labelStyle.Render(d.ID), d.Name)
			summary := wrapStyle(contentWidth - 4).Render(d.Summary)
			// indent each line of the wrapped summary
			summary = indentBlock(summary, "  ")
			rows = append(rows, header+"\n"+dimStyle.Render(summary))
		}
		sections = append(sections, renderSection(
			fmt.Sprintf("Documents (%d)", len(a.Documents)),
			strings.Join(rows, "\n\n"),
		))
	}

	// Decisions.
	if len(a.Decisions) > 0 {
		var items []string
		for _, d := range a.Decisions {
			var lines []string

			// First line: badge + description on the same line.
			badge := ""
			if d.Status != nil {
				badge = renderStatusBadge(*d.Status) + "  "
			}
			// Wrap description accounting for the badge + indent.
			badgeWidth := lipgloss.Width(badge)
			firstIndent := "  " // 2-space indent
			contIndent := firstIndent + strings.Repeat(" ", badgeWidth)
			descWidth := contentWidth - len(contIndent)
			wrapped := wrapStyle(descWidth).Render(d.Description)
			descLines := strings.Split(wrapped, "\n")
			for i, l := range descLines {
				if i == 0 {
					lines = append(lines, firstIndent+badge+l)
				} else if l != "" {
					lines = append(lines, contIndent+l)
				}
			}

			// Citation or reasoning below, further indented.
			citIndent := contIndent + "  "
			citWidth := contentWidth - len(citIndent)
			if d.Source != nil {
				lines = append(lines, citIndent+renderCitation(*d.Source, citWidth))
			} else if d.Reasoning != "" {
				r := wrapStyle(citWidth).Render(d.Reasoning)
				lines = append(lines, indentBlock(citationStyle.Render(r), citIndent))
			}

			items = append(items, strings.Join(lines, "\n"))
		}
		sections = append(sections, renderSection(
			fmt.Sprintf("Decisions (%d)", len(a.Decisions)),
			strings.Join(items, "\n\n"),
		))
	}

	// Owners.
	if len(a.Owners) > 0 {
		var items []string
		for _, o := range a.Owners {
			line := fmt.Sprintf("  %s -> %s", labelStyle.Render(o.Owner), o.Item)
			line = wrapStyle(contentWidth).Render(line)
			if o.Source != nil {
				line += "\n" + indentBlock(renderCitation(*o.Source, contentWidth-6), "    ")
			}
			items = append(items, line)
		}
		sections = append(sections, renderSection(
			fmt.Sprintf("Ownership (%d)", len(a.Owners)),
			strings.Join(items, "\n"),
		))
	}

	// Deadlines.
	if len(a.Deadlines) > 0 {
		var items []string
		for _, d := range a.Deadlines {
			line := fmt.Sprintf("  %s -> %s", d.Item, labelStyle.Render(d.Date))
			line = wrapStyle(contentWidth).Render(line)
			if d.Source != nil {
				line += "\n" + indentBlock(renderCitation(*d.Source, contentWidth-6), "    ")
			}
			items = append(items, line)
		}
		sections = append(sections, renderSection(
			fmt.Sprintf("Deadlines (%d)", len(a.Deadlines)),
			strings.Join(items, "\n"),
		))
	}

	// Contradictions.
	if len(a.Contradictions) > 0 {
		var items []string
		for _, c := range a.Contradictions {
			desc := wrapStyle(contentWidth - 2).Render(warningStyle.Render(c.Description))
			desc = indentBlock(desc, "  ")
			var claims []string
			for _, cl := range c.Claims {
				claim := wrapStyle(contentWidth - 6).Render(cl.Statement)
				claim = indentBlock(claim, "      ")
				claim = "    - " + strings.TrimPrefix(claim, "      ")
				if cl.Source != nil {
					claim += "\n" + indentBlock(renderCitation(*cl.Source, contentWidth-8), "      ")
				}
				claims = append(claims, claim)
			}
			items = append(items, desc+"\n"+strings.Join(claims, "\n"))
		}
		sections = append(sections, renderSection(
			fmt.Sprintf("Contradictions (%d)", len(a.Contradictions)),
			strings.Join(items, "\n\n"),
		))
	}

	body := strings.Join(sections, "\n\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1, 2).
		Width(width).
		Render(body)

	return box
}

func renderSection(title, content string) string {
	header := headerStyle.Render(title)
	divider := dividerStyle.Render(strings.Repeat("─", lipgloss.Width(title)))
	return header + "\n" + divider + "\n" + content
}

func renderStatusBadge(status string) string {
	switch strings.ToLower(status) {
	case "decided":
		return resultHeaderStyle.Render(status)
	case "pending":
		return statusPending.Render(status)
	case "reversed":
		return statusReversed.Render(status)
	default:
		return lipgloss.NewStyle().Faint(true).Render(status)
	}
}

func renderCitation(c agent.Citation, maxWidth int) string {
	excerpt := c.Excerpt
	// Truncate very long excerpts but keep a reasonable amount.
	maxExcerpt := maxWidth * 2
	if maxExcerpt < 80 {
		maxExcerpt = 80
	}
	if len(excerpt) > maxExcerpt {
		excerpt = excerpt[:maxExcerpt] + "..."
	}
	text := fmt.Sprintf("%s: \"%s\"", c.DocumentName, excerpt)
	return citationStyle.Render(wrapStyle(maxWidth).Render(text))
}

func indentBlock(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "\n")
}
