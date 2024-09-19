package openai

import (
	"github.com/go-resty/resty/v2"
	"fmt"
	"os"
	"log"
)

// Define a structure for the request body
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
	Temperature float64    `json:"temperature"`
}

// Structure for each message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Define a structure to parse the response
type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message ChatMessage `json:"message"`
}

func AskChatGpt(books string){
	apiKey := os.Getenv("OPEN_AI_KEY") //sk-iAPMrIW5JNGREsIZsjbWQzDMeNxy29WG6UloDvP64BT3BlbkFJOLuGBk8gbUTHVRwhjjo-XOKsUzBvamh2Tp7dIIlSsA

	url := "https://api.openai.com/v1/chat/completions"


	// Create a Resty client
	client := resty.New()

	// Create the request payload
	requestBody := ChatRequest{
		Model: "gpt-4", // Change to "gpt-3.5-turbo" if needed
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Can you suggest 5 books similar to the following"+books},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	// Send the POST request
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+apiKey).
		SetBody(requestBody).
		SetResult(&ChatResponse{}). // Response gets automatically unmarshalled into ChatResponse struct
		Post(url)

	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	// Print status code for debugging
	fmt.Println("Status Code:", resp.StatusCode())

	// Check if the response was successful
	if resp.IsError() {
		fmt.Printf("API returned an error: %s\n", resp.String())
		return
	}

	// Extract and print the response message
	chatResponse := resp.Result().(*ChatResponse)
	if len(chatResponse.Choices) > 0 {
		fmt.Printf("ChatGPT Response:", chatResponse.Choices)
	} else {
		fmt.Println("No response from ChatGPT")
	}
}
