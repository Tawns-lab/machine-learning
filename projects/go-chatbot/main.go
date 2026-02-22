package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	model     = anthropic.ModelClaudeOpus4_6
	maxTokens = 8192
	system    = "You are a helpful, friendly assistant. Engage in natural conversation."
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: ANTHROPIC_API_KEY environment variable not set")
		os.Exit(1)
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	scanner := bufio.NewScanner(os.Stdin)
	history := []anthropic.MessageParam{}

	printBanner()

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Println("\nGoodbye!")
			break
		}

		history = append(history, anthropic.NewUserMessage(
			anthropic.NewTextBlock(input),
		))

		fmt.Print("\nClaude: ")
		reply, err := streamResponse(client, history)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nAPI error: %v\n", err)
			// Roll back the failed user message so the conversation stays consistent.
			history = history[:len(history)-1]
			continue
		}
		fmt.Println()

		// Only append the assistant turn after a successful, complete response.
		if reply != "" {
			history = append(history, anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(reply),
			))
		}
	}
}

// streamResponse sends the current conversation to the API using Server-Sent
// Events streaming and prints each text chunk to stdout as it arrives.
// It returns the complete assistant reply text.
func streamResponse(client anthropic.Client, history []anthropic.MessageParam) (string, error) {
	stream := client.Messages.NewStreaming(context.Background(), anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: maxTokens,
		System: []anthropic.TextBlockParam{
			{Text: system},
		},
		Messages: history,
	})

	var sb strings.Builder
	for stream.Next() {
		event := stream.Current()
		// event.Type == "content_block_delta" and event.Delta.Type == "text_delta"
		// are the only events carrying streaming text.
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			fmt.Print(event.Delta.Text)
			sb.WriteString(event.Delta.Text)
		}
	}
	if err := stream.Err(); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func printBanner() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║          LET'S TALK LIVE                 ║")
	fmt.Println("║   Powered by Claude (Opus 4.6)           ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println("Type 'quit' or 'exit' to end the session.")
}
