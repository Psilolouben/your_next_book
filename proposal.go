package main

import (
	"fmt"
	//"github.com/go-resty/resty/v2"
	"encoding/csv"
	"os"
	"log"
	"strings"
	"sort"
	"strconv"
)

func askChatGpt(){
	//client := resty.New()
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

func filteredByShelf(books [][]string, shelfName string)(bks map[string]int){
	bks = make(map[string]int)
	for _, bk := range books {
		if strings.Contains(bk[18], shelfName) {
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

	rMap := filteredByShelf(r, "read")

	sort.Slice(rMap, func(i, j string) bool {
        return rMap[i].Value > rMap[j].Value
    })

	fmt.Println(rMap)
}
