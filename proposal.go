package main

import (
	"encoding/csv"
	"os"
	"log"
	"strings"
	"sort"
	"strconv"
	"book_proposals/models"
	"marky/openai"
	"github.com/joho/godotenv"
	//"fmt"
)

func sortBooksByRating(bks []book_proposals.Book)(barr []book_proposals.Book){
	var arr []book_proposals.Book

	for _, key := range bks {
		arr = append(arr, key)
	}

	sort.Slice(arr, func(i, j int) bool { return arr[i].Rating > arr[j].Rating })
	return arr
}

func filteredByShelfAndRating(sheet_books [][]string, shelfName string)(books []book_proposals.Book){
	var bks []book_proposals.Book
	for _, bk := range sheet_books {
		if (strings.Contains(bk[18], shelfName) && (bk[7] == "5")) {
			book_rating, _ := strconv.Atoi(bk[7])
			bks = append(bks,
				book_proposals.Book{
					Rating: book_rating,
					Author: bk[2],
					Title: bk[1],
				},
			)
		}
	}
	return bks
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

func constructPromptBookTitles(books []book_proposals.Book)(book_str string) {
	var bks string

	for _, b := range books {
		bks = bks + b.Title + " by " + b.Author + ","
	}

	return bks
}

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

	r := csvData("./goodreads_library_export.csv")

	rMap := filteredByShelfAndRating(r, "read")

	rMapArr := sortBooksByRating(rMap)

	topBooksStr := constructPromptBookTitles(rMapArr)

	//fmt.Println(topBooksStr)
	openai.AskChatGpt(topBooksStr)
}
