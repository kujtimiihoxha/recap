package agent

import (
	"context"
	"fmt"

	"charm.land/fantasy"
	"github.com/kujtimiihoxha/recap/internal/sandbox"
)

func NewInnerLLMFunc(model fantasy.LanguageModel, maxQueryLen int) sandbox.LLMFunc {
	return func(ctx context.Context, llmContext string, query string) (string, error) {
		if len(llmContext) > maxQueryLen {
			return "", fmt.Errorf(
				"llm_query context too large: %d chars (max %d). Split the content into smaller chunks",
				len(llmContext), maxQueryLen,
			)
		}

		prompt := fmt.Sprintf("<context>\n%s\n</context>\n\n%s", llmContext, query)

		ag := fantasy.NewAgent(
			model,
			fantasy.WithSystemPrompt(innerSystemPrompt),
		)
		result, err := ag.Generate(ctx, fantasy.AgentCall{Prompt: prompt})
		if err != nil {
			return "", err
		}
		return result.Response.Content.Text(), nil
	}
}
