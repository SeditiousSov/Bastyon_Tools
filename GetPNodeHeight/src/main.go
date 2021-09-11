package main

import(
	"fmt"
	"time"
	"strings"
	"context"

	"net/http"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
	"github.com/gorilla/mux"
)

func handleGetLatestBlock(w http.ResponseWriter, r *http.Request) {
	var nodes []*cdp.Node
	var block string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	err := chromedp.Run(ctx, 
			    chromedp.Navigate("https://pocketnet.app/blockexplorer/"), 
			    chromedp.Sleep(3 * time.Second),
			    chromedp.Nodes("A", &nodes),
	)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	for _, node := range nodes {
		if node.NodeName == "A" {
			href := node.AttributeValue("href")
			if ! strings.Contains(href, "block/") {
				continue
			}

			for _, child := range node.Children {
				if child.NodeName == "#text" {
					nodeval := child.NodeValue
					if strings.HasPrefix(nodeval, " ") {
						block = strings.TrimSpace(nodeval)
						break
					}
				}
			}
		}
	}

	if block == "" {
		fmt.Fprintf(w, "failed to get block\n")
		return
	}

	fmt.Fprintf(w, block + "\n")
}


func main() {
	router := mux.NewRouter()
        router.HandleFunc("/getheight", handleGetLatestBlock)
	err := http.ListenAndServe(":8181", router)
	if err != nil {
		fmt.Println(err.Error())
	}
}
