package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/fantasy"
	"charm.land/fantasy/providers/openai"
	"github.com/charmbracelet/fang"
	"github.com/kujtimiihoxha/recap/internal/agent"
	"github.com/kujtimiihoxha/recap/internal/render"
	"github.com/kujtimiihoxha/recap/internal/sandbox"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "recap",
		Short: "Analyse meeting transcripts and ask questions about them",
	}

	root.PersistentFlags().String("model", "", "model name (default: gpt-5.4, env: MODEL)")
	root.PersistentFlags().Int64("context-limit", agent.DefaultContextLimit, "token budget for the context window; summarization triggers at 90% (0 to disable)")

	analyseCmd := &cobra.Command{
		Use:   "analyse <path>",
		Short: "Analyse meeting transcripts at the given path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docs, err := loadDocuments(args[0])
			if err != nil {
				return err
			}
			model, err := initModel(cmd)
			if err != nil {
				return err
			}
			contextLimit, _ := cmd.Flags().GetInt64("context-limit")
			return runAnalyse(cmd.Context(), model, docs, contextLimit)
		},
	}

	askCmd := &cobra.Command{
		Use:   "ask <path> <question>",
		Short: "Ask a question about the content of the transcripts at the given path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			docs, err := loadDocuments(args[0])
			if err != nil {
				return err
			}
			model, err := initModel(cmd)
			if err != nil {
				return err
			}
			contextLimit, _ := cmd.Flags().GetInt64("context-limit")
			return runAsk(cmd.Context(), model, docs, args[1], contextLimit)
		},
	}

	root.AddCommand(analyseCmd, askCmd)

	if err := fang.Execute(context.Background(), root); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initModel(cmd *cobra.Command) (fantasy.LanguageModel, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	provider, err := openai.New(openai.WithAPIKey(apiKey), openai.WithUseResponsesAPI())
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}

	modelName, _ := cmd.Flags().GetString("model")
	if modelName == "" {
		modelName = os.Getenv("MODEL")
	}
	if modelName == "" {
		modelName = "gpt-5.4"
	}

	model, err := provider.LanguageModel(context.Background(), modelName)
	if err != nil {
		return nil, fmt.Errorf("create model: %w", err)
	}
	return model, nil
}

func loadDocuments(path string) ([]sandbox.Document, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("read dir %s: %w", path, err)
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			if !isTextFile(e.Name()) {
				continue
			}
			files = append(files, filepath.Join(path, e.Name()))
		}
	} else {
		if !isTextFile(filepath.Base(path)) {
			return nil, fmt.Errorf("unsupported file type: %s (only .txt and .md files are supported)", path)
		}
		files = []string{path}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found at %s", path)
	}

	docs := make([]sandbox.Document, 0, len(files))
	for i, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f, err)
		}
		docs = append(docs, sandbox.Document{
			ID:      fmt.Sprintf("doc-%d", i+1),
			Name:    filepath.Base(f),
			Content: string(content),
		})
	}

	return docs, nil
}

func isTextFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".txt" || ext == ".md"
}

func newAgent(model fantasy.LanguageModel, docs []sandbox.Document, contextLimit int64) (*agent.Agent, *sandbox.Sandbox, error) {
	maxQueryLen := agent.MaxLLMQueryContextLen(contextLimit)
	sb, err := sandbox.New(
		sandbox.WithDocuments(docs),
		sandbox.WithLLM(agent.NewInnerLLMFunc(model, maxQueryLen)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create sandbox: %w", err)
	}
	return agent.New(model, sb, contextLimit), sb, nil
}

func streamCallbacks(p *tea.Program) agent.StreamCallbacks {
	return agent.StreamCallbacks{
		OnReasoning: func(text string) {
			p.Send(render.ReasoningMsg(text))
		},
		OnToolCall: func(name, input string) {
			p.Send(render.ToolCallMsg{Name: name, Input: input})
		},
		OnToolResult: func(name, output string) {
			p.Send(render.ToolResultMsg{Name: name, Output: output})
		},
		OnTextDelta: func(text string) {
			p.Send(render.TextDeltaMsg(text))
		},
		OnAnalysis: func(a agent.AnalysisResult) {
			p.Send(render.AnalysisMsg{Result: a})
		},
		OnSummarizing: func() {
			p.Send(render.SummarizingMsg{})
		},
		OnSummary: func(summary string) {
			p.Send(render.SummaryMsg{Text: summary})
		},
		OnDone: func() {
			p.Send(render.DoneMsg{})
		},
	}
}

func runAnalyse(ctx context.Context, model fantasy.LanguageModel, docs []sandbox.Document, contextLimit int64) error {
	ag, sb, err := newAgent(model, docs, contextLimit)
	if err != nil {
		return err
	}
	defer func() { _ = sb.Close() }()

	p := tea.NewProgram(render.New())

	go func() {
		_, err := ag.RunAnalysis(ctx, nil, streamCallbacks(p))
		if err != nil {
			p.Send(render.ErrMsg{Err: err})
		}
	}()

	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	rm, ok := m.(render.Model)
	if ok {
		if rmErr := rm.Err(); rmErr != nil {
			return fmt.Errorf("analysis failed: %w", rmErr)
		}
		fmt.Println(rm.Output())
	}
	return nil
}

func runAsk(ctx context.Context, model fantasy.LanguageModel, docs []sandbox.Document, question string, contextLimit int64) error {
	ag, sb, err := newAgent(model, docs, contextLimit)
	if err != nil {
		return err
	}
	defer func() { _ = sb.Close() }()

	p := tea.NewProgram(render.New())

	go func() {
		_, err := ag.RunChat(ctx, nil, question, streamCallbacks(p))
		if err != nil {
			p.Send(render.ErrMsg{Err: err})
			return
		}
		p.Send(render.DoneMsg{})
	}()

	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}
	rm, ok := m.(render.Model)
	if ok {
		if rmErr := rm.Err(); rmErr != nil {
			return fmt.Errorf("chat failed: %w", rmErr)
		}
		fmt.Println(rm.Output())
	}
	return nil
}
