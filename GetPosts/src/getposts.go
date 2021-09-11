package main
/*
#include <stdio.h>
#include <unistd.h>
#include <termios.h>
char getch(){
    char ch = 0;
    struct termios old = {0};
    fflush(stdout);
    if( tcgetattr(0, &old) < 0 ) perror("tcsetattr()");
    old.c_lflag &= ~ICANON;
    old.c_lflag &= ~ECHO;
    old.c_cc[VMIN] = 1;
    old.c_cc[VTIME] = 0;
    if( tcsetattr(0, TCSANOW, &old) < 0 ) perror("tcsetattr ICANON");
    if( read(0, &ch,1) < 0 ) perror("read()");
    old.c_lflag |= ICANON;
    old.c_lflag |= ECHO;
    if(tcsetattr(0, TCSADRAIN, &old) < 0) perror("tcsetattr ~ICANON");
    return ch;
}
*/
import "C"

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"bufio"
	"strings"

	"os/exec"
	"net/url"
	"net/http"
	"os/signal"
	"io/ioutil"
	"encoding/json"
)

type Post struct {
	Result struct {
		Height   int `json:"height"`
		Contents []struct {
			Txid     string   `json:"txid"`
			Address  string   `json:"address"`
			Time     string   `json:"time"`
			L        string   `json:"l"`
			C        string   `json:"c"`
			M        string   `json:"m"`
			U        string   `json:"u"`
			Type     string   `json:"type"`
			ScoreSum string   `json:"scoreSum"`
			ScoreCnt string   `json:"scoreCnt"`
			T        []string `json:"t"`
			I        []string `json:"i"`
			S        struct {
				A      interface{} `json:"a"`
				V      interface{} `json:"v"`
				Videos interface{} `json:"videos"`
				Image  interface{} `json:"image"`
			} `json:"s"`
			Comments    int `json:"comments"`
			Userprofile struct {
				Address    string  `json:"address"`
				Name       string  `json:"name"`
				ID         int     `json:"id"`
				I          string  `json:"i"`
				B          string  `json:"b"`
				R          string  `json:"r"`
				Reputation float64 `json:"reputation"`
				Postcnt    int     `json:"postcnt"`
				Rc         int     `json:"rc"`
			} `json:"userprofile"`
		} `json:"contents"`
		ContentsTotal int `json:"contentsTotal"`
	} `json:"result"`
	Error interface{} `json:"error"`
	ID    string      `json:"id"`
}

type Payload struct {
        Jsonrpc string   `json:"jsonrpc"`
        ID      string   `json:"id"`
        Method  string   `json:"method"`
        Params  []string `json:"params"`
}

var uname string
var pass string

func main() {
	var posts Post

	cookie := os.Getenv("BASTYON_PASSWORD")
        parts := strings.Split(string(cookie), ":")
        uname = parts[0]
        pass = parts[1]

        signalChannel := make(chan os.Signal, 2)
        signal.Notify(signalChannel, os.Interrupt)
        go func() {
                for {
                        <-signalChannel

			cmd := exec.Command("reset")
			cmd.Stdout = os.Stdout
			cmd.Run()
			os.Exit(1)
                }
        }()

	rpccmd := `{"jsonrpc": "1.0", "id":"curltest", "method":"gethistoricalstrip", "params": [0, "", 300, "en"] }`
        body := bytes.NewReader([]byte(rpccmd))


	// Change to localhost (127.0.0.1) or your own node
        req, err := http.NewRequest("POST", "http://74.208.128.141:37071", body)
        if err != nil {
                fmt.Println(err.Error())
                return
        }

        req.SetBasicAuth(uname, pass)
        req.Header.Set("Content-Type", "text/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	jsn, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = json.Unmarshal(jsn, &posts)
	if err != nil {
		fmt.Printf("%+v\n", posts)
		fmt.Println(err.Error())
		return
	}

	for _, post := range posts.Result.Contents {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()

		decoded, err := url.QueryUnescape(post.M)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		fmt.Print(post.Userprofile.Name + "\t\t")
		if len(post.T) != 0 {
			for _, t := range post.T {
				fmt.Print("#" + t + " ")
			}

			fmt.Println("")
		}

		fmt.Println("")
		fmt.Println("https://bastyon.com/post?s=" + post.Txid)
		fmt.Println("")

		fmt.Println("----------------------")
		fmt.Println(decoded)
		fmt.Println(" ")

		if len(post.I) != 0 {
			for _, i := range post.I {
				fmt.Print(i + " ")
			}
			fmt.Println("")
		}

		fmt.Println(" ")

		key := string(C.getch())

		switch key {
		case "q":
			cmd := exec.Command("reset")
			cmd.Stdout = os.Stdout
			cmd.Run()
			os.Exit(0)
		case "b":
			go func() {
				cmd := exec.Command("brave-browser", "https://bastyon.com/post?s=" + post.Txid)
				cmd.Run()
			}()

			fmt.Println("Press Enter To Proceed")
			C.getch()
		case "o":
			var openlist []string
			if len(post.I) != 0 {
				for _, i := range post.I {
					openlist = append(openlist, i)
				}
			}

			if len(post.U) != 0 {
				durl, err := url.QueryUnescape(post.U)
				if err != nil {
					fmt.Println(err.Error())
					time.Sleep(500 * time.Millisecond)
					continue
				}

				openlist = append(openlist, durl)
			}

			if len(openlist) > 0 {
				go func() {
					cmd := exec.Command("brave-browser", openlist...)
					cmd.Run()
				}()

				fmt.Println("Press Enter To Proceed")
				C.getch()
			}
		case "p":
			var params []string

			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()

			fmt.Println("What's New? (Enter To Complete): ")
			reader := bufio.NewReader(os.Stdin)

			rawtext, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			text := strings.TrimSpace(rawtext)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			
			params = append(params, text)

			fmt.Println("")
			fmt.Println("Add Categories/Tags (Enter To Complete): ")
		
			tagsraw, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
	
			tagsraw = strings.TrimSpace(tagsraw)
			tagsraw = strings.Replace(tagsraw, "#", "", -1)
			tags := strings.Split(tagsraw, " ")

			params = append(params, tags...)

			fmt.Println("Posting... ")	
			cmd = exec.Command("/usr/bin/bpost", params...)
			cmd.Stdout = os.Stdout
			cmd.Run()
			fmt.Println("Complete")
			time.Sleep(2 * time.Second)
		}
	}
}
