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

type Addr struct {
	Name string `json:"name"`
	Address string `json:"address"`
}

type Posts struct {
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
		LastComment struct {
			ID         string `json:"id"`
			Postid     string `json:"postid"`
			Address    string `json:"address"`
			Time       string `json:"time"`
			TimeUpd    string `json:"timeUpd"`
			Block      string `json:"block"`
			Msg        string `json:"msg"`
			Parentid   string `json:"parentid"`
			Answerid   string `json:"answerid"`
			ScoreUp    string `json:"scoreUp"`
			ScoreDown  string `json:"scoreDown"`
			Reputation string `json:"reputation"`
			Edit       bool   `json:"edit"`
			Deleted    bool   `json:"deleted"`
			MyScore    int    `json:"myScore"`
			Children   string `json:"children"`
		} `json:"lastComment,omitempty"`
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
		Reposted int `json:"reposted,omitempty"`
	} `json:"contents"`
	ContentsTotal int `json:"contentsTotal"`
}


var cli string

func CleanURL(url *string) {
	parts := strings.Split(*url, "=")
	raw := strings.Split(parts[1], "&")
	*url = raw[0]
}

func main() {
	var rates []Rates
	var addr []Addr
	var posts Posts

	if runtime.GOOS == "windows" {
		cli = ``
	} else {
	        cli = `/usr/local/bin/pocketcoin-cli`
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " \"<bastyon_name>\"")
		return
	}

	name := os.Args[1]

	cmd := exec.Command(cli, "getuseraddress", name)
	ajsn, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = json.Unmarshal(ajsn, &addr)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cmd = exec.Command(cli, "getusercontents", addr[0].Address)
	pjsn, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	
	err = json.Unmarshal(pjsn, &posts)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	
	for _, post := range posts.Contents {
		cmd = exec.Command(cli, "getpostscores", post.Txid)
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
			if rate.Value == "1" {
				fmt.Println(post.Txid + " - " + rate.Name + ": " + rate.Value)
			}
		}
	}
}
