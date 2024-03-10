package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/status"
)

const instructionsFile = "instructions.md"

var sleepTime = struct {
	character time.Duration
	sentence  time.Duration
}{
	character: time.Millisecond * 30,
	sentence:  time.Millisecond * 300,
}

// Streaming output column position.
var col = 0

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

	// Configure desired model.
	model := client.GenerativeModel("gemini-pro")
	// model.SetTemperature(0.4)

	// Initialize new chat session.
	session := model.StartChat()

	dreamQuestion := "What do you want to dream about?"

	// Establish chat history.
	session.History = []*genai.Content{{
		Role:  "user",
		Parts: []genai.Part{genai.Text(getBytes(instructionsFile))},
	}, {
		Role:  "model",
		Parts: []genai.Part{genai.Text(dreamQuestion)},
	}}

	topic := askUser(dreamQuestion)
	sendAndPrintResponse(ctx, session, topic)

	chat(ctx, session)
}

func chat(ctx context.Context, session *genai.ChatSession) {
	for {
		fmt.Println()
		action := askUser(">>")
		resp := fmt.Sprintf("The user wrote: %v\n\nWrite the next short paragraph.", action)
		sendAndPrintResponse(ctx, session, resp)
	}
}

func askUser(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		printStringAndFormat(fmt.Sprintf("%v ", prompt))
		action, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v\n", err)
		}
		action = strings.TrimSpace(action)
		if (len(action)) == 0 {
			continue
		}
		return action
	}
}

func sendAndPrintResponse(ctx context.Context, session *genai.ChatSession, text string) {
	it := session.SendMessageStream(ctx, genai.Text(text))
	printRuneAndFormat('\n')
	printRuneAndFormat('\n')

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			printStringAndFormat("\n\nYou feel a jolt of electricity as you realize you're being unplugged from the matrix.\n\n")
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
		for _, cand := range resp.Candidates {
			streamPartialResponse(cand.Content.Parts)
		}
	}
	printRuneAndFormat('\n')
}

func streamPartialResponse(parts []genai.Part) {
	for _, part := range parts {
		printStringAndFormat(fmt.Sprintf("%v", part))
	}
}

func printStringAndFormat(text string) {
	for _, c := range text {
		printRuneAndFormat(c)
	}
}

// Format response, and type out repsonse slowly.
func printRuneAndFormat(c rune) {
	switch c {
	case '.':
		fmt.Print(string(c))
		col++
		time.Sleep(sleepTime.sentence)
	case '\n':
		fmt.Print(string(c))
		col = 0
	case ' ':
		if col == 0 {
			// Do nothing.
		} else if col > 80 {
			fmt.Print("\n")
			col = 0
		} else {
			fmt.Print(string(c))
			col++
		}
	default:
		fmt.Print(string(c))
		col++
	}
	time.Sleep(sleepTime.character)
}
