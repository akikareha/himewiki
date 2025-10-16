package filter

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"

	"github.com/akikareha/himewiki/internal/config"
)

func gnomeWithChatGPT(cfg *config.Config, title string, content string) (string, error) {
	apiKey := cfg.Gnome.Key
	if apiKey == "" {
		return "", fmt.Errorf("Gnome filter key (OpenAI API key) not set")
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
				openai.SystemMessage(cfg.Gnome.System + "\n" + cfg.Filter.Common),
				openai.UserMessage(cfg.Gnome.Prompt + message),
			},
			Temperature: openai.Float(cfg.Gnome.Temperature),
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	answer := resp.Choices[0].Message.Content

	return answer, nil
}
