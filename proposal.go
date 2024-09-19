package main

import "fmt"
import "github.com/go-resty/resty/v2"

func main() {
	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		Get("https://httpbin.org/get")

	if err == nil {
		fmt.Println(resp.Time())
	} else {
		fmt.Println(err)
	}
}
