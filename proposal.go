package main

import "fmt"
import "github.com/go-resty/resty/v2"
import "encoding/csv"
import "strings"

func main() {
	client := resty.New()
	//csv_file_path := "./goodreads_library_export.csv"

	in := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
	r := csv.NewReader(strings.NewReader(in))

	fmt.Println(r)
	resp, err := client.R().
		EnableTrace().
		Get("https://httpbin.org/get")

	if err == nil {
		fmt.Println(resp.Time())
	} else {
		fmt.Println(err)
	}
}
