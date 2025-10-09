package filter

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"

	"github.com/akikareha/himewiki/internal/config"
)

func withChatGPT(cfg *config.Config, title string, content string) (string, error) {
	apiKey := cfg.Filter.Key
	if apiKey == "" {
		return "", fmt.Errorf("Filter key (OpenAI API key) not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	message := "title: " + title + "\n\ncontent:\n" + content

	resp, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model: openai.ChatModelGPT4o,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(cfg.Filter.System),
				openai.UserMessage(cfg.Filter.Prompt + message),
			},
			Temperature: openai.Float(cfg.Filter.Temperature),
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	answer := resp.Choices[0].Message.Content

	statusIndex := strings.Index(answer, "STATUS:")
	if statusIndex == -1 {
		return "", fmt.Errorf("no status in response")
	}
	answer = answer[statusIndex + 7:]
	statusEndIndex := strings.Index(answer, "\n")
	if statusEndIndex == -1 {
		return "", fmt.Errorf("no line end for status in response")
	}
	status := strings.TrimSpace(answer[:statusEndIndex])
	if status != "ok" {
		return "", fmt.Errorf("rejected by AI filter")
	}
	answer = answer[statusEndIndex:]

	contentIndex := strings.Index(answer, "CONTENT:")
	if contentIndex == -1 {
		return "", fmt.Errorf("no content in response")
	}
	answer = answer[contentIndex + 8:]
	contentEndIndex := strings.Index(answer, "\n")
	if contentEndIndex == -1 {
		return "", fmt.Errorf("no line end for content in response")
	}
	answer = answer[contentEndIndex:]

	return answer, nil
}
