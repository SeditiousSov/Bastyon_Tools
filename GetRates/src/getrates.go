package main

import (
	"os"
	"fmt"
	"strings"
	"runtime"

	"os/exec"
	"encoding/json"
)

type Rates struct {
	Posttxid   string `json:"posttxid"`
	Address    string `json:"address"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Reputation string `json:"reputation"`
	Value      string `json:"value"`
}

var cli string

func CleanURL(url *string) {
	parts := strings.Split(*url, "=")
	raw := strings.Split(parts[1], "&")
	*url = raw[0]
}

func main() {
	var rates []Rates

	if runtime.GOOS == "windows" {
		cli = ``
	} else {
	        cli = `/usr/local/bin/pocketcoin-cli`
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " \"<post_url>\"")
		return
	}

	url := os.Args[1]

	if ! strings.Contains(url, "=") {
		fmt.Println("Invalid URL To Parse")
		fmt.Println("Ex: " + os.Args[0] + " 'https://bastyon.com/index?i=68b3238b00b698e85cf870351d1080ffee640d02b7d687a77a923a9dbdacbe95&num=0&mimagegallery=true'")
		return
	}

	if ! strings.Contains(url, "&") {
		fmt.Println("Invalid URL To Parse")
		fmt.Println("Ex: " + os.Args[0] + " 'https://bastyon.com/index?i=68b3238b00b698e85cf870351d1080ffee640d02b7d687a77a923a9dbdacbe95&num=0&mimagegallery=true'")
		return
	}

	CleanURL(&url)

        cmd := exec.Command(cli, "getpostscores", url)
        jsn, err := cmd.Output()
        if err != nil {
                fmt.Println(err.Error())
                return
        }

        err = json.Unmarshal(jsn, &rates)
        if err != nil {
                fmt.Println(err.Error())
                return
        }

	for _, rate := range rates {
		fmt.Println(rate.Name + ": " + rate.Value)
	}
}
