package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/option"
	"google.golang.org/grpc/status"
)

const initialPromptFile = "prompt.md"

// getBytes returns the file contents as bytes.
func getBytes(path string) []byte {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading file bytes %v: %v\n", path, err)
	}
	return bytes
}

func main() {
	ctx := context.Background()

	// New client, using API key authorization.
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatalf("Error creating client: %v\n", err)
	}
	defer client.Close()
	// log.Printf("client: %v", client)

	// Configure desired model.
	model := client.GenerativeModel("gemini-pro")
	model.SetTemperature(1)
	// log.Printf("model: %v", model)

	// Initialize new chat session.
	session := model.StartChat()
	// log.Printf("session: %v", session)

	// // Establish chat history.
	// session.History = []*genai.Content{{
	// 	Role:  "user",
	// 	Parts: []genai.Part{genai.Text(getBytes(initialPromptFile))},
	// }}

	send(ctx, session, string(getBytes(initialPromptFile)))
	chat(ctx, session)
}

func chat(ctx context.Context, session *genai.ChatSession) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\n>> ")
		action, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v\n", err)
		}

		send(ctx, session, action)
	}
}

func send(ctx context.Context, session *genai.ChatSession, text string) {
	resp, err := session.SendMessage(ctx, genai.Text(text))
	if err != nil {
		log.Printf("Error sending message: err=%v\n", err)

		var ae *apierror.APIError
		if errors.As(err, &ae) {
			log.Printf("ae.Reason(): %v\n", ae.Reason())
			log.Printf("ae.Details().Help.GetLinks(): %v\n", ae.Details().Help.GetLinks())
		}

		if s, ok := status.FromError(err); ok {
			log.Printf("s.Message: %v\n", s.Message())
			for _, d := range s.Proto().Details {
				log.Printf("- Details: %v\n", d)
			}
		}
		os.Exit(1)
	}
	// log.Printf("resp: %v", resp)

	// Display the response.
	for _, part := range resp.Candidates[0].Content.Parts {
		fmt.Printf("\n\n%v\n", part)
	}
}
