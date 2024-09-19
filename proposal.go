package main

import (
	"fmt"
	//"github.com/go-resty/resty/v2"
	"encoding/csv"
	"os"
	"log"
	"strings"
)

func filteredByShelf(books [][]string, shelfName string)(bks [][]string){
	for _, bk := range books {
		if Contains(bk[16], shelfName) {
			bks = append(bks, bk)
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
	//client := resty.New()
	
	r := csvData("./goodreads_library_export.csv")
	

	r = filteredByShelf(r, "looking-for")
	fmt.Println(r)

	//resp, err := client.R().
	//	EnableTrace().
	//	Get("https://httpbin.org/get")
//
	//if err == nil {
	//	fmt.Println(resp.Time())
	//} else {
	//	fmt.Println(err)
	//}
}
