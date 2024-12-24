package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
)

var (
	FALSE = false
	TRUE  = true
)

func main() {
	ctx := context.Background()

	var ollamaRawUrl string
	if ollamaRawUrl = os.Getenv("OLLAMA_HOST"); ollamaRawUrl == "" {
		ollamaRawUrl = "http://localhost:11434"
	}

	var toolsLLM string
	if toolsLLM = os.Getenv("TOOLS_LLM"); toolsLLM == "" {
		toolsLLM = "allenporter/xlam:1b"
		//toolsLLM = "qwen2.5:1.5b"
	}

	url, _ := url.Parse(ollamaRawUrl)

	client := api.NewClient(url, http.DefaultClient)

	// Define some tools
	helloTool := map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "hello",
			"description": "Say hello to a given person with his name",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The name of the person",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	addNumbersTool := map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "add_numbers",
			"description": "Add two numbers",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"number1": map[string]any{
						"type":        "number",
						"description": "The first number",
					},
					"number2": map[string]any{
						"type":        "number",
						"description": "The second number",
					},
				},
				"required": []string{"number1", "number2"},
			},
		},
	}

	tools := []any{helloTool, addNumbersTool}
	// transform tools to json
	//jsonTools, _ := json.MarshalIndent(tools, "", "  ")
	jsonTools, _ := json.Marshal(tools)

	//fmt.Println(string(jsonTools))
	
	var toolsList api.Tools
	jsonErr := json.Unmarshal(jsonTools, &toolsList)
	if jsonErr != nil {
		log.Fatalln("😡", jsonErr)
	}

	// Prompt construction
	messages := []api.Message{
		{Role: "user", Content: "Say hello to Bob"},
		{Role: "user", Content: "add 28 to 12"},
		{Role: "user", Content: "Say hello to Sarah"},
	}

	req := &api.ChatRequest{
		Model: toolsLLM,
		Messages: messages,
		Options: map[string]interface{}{
			"temperature":   0.0,
			"repeat_last_n": 2,
		},
		Tools:  toolsList,
		Stream: &FALSE,
		//Format: json.RawMessage(`"json"`),
	}

	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
		//fmt.Println("🖐️", resp.Message.ToolCalls)

		for _, toolCall := range resp.Message.ToolCalls {
			fmt.Println(toolCall.Function.Name, toolCall.Function.Arguments)
		}

		return nil
	})

	if err != nil {
		log.Fatalln("😡", err)
	}

	fmt.Println()
	fmt.Println("🚀", ollamaRawUrl, toolsLLM)

}
