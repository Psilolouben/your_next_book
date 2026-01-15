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
	apiKey := os.Getenv("OPEN_AI_KEY")

	url := "https://api.openai.com/v1/chat/completions"


	// Create a Resty client
	client := resty.New()

	// Create the request payload
	requestBody := ChatRequest{
		Model: "gpt-4", // Change to "gpt-3.5-turbo" if needed
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a literary recommendation engine with expert knowledge of books, genres, literary movements, and reader co-reading patterns. You strictly follow scoring rules, exclusion rules, and output format requirements."},
			{
				Role: "user",
				Content: `I want you to suggest 8 books that you believe
				I would like based on the book titles of my top rated books
				in the following section which are featured in a {title} by {author} format. The list is this` + books +
				`. Do not recommend books based solely based on the authors I seem to like but on what people who have similar taste as I do usually read as well.
				Books that match more than one of my top rated books at the same time should be considered higher recommended and should have higher priority.
				Here are some examples:
				- A person has read Hobbit and two of the Lord of the rings books so it the third Lord of the Rings book matches with 3 of Tolkien's books. This would be high priority.
				- A person has read some Hegel's books and Hegel is notorious for his clash with Kierkegaard, thus many people who read Hegel read some Kierkegaard as well. Kierkegaard should be recommended.
				- A person loves The Road by MacCarthy and The Passage by Justin Cronin. He seems to like post apocalyptic books so these books should be recommended as well.
				- A person has read a book by Stephen King. He would likely like another book by Stephen King but from all the above this would be lowest in priority.

				Each recommended book is represented by an object with the following keys:
					- Recommended Book Title,
					- Author,
					- Reason of recommendation,
					- Calculated Score

				Calculated Score is calculated with the following scoring system:
				- +3 points: strong thematic or genre overlap with TWO OR MORE favorite books
				- +2 points: commonly co-read by readers with similar taste
				- +1 point: same author as a favorite book

				The "Reason of recommendation" field MUST explicitly list:
				- Genres or themes and how many favorite books they match
				- Reader behavior (co-read patterns), if applicable
				- Author overlap ONLY if applicable

				Each reason must reference the scoring criteria used.
				For example "Genres: noir, crime, drama matching with X of your favorite books, Author: Y books of Manchette already in your favorites list"

				Exclusion rules:
				- If a book already appears in the input list, DO NOT include it in the results.

				Output requirements:
				- Return ONLY valid JSON
				- The top-level JSON value MUST be an array
				- Each element of the array MUST be a JSON object (hash/map) with the keys described above
				- Each object MUST use the same keys
				- Do NOT wrap the array in another object
				- Do NOT number the items
				- Do NOT include any text outside the JSON
				After gathering the recommended books and return the json with all the books with a calculated score greater than 3.`,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.5,
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
