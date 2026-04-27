package assetmai

import (
	"context"
	"github.com/openai/openai-go"
	mai "github.com/manageai-inet/Eino-ManageAI-Extension/components/core/manageai"
)

type LLM struct {
	client *mai.Client
}

func NewLLM(client *mai.Client) *LLM {
	return &LLM{client: client}
}

func NewLLMFromEnv() *LLM {
	client, err := mai.GetDefaultClient()
	if err != nil {
		panic(err.Error())
	}
	return &LLM{client: client}
}

func (m *LLM) Generate(ctx context.Context, messages []openai.ChatCompletionMessage, tools []openai.ChatCompletionToolParam, toolChoice *openai.ChatCompletionToolChoiceOptionUnionParam) (*openai.ChatCompletionMessage, error) {
	maiMessages := []any{}
	for _, m := range messages {
		switch m.Role {
		case "user":
			maiMessages = append(maiMessages, map[string]any{
				"role": "user",
				"content": m.Content,
			})
		case "assistant":
			maiMessages = append(maiMessages, map[string]any{
				"role": "assistant",
				"content": m.Content,
			})
		case "system":
			maiMessages = append(maiMessages, map[string]any{
				"role": "system",
				"content": m.Content,
			})
		case "tool":
			maiMessages = append(maiMessages, map[string]any{
				"role": "tool",
				"content": m.Content,
			})
		}
	}
	opts := mai.ChatCompletionOptions{}
	maiTools := []mai.Tools{}
	for _, t := range tools {
		maiTools = append(maiTools, mai.Tools{
			Type: "function",
			Function: mai.ToolDefinition{
				Name: t.Function.Name,
				Description: t.Function.Description.Value,
				Parameters: t.Function.Parameters,
				Strict: t.Function.Strict.Value,
			},
		})
	}		
	if toolChoice != nil {
		if toolChoice.OfChatCompletionNamedToolChoice != nil {
			opts.WithTools(maiTools, &toolChoice.OfChatCompletionNamedToolChoice.Function.Name)
		} else if toolChoice.OfAuto.Value != "" {
			opts.WithTools(maiTools, &toolChoice.OfAuto.Value)
		} else {
			auto := "auto"
			opts.WithTools(maiTools, &auto)
		}
	} else {
		opts.WithTools(maiTools, nil)
	}
	response, err := m.client.ChatCompletion(maiMessages, &opts)
	if err != nil {
		return nil, err
	}
	
	toolCalls := []openai.ChatCompletionMessageToolCall{}
	if response.Choices[0].Message.ToolsCalls != nil && len(*response.Choices[0].Message.ToolsCalls) > 0 {
		for _, toolCall := range *response.Choices[0].Message.ToolsCalls {
			toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCall{
				Type: "function",
				ID: *toolCall.Id,
				Function: openai.ChatCompletionMessageToolCallFunction{
					Name: *toolCall.Function.Name,
					Arguments: *toolCall.Function.Arguments,
				},
			})
		}
	}
		
	openaiMessage := openai.ChatCompletionMessage{
		Role: "assistant",
		Content: *response.Choices[0].Message.Content,
		ToolCalls: toolCalls,
	}
	return &openaiMessage, nil
}