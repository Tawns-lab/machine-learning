package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	model     = anthropic.ModelClaudeOpus4_6
	maxTokens = 8192
	system    = "You are a helpful, friendly assistant. Engage in natural conversation."

	// scannerMaxSize allows up to 1 MiB per input line so large pastes are
	// never silently truncated by the default 64 KiB bufio.Scanner limit.
	scannerMaxSize = 1 << 20
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: ANTHROPIC_API_KEY environment variable is not set")
		os.Exit(1)
	}

	// Cancel on SIGINT (Ctrl+C) or SIGTERM so in-flight API requests stop cleanly.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	// Buffered stdout: reduces syscall overhead for the many small streaming
	// writes, while we flush explicitly before each blocking Scan call.
	out := bufio.NewWriter(os.Stdout)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, scannerMaxSize), scannerMaxSize)

	var history []anthropic.MessageParam

	printBanner(out)
	_ = out.Flush()

	for {
		fmt.Fprint(out, "\nYou: ")
		_ = out.Flush() // must appear before Scan blocks on stdin

		if !scanner.Scan() || ctx.Err() != nil {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Fprintln(out, "\nGoodbye!")
			_ = out.Flush()
			break
		}

		history = append(history, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))

		fmt.Fprint(out, "\nClaude: ")
		_ = out.Flush()

		reply, err := streamResponse(ctx, &client, out, history)
		fmt.Fprintln(out) // newline after streamed content
		_ = out.Flush()

		switch {
		case err == nil:
			// Full response received – add to history.
			if reply != "" {
				history = append(history, anthropic.NewAssistantMessage(anthropic.NewTextBlock(reply)))
			}

		case errors.Is(err, context.Canceled):
			// Signal received mid-stream. Preserve whatever arrived so the
			// conversation history remains coherent, then exit.
			if reply != "" {
				history = append(history, anthropic.NewAssistantMessage(anthropic.NewTextBlock(reply)))
			}
			fmt.Fprintln(out, "(response interrupted – exiting)")
			_ = out.Flush()
			return

		default:
			fmt.Fprintf(os.Stderr, "API error: %v\n", err)
			// Roll back the user turn so history stays consistent.
			history = history[:len(history)-1]
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "stdin error: %v\n", err)
	}
}

// streamResponse sends the current conversation to the Claude API via
// Server-Sent Events and writes each text chunk to w as it arrives.
// The context controls cancellation; context.Canceled is returned when the
// caller's signal fires. The accumulated reply is always returned even when
// the stream is interrupted, so callers can preserve a partial response.
func streamResponse(
	ctx context.Context,
	client *anthropic.Client,
	w *bufio.Writer,
	history []anthropic.MessageParam,
) (string, error) {
	stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: maxTokens,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Messages:  history,
	})

	var sb strings.Builder
	for stream.Next() {
		event := stream.Current()
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			sb.WriteString(event.Delta.Text)
			fmt.Fprint(w, event.Delta.Text)
			_ = w.Flush() // flush each chunk for real-time display
		}
	}
	return sb.String(), stream.Err()
}

func printBanner(w *bufio.Writer) {
	fmt.Fprintln(w, "╔══════════════════════════════════════════╗")
	fmt.Fprintln(w, "║          LET'S TALK LIVE                 ║")
	fmt.Fprintln(w, "║   Powered by Claude (Opus 4.6)           ║")
	fmt.Fprintln(w, "╚══════════════════════════════════════════╝")
	fmt.Fprintln(w, "Type 'quit' or 'exit' to end the session.")
}
