package main

import (
	"os"
	"fmt"
	"bytes"
	"strings"

	"io/ioutil"
	"net/http"
	"encoding/json"

        "github.com/gorilla/mux"
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

var uname string
var pass string

func GetHTMLPage() (string) {
	html := `
	<html>
		<head>
			<title>LightWeight Bastion</title>
			<script type="text/javascript">
			function load()
			{
				gethistoricalstrip();
				setInterval(function() {
					gethistoricalstrip();	
				}, 90 * 1000);
			}

			function gethistoricalstrip() 
			{
				var display = "";

				cb = document.getElementById("pause");
				content = document.getElementById("content");

				if (cb.checked) {
					return;
				}

				xhttp = new XMLHttpRequest();
				xhttp.onreadystatechange = function() {
					if (this.readyState == 4 && this.status == 200) {
						if (! this.responseText.startsWith("{")) {
							alert(this.responseText);
							return
						}
						
						posts = JSON.parse(this.responseText);
						for (i = 0; i < posts.result.contents.length; i++) {
							post = posts.result.contents[i];
							display += '<div id="' + post.txid + '">';
							display += post.userprofile.name + "<br/>";
							display += decodeURIComponent(post.m);
							display += "<br/>"
							for (j = 0; j < post.i.length; j++) {
								display += '<img src="' + post.i[j] + '" /><br/>';
							}
							display += "</div><br/><hr/><br/>";
						}

						content.innerHTML = display;

						return;
					}
				};

				xhttp.open("GET", "/gethistoricalstrip", true);
				xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
				xhttp.send("");
			}
			</script>
		</head>
		<body onload="javascript:gethistoricalstrip();">
			<h1>Lightweight Bastyon</h1>
			<br/>
			<button type="button" onclick="location.reload();">Refresh</button>
			<input type="checkbox" id="pause" name="pause" value="Pause" />
			<label for="pause">Pause Feed</label>
			<hr/><br/><br/>
			<div id="content">
			</div>
		</body>
	</html>
	`

	return html
}

func handleWhoAreYou(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "LightWeight Bastyon")
}

func handlePing(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "pong")
}

func handleMain(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, GetHTMLPage())
}

func handleGetHistoricalStrip(w http.ResponseWriter, r *http.Request) {
	var posts Post

	rpccmd := `{"jsonrpc": "1.0", "id":"curltest", "method":"gethistoricalstrip", "params": [0, "", 300, "en"] }`
        body := bytes.NewReader([]byte(rpccmd))

        req, err := http.NewRequest("POST", "http://localhost:37071", body)
        if err != nil {
                fmt.Fprintf(w, err.Error())
                return
        }

        req.SetBasicAuth(uname, pass)
        req.Header.Set("Content-Type", "text/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer resp.Body.Close()

	jsn, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	err = json.Unmarshal(jsn, &posts)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Write(jsn)
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "--help" || os.Args[1] == "help" {
			fmt.Println("Server: " + os.Args[0])
			fmt.Println("File: " + os.Args[0] + " <filename>")
			return
		}

		// Write HTML To File		
	}

	// If you have set an rpc user/password
	// Delete From Here
        cookie, err := ioutil.ReadFile("/home/pnet/.pocketcoin/.cookie")
        if err != nil {
                fmt.Println(err.Error())
                return
        }

        parts := strings.Split(string(cookie), ":")
        uname = parts[0]
        pass = parts[1]
	// To Here

	// Rpc Login Data (Uncomment Below)
	// uname = "rpc_user_name"
	// pass = "rpc_password"

        router := mux.NewRouter()
        router.HandleFunc("/whoareyou", handleWhoAreYou)
        router.HandleFunc("/ping", handlePing)
        router.HandleFunc("/", handleMain)
        router.HandleFunc("/gethistoricalstrip", handleGetHistoricalStrip)

        err = http.ListenAndServe(":8181", router)
        if err != nil {
                fmt.Println("ListenAndServe: ", err)
        }
}
