package main

import (
	"encoding/csv"
	"os"
	"log"
	"strings"
	"sort"
	"strconv"
	"marky/openai"
)

func sortBooksByRating(bks map[string]int)(arr []string){
	arr = make([]string, 0, len(bks))
	for key := range bks {
		arr = append(arr, key)
	}

	sort.Slice(arr, func(i, j int) bool { return bks[arr[i]] > bks[arr[j]] })
	return arr
}

func filteredByShelfAndRating(books [][]string, shelfName string)(bks map[string]int){
	bks = make(map[string]int)
	for _, bk := range books {
		if (strings.Contains(bk[18], shelfName) && (bk[7] == "5")) {
			bks[bk[1]], _ = strconv.Atoi(bk[7])
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

func main() {
	r := csvData("./goodreads_library_export.csv")

	rMap := filteredByShelfAndRating(r, "read")

	rMapArr := sortBooksByRating(rMap)

	topBooksStr := strings.Join(rMapArr[:], ",")
	//fmt.Printf(topBooksStr)
	openai.AskChatGpt(topBooksStr)
}
