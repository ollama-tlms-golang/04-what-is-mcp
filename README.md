# "Tool Calling" and Ollama

## Prerequisites

You should have read the previous article [Developing Generative AI Applications in Go with Ollama](https://k33g.hashnode.dev/developing-generative-ai-applications-in-go-with-ollama-and-tiny-models) to understand the basics of using Ollama and baby LLMs.

For this new article, we'll use a this LLM, **[qwen2.5:1.5b](https://ollama.com/library/qwen2.5:1.5b)**. So run the following command to download it:

```bash
ollama pull qwen2.5:1.5b
```

## Introduction

In July 2024, Ollama announced support for tool calling (also known as function calling) for LLMs that supported it. So, what is tool calling?

## Explanations

For models that support tool calling, the principle is to provide a list of tools to the model. For example, below I explain to the model that it has a `say_hello` function and an `add_numbers` function:

```json
 "tools": [
    {
      "function": {
        "description": "Say hello to a given person with his name",
        "name": "say_hello",
        "parameters": {
          "properties": {
            "name": {
              "description": "The name of the person",
              "type": "string"
            }
          },
          "required": [
            "name"
          ],
          "type": "object"
        }
      },
      "type": "function"
    },
    {
      "function": {
        "description": "Add two numbers",
        "name": "add_numbers",
        "parameters": {
          "properties": {
            "number1": {
              "description": "The first number",
              "type": "number"
            },
            "number2": {
              "description": "The second number",
              "type": "number"
            }
          },
          "required": [
            "number1",
            "number2"
          ],
          "type": "object"
        }
      },
      "type": "function"
    }
  ]
}
```

From this list, when I send a prompt to the model with messages like `"Say hello to Bob"`, `"add 28 to 12"`, the LLM will know how to detect the function (or tool) call pattern and provide me with the name of the function to call and extract the parameters to pass to it.

### For example

For example, for a prompt like this:

```json
  "messages": [
    {
      "role": "user",
      "content": "Say hello to Bob"
    },
    {
      "role": "user",
      "content": "add 28 to 12"
    },
    {
      "role": "user",
      "content": "Say hello to Sarah"
    }
  ]
```

The model will detect three function calls and provide me with the function calls to execute:

```json
"tool_calls": [
	{
	"function": {
		"name": "say_hello",
		"arguments": {
		"name": "Bob"
		}
	}
	},
	{
	"function": {
		"name": "add_numbers",
		"arguments": {
		"number1": 12,
		"number2": 28
		}
	}
	},
	{
	"function": {
		"name": "say_hello",
		"arguments": {
		"name": "Sarah"
		}
	}
	}
]
```

**Note**: It's important to note that the model doesn't know how to execute the functions, it just extracts them. It's up to you to implement these functions.

## Using Curl and Ollama

One of my favorite models that supports tool calling is the `qwen2.5:1.5b` model (theoretically its smaller sibling `qwen2.5:0.5b` also supports tool calling, but much less effectively). You can test it very easily with a simple `curl` command (and install `jq` to format the JSON, it will look nicer):

```bash
#!/bin/bash 
SERVICE_URL="http://localhost:11434"
read -r -d '' DATA <<- EOM
{
  "model": "qwen2.5:1.5b",
  "messages": [
    {
      "role": "user",
      "content": "Say hello to Bob"
    },
    {
      "role": "user",
      "content": "add 28 to 12"
    },
    {
      "role": "user",
      "content": "Say hello to Sarah"
    }
  ],
  "stream": false,
  "tools": [
    {
      "function": {
        "description": "Say hello to a given person with his name",
        "name": "say_hello",
        "parameters": {
          "properties": {
            "name": {
              "description": "The name of the person",
              "type": "string"
            }
          },
          "required": [
            "name"
          ],
          "type": "object"
        }
      },
      "type": "function"
    },
    {
      "function": {
        "description": "Add two numbers",
        "name": "add_numbers",
        "parameters": {
          "properties": {
            "number1": {
              "description": "The first number",
              "type": "number"
            },
            "number2": {
              "description": "The second number",
              "type": "number"
            }
          },
          "required": [
            "number1",
            "number2"
          ],
          "type": "object"
        }
      },
      "type": "function"
    }
  ]
}
EOM

curl --no-buffer ${SERVICE_URL}/api/chat \
    -H "Content-Type: application/json" \
    -d "${DATA}" | jq '.'
``` 

And you'll get a result similar to this:

```json
{
  "model": "qwen2.5:1.5b",
  "created_at": "2024-12-23T09:01:42.086686Z",
  "message": {
    "role": "assistant",
    "content": "",
    "tool_calls": [
      {
        "function": {
          "name": "say_hello",
          "arguments": {
            "name": "Bob"
          }
        }
      },
      {
        "function": {
          "name": "add_numbers",
          "arguments": {
            "number1": 28,
            "number2": 12
          }
        }
      },
      {
        "function": {
          "name": "say_hello",
          "arguments": {
            "name": "Sarah"
          }
        }
      }
    ]
  },
  "done_reason": "stop",
  "done": true,
  "total_duration": 3157877875,
  "load_duration": 579550042,
  "prompt_eval_count": 244,
  "prompt_eval_duration": 1794000000,
  "eval_count": 70,
  "eval_duration": 580000000
}
```

Now let's see how to do the same thing in Go with the Ollama Go API.

## Using the Ollama Go API for Tool Calling

Here's the complete Go code to do the same thing as the previous `curl` script (explanations follow):

```golang
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
		toolsLLM = "qwen2.5:1.5b"
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
	jsonTools, _ := json.Marshal(tools)

	// Unmarshal tools to create the tools list
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
	}

	err := client.Chat(ctx, req, func(resp api.ChatResponse) error {

		for _, toolCall := range resp.Message.ToolCalls {
			fmt.Println(toolCall.Function.Name, toolCall.Function.Arguments)
		}

		return nil
	})

	if err != nil {
		log.Fatalln("😡", err)
	}
}
```

### Explanations

Here's a step-by-step explanation of the code:

1. **Tool Definition**
- **First tool**: `helloTool`
  - Function to say hello to someone
  - Takes a "name" parameter of type string
  - Example: "Say hello to Bob"

- **Second tool**: `addNumbersTool`
  - Function to add two numbers
  - Takes two parameters "number1" and "number2"
  - Example: "add 28 to 12"

2. **Message Preparation for the prompt**
```go
messages := []api.Message{
    {Role: "user", Content: "Say hello to Bob"},
    {Role: "user", Content: "add 28 to 12"},
    {Role: "user", Content: "Say hello to Sarah"},
}
```

3. **Request Configuration**
```go
req := &api.ChatRequest{
    Model: toolsLLM,                 // Model to use (qwen2.5:1.5b)
    Messages: messages,              // List of messages
    Options: map[string]interface{}{
        "temperature": 0.0,          
        "repeat_last_n": 2,          
    },
    Tools: toolsList,                // List of available tools
    Stream: &FALSE,                  // No response streaming
}
```

4. **Sending and Processing the Request**
```go
err := client.Chat(ctx, req, func(resp api.ChatResponse) error {
    // For each tool call in the response
    for _, toolCall := range resp.Message.ToolCalls {
        // Display the function name and its arguments
        fmt.Println(toolCall.Function.Name, toolCall.Function.Arguments)
    }
    return nil
})
```

This code allows you to:
1. Connect to a local Ollama server
2. Define two tools (functions) available to the model
3. Send three messages requesting to use these tools
4. Retrieve and display the tool call results

All you need to do now is run the code to see the results.

### Execution

```bash
go run main.go
```

And you'll get the following results:

```bash
hello map[name:Bob]
add_numbers map[number1:28 number2:12]
hello map[name:Sarah]
```

There you have it, now you know how to use tool calling with Ollama. 

Now all you need to do is implement the `hello` and `add_numbers` functions so your application can execute them.

