package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"book_proposals/goodreads"
	book_proposals "book_proposals/models"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Output struct {
	Books []book_proposals.Book `json:"books"`
}

type Input struct{} // empty input

type SyncInput struct{}

type SyncOutput struct {
	Message string `json:"message"`
}

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

	binDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	bks, _ := LoadBooksFromCSV(filepath.Join(binDir, "goodreads_library_export.csv"))
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
			Status: bk[18],
			YearPublished: bk[12],
		})
	}

	return bks, nil
}

func syncGoodreadsExport(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SyncInput,
) (*mcp.CallToolResult, SyncOutput, error) {
	binDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if err := goodreads.FetchExport(filepath.Join(binDir, "goodreads_library_export.csv")); err != nil {
		return nil, SyncOutput{}, err
	}
	return nil, SyncOutput{Message: "Export downloaded successfully. You can now call fetch_books_list."}, nil
}

func ToolDescription() string {
	return `You have practically all the knowledge of books in the worldk. You have my Goodreads book list consisting of entries containing the book title, author, my rating and the shelves I have the book on.
			  When the book is on the "read" shelf it means I have already read it, "to-read" is the once I am planning to read, "looking-for" for the ones I am thinking of buying.
				I want you to suggest books that you believe I would like based on the book titles of my top rated books on Goodreads. Do not recommend books based solely based on the authors I seem to like but on what people who have similar taste as I do usually read as well.
				Do not recommend books that are on my "read" list either. But feel free to recommend books that on my "to-read" list if you really think I would like them or books that are in none of my lists.
				Books that match more than one of my top rated books at the same time should be considered higher recommended and should have higher priority.
				The books do not need to be among the owns on my to-read list but the client can feel free to propose books outside of my list altogether.
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
				- +5 points: strong thematic or genre overlap with TWO OR MORE favorite books
				- +3 points: commonly co-read by readers with similar taste
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
				After gathering the recommended books and return the json with the top 10 books with the highest calculated scores in
				descending order of calculated score`
}

func main() {
	// Resolve paths relative to the binary so the server works regardless of working directory
	binDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	godotenv.Load(filepath.Join(binDir, ".env"))

	server := mcp.NewServer(&mcp.Implementation{Name: "booklist", Version: "v1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "fetch_books_list",
		Description: ToolDescription(),
	}, fetchBooksList)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "sync_goodreads_export",
		Description: "Opens a browser window, logs into Goodreads if needed, triggers a library export, waits for it to be ready, and downloads it — all in one step. Call this to refresh the local book list before calling fetch_books_list.",
	}, syncGoodreadsExport)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
