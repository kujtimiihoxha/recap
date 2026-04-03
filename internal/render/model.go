package render

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/kujtimiihoxha/recap/internal/agent"
	"github.com/tidwall/pretty"
)

type (
	ReasoningMsg string
	ToolCallMsg  struct {
		Name  string
		Input string
	}
	ToolResultMsg struct {
		Name   string
		Output string
	}
	TextDeltaMsg    string
	SummarizingMsg  struct{}
	SummaryMsg      struct{ Text string }
	AnalysisMsg     struct{ Result agent.AnalysisResult }
	DoneMsg         struct{}
	ErrMsg          struct{ Err error }
)

type phase int

const (
	phaseIdle phase = iota
	phaseReasoning
	phaseToolCall
	phaseToolResult
	phaseText
	phaseSummarizing
	phaseAnalysis
	phaseDone
)

type Model struct {
	viewport viewport.Model
	spinner  spinner.Model
	glam     *glamour.TermRenderer
	width    int
	height   int

	reasonBuf strings.Builder
	textBuf   strings.Builder
	blocks    []string
	current   phase
	waitingTool string
	err         error
}

func New() Model {
	vp := viewport.New()
	s := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7"))),
	)
	glam, _ := glamour.NewTermRenderer(
		glamour.WithStyles(customStyle()),
		glamour.WithWordWrap(80),
	)
	return Model{
		viewport: vp,
		spinner:  s,
		glam:     glam,
		current:  phaseIdle,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		innerW := max(msg.Width-appPadX*2, 40)
		innerH := max(msg.Height-2, 4) // top + bottom padding
		m.viewport.SetWidth(innerW)
		m.viewport.SetHeight(innerH)
		glam, err := glamour.NewTermRenderer(
			glamour.WithStyles(customStyle()),
			glamour.WithWordWrap(max(innerW-2, 40)),
		)
		if err == nil {
			m.glam = glam
		}
		m.updateViewport()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.updateViewport()
		return m, cmd

	case ReasoningMsg:
		m.transitionTo(phaseReasoning)
		m.reasonBuf.WriteString(string(msg))
		m.updateViewport()
		return m, nil

	case ToolCallMsg:
		m.transitionTo(phaseToolCall)
		block := m.renderToolCall(msg.Name, msg.Input)
		m.blocks = append(m.blocks, block)
		m.waitingTool = msg.Name
		m.updateViewport()
		return m, nil

	case ToolResultMsg:
		m.transitionTo(phaseToolResult)
		m.waitingTool = ""
		block := m.renderToolResult(msg.Name, msg.Output)
		m.blocks = append(m.blocks, block)
		m.updateViewport()
		return m, nil

	case TextDeltaMsg:
		m.transitionTo(phaseText)
		m.textBuf.WriteString(string(msg))
		m.updateViewport()
		return m, nil

	case SummarizingMsg:
		m.transitionTo(phaseSummarizing)
		m.updateViewport()
		return m, nil

	case SummaryMsg:
		block := m.renderSummary(msg.Text)
		m.blocks = append(m.blocks, block)
		m.current = phaseIdle // back to normal operation for the next agent loop
		m.updateViewport()
		return m, nil

	case AnalysisMsg:
		m.transitionTo(phaseAnalysis)
		block := renderAnalysis(msg.Result, max(m.width-appPadX*2, 40))
		m.blocks = append(m.blocks, block)
		m.updateViewport()
		return m, nil

	case ErrMsg:
		m.err = msg.Err
		return m, tea.Quit

	case DoneMsg:
		m.finalizeCurrent()
		m.current = phaseDone
		m.updateViewport()
		return m, tea.Quit

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

const appPadX = 2 // horizontal padding on each side

func (m Model) View() tea.View {
	if m.err != nil {
		return tea.NewView(fmt.Sprintf("Error: %v\n", m.err))
	}
	content := lipgloss.NewStyle().
		Padding(1, appPadX).
		Render(m.viewport.View())
	return tea.NewView(content)
}

func (m Model) Err() error {
	return m.err
}

func (m Model) Output() string {
	parts := slices.Clone(m.blocks)

	switch m.current {
	case phaseReasoning:
		if m.reasonBuf.Len() > 0 {
			parts = append(parts, m.renderReasoning(m.reasonBuf.String()))
		}
	case phaseText:
		if m.textBuf.Len() > 0 {
			parts = append(parts, m.renderGlamour(m.textBuf.String()))
		}
	}

	return strings.Join(parts, "\n")
}

func (m *Model) transitionTo(next phase) {
	if m.current == next {
		return
	}
	m.finalizeCurrent()
	m.current = next
}

func (m *Model) finalizeCurrent() {
	switch m.current {
	case phaseReasoning:
		if m.reasonBuf.Len() > 0 {
			block := m.renderReasoning(m.reasonBuf.String())
			m.blocks = append(m.blocks, block)
			m.reasonBuf.Reset()
		}
	case phaseText:
		if m.textBuf.Len() > 0 {
			block := m.renderGlamour(m.textBuf.String())
			m.blocks = append(m.blocks, block)
			m.textBuf.Reset()
		}
	}
}

func (m *Model) updateViewport() {
	parts := slices.Clone(m.blocks)
	switch m.current {
	case phaseIdle:
		parts = append(parts, m.spinnerLine("Thinking"))
	case phaseReasoning:
		if m.reasonBuf.Len() > 0 {
			parts = append(parts, m.renderReasoning(m.reasonBuf.String()))
		} else {
			parts = append(parts, m.spinnerLine("Thinking"))
		}
	case phaseText:
		if m.textBuf.Len() > 0 {
			parts = append(parts, m.renderGlamour(m.textBuf.String()))
		}
	case phaseSummarizing:
		parts = append(parts, m.spinnerLine("Summarizing context"))
	case phaseToolCall:
		if m.waitingTool != "" {
			parts = append(parts, m.spinnerLine("Running "+m.waitingTool))
		}
	}

	content := strings.Join(parts, "\n")
	content = strings.TrimRightFunc(content, unicode.IsSpace)

	wasAtBottom := m.viewport.AtBottom() || m.viewport.TotalLineCount() == 0
	m.viewport.SetContent(content)
	if wasAtBottom {
		m.viewport.GotoBottom()
	}
}

func (m *Model) spinnerLine(label string) string {
	return m.spinner.View() + " " + dimStyle.Render(label+"...")
}

func (m *Model) renderReasoning(text string) string {
	header := dimStyle.Render("Thinking...")
	rendered := m.renderGlamour(text)
	rendered = dimStyle.Render(rendered)
	return header + "\n" + rendered
}

func (m *Model) renderToolCall(name, input string) string {
	header := toolHeaderStyle.Render(" " + name + " ")

	if name == "run_code" {
		code := extractCodeFromInput(input)
		if code != "" {
			code = beautifyJS(code)
			md := fmt.Sprintf("```javascript\n%s\n```", code)
			rendered := m.renderGlamour(md)
			return header + "\n" + rendered
		}
	}

	if json.Valid([]byte(input)) {
		formatted := string(pretty.Pretty([]byte(input)))
		md := fmt.Sprintf("```json\n%s\n```", strings.TrimSpace(formatted))
		rendered := m.renderGlamour(md)
		return header + "\n" + rendered
	}

	return header + "\n" + input
}

func (m *Model) renderToolResult(name, output string) string {
	header := resultHeaderStyle.Render(" Result ")

	display := output
	if len(display) > 2000 {
		display = display[:1500] + fmt.Sprintf("\n\n... (%d chars truncated) ...\n\n", len(display)-2000) + display[len(display)-500:]
	}

	if json.Valid([]byte(strings.TrimSpace(display))) {
		formatted := string(pretty.Pretty([]byte(strings.TrimSpace(display))))
		md := fmt.Sprintf("```json\n%s\n```", strings.TrimSpace(formatted))
		rendered := m.renderGlamour(md)
		return header + "\n" + rendered
	}

	return header + "\n" + dimStyle.Render(display)
}

func (m *Model) renderSummary(text string) string {
	header := summaryHeaderStyle.Render(" Context Summary ")
	rendered := m.renderGlamour(text)
	return header + "\n" + rendered
}

func (m *Model) renderGlamour(text string) string {
	rendered, err := m.glam.Render(text)
	if err != nil {
		return text
	}
	if m.width > 0 {
		rendered = lipgloss.NewStyle().MaxWidth(m.width).Render(rendered)
	}
	return strings.TrimRight(rendered, "\n")
}

func extractCodeFromInput(input string) string {
	var v struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal([]byte(input), &v); err != nil {
		return ""
	}
	return v.Code
}

