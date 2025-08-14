package llm

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	llmApi "github.com/ollama/ollama/api"
)

func LLMClient(ctx context.Context) (*llmApi.Client, error) {
	llmClient, err := llmApi.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %v", err)
	}

	return llmClient, nil
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

func model(llmClient *llmApi.Client, ctx context.Context) (string, error) {
	supportedModels := []string{"deepseek-r1", "llama3.2", "mistral", "phi4"}

	models, err := llmClient.List(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list models: %v", err)
	}
	fmt.Println("Available models:")
	for _, model := range models.Models {
		fmt.Println(model.Name)
	}

	fmt.Println("Select a model to use:")
	fmt.Println("0. Cancel")
	for i, model := range supportedModels {
		fmt.Printf("%d. %s\n", i+1, model)
	}

	consoleReader := bufio.NewReader(os.Stdin)
	input, _ := consoleReader.ReadByte()
	modelIndex := int(input - '0')

	// ESC = 27 and Ctrl-C = 3
	if input == 27 || input == 3 || modelIndex == 0 {
		return "", fmt.Errorf("setup cancelled")
	}

	if modelIndex > len(supportedModels) {
		return "", fmt.Errorf("invalid model index")
	}
	selectedModel := supportedModels[modelIndex-1]

	modelInstalled := false
	for _, m := range models.Models {
		if strings.Contains(m.Name, selectedModel) {
			modelInstalled = true
			break
		}
	}

	if !modelInstalled {
		fmt.Println("Model not installed. Pulling:", selectedModel)
		err = pullModel(llmClient, selectedModel, ctx)
		if err != nil {
			return selectedModel, err
		}
	}

	return selectedModel, nil
}

func Predict(llmClient *llmApi.Client, ctx context.Context, text string) (string, error) {
	systemMessage := `
Predict the continuation of the given text up to max 1-2 words including the last word fragment if it is incomplete.
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
		Model:    "llama3.2",
		Messages: messages,
		Stream:   new(bool),
	}

	response := ""
	respFunc := func(resp llmApi.ChatResponse) error {
		response = resp.Message.Content
		return nil
	}

	err := llmClient.Chat(ctx, req, respFunc)
	if err != nil {
		return "", fmt.Errorf("failed to generate description: %v", err)
	}

	return response, nil
}
