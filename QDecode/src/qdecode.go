package main

import (
	"os"
	"io"
	"fmt"
	"bufio"
	"strings"

	"net/url"
)

func main() {
	var output string

	info, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if info.Mode() & os.ModeNamedPipe == 0 {
		fmt.Println("This command is intended to work with pipes")
		fmt.Println("Usage: echo hello%20world | " + os.Args[0])
		return
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}

		output += string(input)
	}

	output = strings.TrimSpace(output)
	decoded, err := url.QueryUnescape(output)

	fmt.Println(decoded)
}
