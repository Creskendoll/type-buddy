package llm

import (
	"context"
	"fmt"

	llmApi "github.com/ollama/ollama/api"
)

type LLMClient struct {
	model  string
	client *llmApi.Client
	ctx    context.Context
}

func New(ctx context.Context, model string) (*LLMClient, error) {
	client, err := llmApi.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %v", err)
	}
	llm := LLMClient{
		client: client,
		ctx:    ctx,
	}

	if model == "" {
		llm.model = "llama3.2"
	} else {
		llm.model = model
	}

	return &llm, nil
}

func pullModel(llmClient *llmApi.Client, model string, ctx context.Context) error {
	req := &llmApi.PullRequest{
		Model: model,
		Name:  model,
	}

	gb := float32(1024.0 * 1024.0 * 1024.0)
	progressFunc := func(resp llmApi.ProgressResponse) error {
		if resp.Total > 0 {
			total := float32(resp.Total)
			completed := float32(resp.Completed)
			fmt.Print("\033[G\033[K")
			fmt.Printf("Downloading %.2f GB / %.2f GB (%.2f%%)\n", completed/gb, total/gb, completed/total*100.0)
			fmt.Print("\033[A")
		}
		return nil
	}

	err := llmClient.Pull(ctx, req, progressFunc)
	if err != nil {
		return fmt.Errorf("failed to pull model: %v", err)
	}

	return nil
}

func (llm LLMClient) ListModels() (*llmApi.ListResponse, error) {
	models, err := llm.client.List(llm.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %v", err)
	}
	return models, nil
}

func (llm LLMClient) Predict(text string) (string, error) {
	systemMessage := `
Predict the continuation of the given text up to max 1-2 words. Complete the last partial word if it is incomplete.
Suggest up to max 4 words if you are very confident.
Respond with {KO} if you cannot predict or the text does not make sense.
Include only the predicted text and nothing more.
`

	messages := []llmApi.Message{
		{
			Role:    "system",
			Content: systemMessage,
		},
		{
			Role:    "user",
			Content: text,
		},
	}

	req := &llmApi.ChatRequest{
		Model:    llm.model,
		Messages: messages,
		Stream:   new(bool),
	}

	response := ""
	respFunc := func(resp llmApi.ChatResponse) error {
		response = resp.Message.Content
		return nil
	}

	err := llm.client.Chat(llm.ctx, req, respFunc)
	if err != nil {
		return "{KO}", fmt.Errorf("failed to generate description: %v", err)
	}

	return response, nil
}

func (llm LLMClient) Correct(text string) (string, error) {
	systemMessage := `
	Correct the given text if it is incorrect.
	Do not be too aggressive with the corrections. Do not correct abbreviations.
	Respond with {KO} if the text is correct or is good enough.
	Include only the corrected text and nothing more.
	`

	messages := []llmApi.Message{
		{
			Role:    "system",
			Content: systemMessage,
		},
		{
			Role:    "user",
			Content: text,
		},
	}

	req := &llmApi.ChatRequest{
		Model:    llm.model,
		Messages: messages,
		Stream:   new(bool),
	}

	response := ""
	respFunc := func(resp llmApi.ChatResponse) error {
		response = resp.Message.Content
		return nil
	}

	err := llm.client.Chat(llm.ctx, req, respFunc)
	if err != nil {
		return "{KO}", fmt.Errorf("failed to generate description: %v", err)
	}

	return response, nil
}
