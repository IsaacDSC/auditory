package main

import (
	"fmt"
	"net/url"
)

func main() {

	queryUrl, err := url.Parse("?token=1234567890&api_key=1234567890")
	if err != nil {
		panic(err)
	}

	for key, value := range queryUrl.Query() {
		fmt.Println(key, value)
	}
}
