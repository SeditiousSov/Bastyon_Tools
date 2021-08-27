package main

import (
	"fmt"
	"strings"

	"os/exec"
	"net/http"
	"encoding/json"
	
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
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

func handleProcess(w http.ResponseWriter, r *http.Request) {
	var rates []Rates
	var response string

        burl := r.FormValue("burl")

        if len(burl) == 0 {
                fmt.Fprintf(w, "Missing 'burl' Parameter")
                return
        }

	CleanURL(&burl)
	cmd := exec.Command(cli, "getpostscores", burl)
	jsn, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(w, "failed," + err.Error())
		return
	}

	err = json.Unmarshal(jsn, &rates)
	if err != nil {
		fmt.Fprintf(w, "failed," + err.Error())
		return
	}

	for _, rate := range rates {
		response += rate.Name + ": " + rate.Value + "<br/>"
	}

	fmt.Fprintf(w, response)
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	html := `
		<html>
			<head>
				<title>Tattle</title>
				<script type="text/javascript">
				function handle_click() {
					url = document.getElementById('burl').value;
					if (url == "") {
						alert("Missing URL");
						return;
					}

					xhttp = new XMLHttpRequest();
					xhttp.onreadystatechange = function() {
						if (this.readyState == 4 && this.status == 200) {
							cdiv = document.getElementById('rates');
							cdiv.innerHTML = this.responseText;
							return;
						}
					};

					encurl = encodeURIComponent(url)

					xhttp.open("POST", "/process", true);
					xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");

					xhttp.send('burl=' + encurl);
				}
				</script>
			</head>
			<body>
				<center>
					<h1>Tattle Tale</h1>
					<hr/><br/>
					<b>Bastyon URL</b><br/><br/>
					<input type="text" id="burl" name="burl" placeholder="https://bastyon.com/index?s=dfef303ac7c4e8c8754450bd37ad26ed7dfcecb25e844ed7b4efb49e25284b01&mpost=true" />
					<br/>
					<br/>
					<button id="clicker" onclick="javascript:handle_click()">Get Rates</button>
					<br/><br/>
					<div id="rates">
					</div>
				</center>
			</body>
		</html>
	`

	fmt.Fprintf(w, html)
}

func main() {
        cli = `/usr/local/bin/pocketcoin-cli`

        router := mux.NewRouter()
	router.HandleFunc("/process", handleProcess)
        router.HandleFunc("/", handleMain)

        corsObj := handlers.AllowedOrigins([]string{"*"})
        err := http.ListenAndServe(":80", handlers.CORS(corsObj)(router))
        if err != nil {
                fmt.Println("ListenAndServe: " + err.Error())
        }
}
