package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/atotto/clipboard"
)

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents the structure of the request payload
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// OpenAIResponse represents the structure of the response from OpenAI
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func sendToOpenAI(prompt string, apiKey string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// Define the system message and user prompt
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a problem solver and only return the correct answer, without any other text.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	requestBody := OpenAIRequest{
		Model:    "gpt-4o", // Change this to the appropriate model
		Messages: messages,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: %s", body)
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) > 0 {
		return openAIResp.Choices[0].Message.Content, nil
	}

	return "", nil
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	var lastText string

	for {
		text, err := clipboard.ReadAll()
		if err != nil {
			log.Fatalf("Failed to read clipboard: %v", err)
		}

		if text != lastText {
			fmt.Printf("New clipboard text detected: %s\n", text)
			response, err := sendToOpenAI(text, apiKey)
			if err != nil {
				log.Printf("Error sending to OpenAI: %v", err)
			} else {
				fmt.Printf("Response from OpenAI: %s\n", response)
			}
			lastText = text
		}

		time.Sleep(2 * time.Second) // Adjust the sleep duration as needed
	}
}
