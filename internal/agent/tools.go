package agent

import (
	"context"
	"fmt"

	"charm.land/fantasy"
	"github.com/kujtimiihoxha/recap/internal/sandbox"
)

const maxOutputLen = 25_000

type RunCodeInput struct {
	Code string `json:"code" description:"JavaScript code to execute in the sandbox"`
}

func NewRunCodeTool(sb *sandbox.Sandbox, maxQueryLen int) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"run_code",
		runCodeToolDescription(maxQueryLen),
		func(ctx context.Context, input RunCodeInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			output, err := sb.Eval(ctx, input.Code)
			if err != nil {
				msg := fmt.Sprintf("Error: %s", err)
				if output != "" {
					msg = fmt.Sprintf("stdout:\n%s\nError: %s", truncateOutput(output), err)
				}
				return fantasy.NewTextErrorResponse(msg), nil
			}
			if output == "" {
				return fantasy.NewTextResponse("(no output)"), nil
			}
			return fantasy.NewTextResponse(truncateOutput(output)), nil
		},
	)
}

func NewSubmitAnalysisTool() fantasy.AgentTool {
	return fantasy.NewAgentTool(
		"submit_analysis",
		"Submit the structured analysis of the uploaded documents. Call this once you have completed your analysis.",
		func(_ context.Context, _ SubmitAnalysisInput, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
			return fantasy.NewTextResponse("Analysis submitted."), nil
		},
	)
}

func truncateOutput(s string) string {
	if len(s) <= maxOutputLen {
		return s
	}
	// Keep head and tail so the model sees both boundaries.
	tailLen := maxOutputLen / 5 // 20% for tail
	headLen := maxOutputLen - tailLen
	return s[:headLen] + fmt.Sprintf("\n\n... (%d chars truncated) ...\n\n", len(s)-headLen-tailLen) + s[len(s)-tailLen:]
}
