# go-chatbot

A terminal chatbot that streams responses live from the Claude API.

## Requirements

- Go 1.24+
- An [Anthropic API key](https://console.anthropic.com/)

## Setup

```bash
cd projects/go-chatbot
go mod download
```

## Usage

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
go run .
```

Or build a binary first:

```bash
go build -o chatbot .
./chatbot
```

Type `quit` or `exit` (or press **Ctrl+D**) to end the session.
Press **Ctrl+C** at any time to interrupt a response and exit cleanly.

## How it works

- Maintains the full conversation history across turns.
- Streams each response token-by-token using Server-Sent Events so you see
  output as it is generated rather than waiting for the full reply.
- If the API returns an error the failed user turn is rolled back so you can
  rephrase and retry without corrupting the history.
- A `SIGINT` or `SIGTERM` signal cancels the in-flight HTTP request
  immediately and saves any partial reply already received before exiting.
