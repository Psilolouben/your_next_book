# your_next_book

A Go-based book recommendation tool that reads a Goodreads library export and surfaces personalized book suggestions using AI.

## What it does

1. Parses a Goodreads CSV export (`goodreads_library_export.csv`) to load the user's book list (title, author, rating, shelf/status, year published).
2. Exposes the book list via an **MCP (Model Context Protocol) server** (`proposal.go`) — the tool is named `fetch_books_list` and runs over stdin/stdout.
3. An MCP client (e.g. Claude Desktop) calls the tool, receives the book list, and uses the embedded prompt in `ToolDescription()` to generate scored recommendations.

The recommendation logic (in the prompt) scores candidates as:
- +5 thematic/genre overlap with 2+ favorites
- +3 co-read patterns among similar readers
- +1 same author as a favorite

Books already on the user's "read" shelf are excluded. Top 10 by score are returned as JSON.

## Structure

```
proposal.go          — MCP server entry point; CSV loading; tool description/prompt
models/book.go       — Book struct
openai/openai.go     — Legacy OpenAI REST client (not used by the MCP server)
goodreads_library_export.csv — User's Goodreads export
```

## Running

```bash
go run proposal.go
```

The server communicates over stdin/stdout (stdio MCP transport). Configure it as an MCP tool in Claude Desktop or another MCP client.

## Environment

Requires a `.env` file (loaded via `godotenv`). The OpenAI client uses `OPEN_AI_KEY` but this is legacy — the MCP server itself does not call any external API directly.

## Notes

- The `openai/` package is a leftover from an earlier direct-API approach and is not wired into the MCP server.
- The scoring and recommendation logic lives entirely in the prompt string inside `ToolDescription()`.
- CSV parsing skips rows where the rating field is not a valid integer (header row and unrated books).
