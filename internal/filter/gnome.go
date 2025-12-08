package filter

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/akikareha/himewiki/internal/config"
	"github.com/akikareha/himewiki/internal/format"
)

func gnomeWithOpenAI(cfg *config.Config, title string, content string) (string, error) {
	apiKey := cfg.Gnome.Key
	if apiKey == "" {
		return "", fmt.Errorf("Gnome filter key (OpenAI API key) not set")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	message := "title: " + title + "\n\ncontent:\n" + content

	mode := format.Detect(cfg, content)
	var formatPrompt string
	if mode == "creole" {
		formatPrompt = cfg.Prompts.Creole
	} else if mode == "markdown" {
		formatPrompt = cfg.Prompts.Markdown
	} else {
		formatPrompt = cfg.Prompts.Nomark
	}

	resp, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model: openai.ChatModelGPT4o,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(cfg.Prompts.Gnome + "\n" + cfg.Prompts.Common + "\n" + formatPrompt),
				openai.UserMessage(message),
			},
			Temperature: openai.Float(cfg.Gnome.Temperature),
			TopP:        openai.Float(cfg.Gnome.TopP),
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
