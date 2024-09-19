package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"encoding/csv"
	"os"
	"log"
)

func main() {
	client := resty.New()
	csv_file_path := "./goodreads_library_export.csv"

	f, err := os.Open(csv_file_path)
	if err != nil {
        log.Fatal("Unable to read input file " + filePath, err)
    }
    defer f.Close()

	r := csv.NewReader(f)
	records, err := csvReader.ReadAll()
    if err != nil {
        log.Fatal("Unable to parse file as CSV for " + filePath, err)
    }

	fmt.Println(records)

	resp, err := client.R().
		EnableTrace().
		Get("https://httpbin.org/get")

	if err == nil {
		fmt.Println(resp.Time())
	} else {
		fmt.Println(err)
	}
}
