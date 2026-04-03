package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openai"
	"github.com/kujtimiihoxha/recap/internal/sandbox"
)

const DefaultContextLimit int64 = 250_000

type StreamCallbacks struct {
	OnReasoning    func(text string)
	OnToolCall     func(name, input string)
	OnToolResult   func(name, output string)
	OnTextDelta   func(text string)
	OnAnalysis    func(AnalysisResult)
	OnSummarizing func()
	OnSummary      func(summary string)
	OnDone         func()
}

type AnalysisRunResult struct {
	Analysis *AnalysisResult
}

type ChatRunResult struct {
	Response string
}

type Agent struct {
	model        fantasy.LanguageModel
	sb           *sandbox.Sandbox
	contextLimit int64
}

func New(model fantasy.LanguageModel, sb *sandbox.Sandbox, contextLimit int64) *Agent {
	return &Agent{model: model, sb: sb, contextLimit: contextLimit}
}

func (a *Agent) RunAnalysis(ctx context.Context, messages []fantasy.Message, callbacks StreamCallbacks) (*AnalysisRunResult, error) {
	prompt := "Analyze the uploaded documents. Extract decisions, ownership, deadlines, and contradictions. Then call submit_analysis with the structured result."

	for {
		stopConditions := []fantasy.StopCondition{
			fantasy.HasToolCall("submit_analysis"),
		}
		if a.contextLimit > 0 {
			stopConditions = append(stopConditions, contextLimitReached(a.contextLimit))
		}

		maxQueryLen := MaxLLMQueryContextLen(a.contextLimit)
		ag := fantasy.NewAgent(
			a.model,
			fantasy.WithSystemPrompt(analyzeSystemPrompt(maxQueryLen)),
			fantasy.WithTools(NewRunCodeTool(a.sb, maxQueryLen), NewSubmitAnalysisTool()),
		)

		result, err := ag.Stream(ctx, newStreamCall(messages, prompt, callbacks, stopConditions))
		if err != nil {
			return nil, fmt.Errorf("analysis stream: %w", err)
		}

		if !needsSummarization(result, a.contextLimit) {
			analysis, err := extractAnalysis(result)
			if err != nil {
				return nil, err
			}
			if callbacks.OnAnalysis != nil {
				callbacks.OnAnalysis(*analysis)
			}
			if callbacks.OnDone != nil {
				callbacks.OnDone()
			}
			return &AnalysisRunResult{Analysis: analysis}, nil
		}

		// Summarize and continue.
		summary, err := a.summarizeAndNotify(ctx, messages, prompt, result, callbacks)
		if err != nil {
			return nil, err
		}
		messages = []fantasy.Message{fantasy.NewUserMessage(summary)}
		prompt = "The previous session was interrupted because the context got too long. " +
			"The initial request was to analyze the uploaded documents. " +
			"Continue from where you left off. When done, call submit_analysis with the structured result."
	}
}

func (a *Agent) RunChat(ctx context.Context, messages []fantasy.Message, question string, callbacks StreamCallbacks) (*ChatRunResult, error) {
	prompt := question

	for {
		var stopConditions []fantasy.StopCondition
		if a.contextLimit > 0 {
			stopConditions = append(stopConditions, contextLimitReached(a.contextLimit))
		}

		maxQueryLen := MaxLLMQueryContextLen(a.contextLimit)
		ag := fantasy.NewAgent(
			a.model,
			fantasy.WithSystemPrompt(chatSystemPrompt(maxQueryLen)),
			fantasy.WithTools(NewRunCodeTool(a.sb, maxQueryLen)),
		)

		result, err := ag.Stream(ctx, newStreamCall(messages, prompt, callbacks, stopConditions))
		if err != nil {
			return nil, fmt.Errorf("chat stream: %w", err)
		}

		if !needsSummarization(result, a.contextLimit) {
			if callbacks.OnDone != nil {
				callbacks.OnDone()
			}
			return &ChatRunResult{Response: result.Response.Content.Text()}, nil
		}

		// Summarize and continue.
		summary, err := a.summarizeAndNotify(ctx, messages, prompt, result, callbacks)
		if err != nil {
			return nil, err
		}
		messages = []fantasy.Message{fantasy.NewUserMessage(summary)}
		prompt = fmt.Sprintf(
			"The previous session was interrupted because the context got too long. "+
				"The user's original question was: %q. Continue from where you left off.", question)
	}
}

func contextLimitReached(limit int64) fantasy.StopCondition {
	threshold := int64(float64(limit) * 0.9)
	return func(steps []fantasy.StepResult) bool {
		if len(steps) == 0 {
			return false
		}
		return steps[len(steps)-1].Usage.InputTokens >= threshold
	}
}

func needsSummarization(result *fantasy.AgentResult, limit int64) bool {
	if limit <= 0 || len(result.Steps) == 0 {
		return false
	}
	threshold := int64(float64(limit) * 0.9)
	return result.Steps[len(result.Steps)-1].Usage.InputTokens >= threshold
}

func collectMessages(initial []fantasy.Message, prompt string, result *fantasy.AgentResult) []fantasy.Message {
	msgs := make([]fantasy.Message, 0, len(initial)+1+len(result.Steps)*2)
	msgs = append(msgs, initial...)
	if prompt != "" {
		msgs = append(msgs, fantasy.NewUserMessage(prompt))
	}
	for _, step := range result.Steps {
		msgs = append(msgs, step.Messages...)
	}
	return msgs
}

func (a *Agent) summarize(ctx context.Context, messages []fantasy.Message) (string, error) {
	ag := fantasy.NewAgent(a.model,
		fantasy.WithSystemPrompt(summarySystemPrompt),
	)
	result, err := ag.Stream(ctx, fantasy.AgentStreamCall{
		Messages: messages,
		Prompt:   "Provide a detailed summary of the conversation above.",
	})
	if err != nil {
		return "", fmt.Errorf("summarize: %w", err)
	}
	return result.Response.Content.Text(), nil
}

func (a *Agent) summarizeAndNotify(ctx context.Context, messages []fantasy.Message, prompt string, result *fantasy.AgentResult, cb StreamCallbacks) (string, error) {
	if cb.OnSummarizing != nil {
		cb.OnSummarizing()
	}

	allMessages := collectMessages(messages, prompt, result)
	summary, err := a.summarize(ctx, allMessages)
	if err != nil {
		return "", err
	}

	if cb.OnSummary != nil {
		cb.OnSummary(summary)
	}
	return summary, nil
}

func newStreamCall(messages []fantasy.Message, prompt string, cb StreamCallbacks, stopWhen []fantasy.StopCondition) fantasy.AgentStreamCall {
	opts := fantasy.ProviderOptions{
		openai.Name: &openai.ResponsesProviderOptions{
			Include: []openai.IncludeType{
				openai.IncludeReasoningEncryptedContent,
			},
			ReasoningEffort:  openai.ReasoningEffortOption(openai.ReasoningEffortMedium),
			ReasoningSummary: fantasy.Opt("auto"),
		},
	}
	return fantasy.AgentStreamCall{
		Messages:        messages,
		Prompt:          prompt,
		StopWhen:        stopWhen,
		ProviderOptions: opts,
		OnReasoningDelta: func(_ string, text string) error {
			if cb.OnReasoning != nil {
				cb.OnReasoning(text)
			}
			return nil
		},
		OnToolCall: func(tc fantasy.ToolCallContent) error {
			if cb.OnToolCall != nil {
				cb.OnToolCall(tc.ToolName, tc.Input)
			}
			return nil
		},
		OnToolResult: func(tr fantasy.ToolResultContent) error {
			if cb.OnToolResult != nil {
				if text, ok := tr.Result.(fantasy.ToolResultOutputContentText); ok {
					cb.OnToolResult(tr.ToolName, text.Text)
				}
			}
			return nil
		},
		OnTextDelta: func(_ string, text string) error {
			if cb.OnTextDelta != nil {
				cb.OnTextDelta(text)
			}
			return nil
		},
	}
}

func extractAnalysis(result *fantasy.AgentResult) (*AnalysisResult, error) {
	for _, step := range result.Steps {
		for _, tc := range step.Content.ToolCalls() {
			if tc.ToolName == "submit_analysis" {
				var input SubmitAnalysisInput
				if err := json.Unmarshal([]byte(tc.Input), &input); err != nil {
					return nil, fmt.Errorf("unmarshal analysis: %w", err)
				}
				return &input.Analysis, nil
			}
		}
	}
	return nil, fmt.Errorf("submit_analysis tool call not found in agent result")
}
