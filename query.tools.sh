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

echo "Sending: ${DATA} on ${SERVICE_URL}"
echo ""

curl --no-buffer ${SERVICE_URL}/api/chat \
    -H "Content-Type: application/json" \
    -d "${DATA}" | jq '.'

echo ""

