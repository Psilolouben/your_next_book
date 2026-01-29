package main

import (
	"encoding/csv"
	"os"
	"log"
	"strconv"
	"book_proposals/models"
  "context"
	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Output struct {
	Books []book_proposals.Book `json:"books"`
}

type Input struct{}  // empty input

func csvData(filePath string)(records [][]string) {
	f, err := os.Open(filePath)
	if err != nil {
        log.Fatal("Unable to read input file " + filePath, err)
    }

	csvReader := csv.NewReader(f)
	records, err = csvReader.ReadAll()
    if err != nil {
        log.Fatal("Unable to parse file as CSV for " + filePath, err)
    }

	f.Close()
	return
}

func fetchBooksList(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input Input,
) (*mcp.CallToolResult,
	Output,
	error,) {

	bks,_ := LoadBooksFromCSV("./goodreads_library_export.csv")
	return nil, Output{Books: bks}, nil
}

func LoadBooksFromCSV(path string) ([]book_proposals.Book, error) {
	bookList := csvData(path)

	var bks []book_proposals.Book
	for _, bk := range bookList {
		rating, err := strconv.Atoi(bk[7])
		if err != nil {
			continue
		}

		bks = append(bks, book_proposals.Book{
			Rating: rating,
			Author: bk[2],
			Title:  bk[1],
		})
	}

	return bks, nil
}

func main() {
	// Create a server with a single tool.
	server := mcp.NewServer(&mcp.Implementation{Name: "booklist", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "fetch_books_list",
		Description: "Returns all books from the Goodreads CSV",
	}, fetchBooksList)
	// Run the server over stdin/stdout, until the client disconnects.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}

	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }
}
