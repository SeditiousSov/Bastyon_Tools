package main

import (
	"fmt"
	"sync"
	"bytes"
	"strings"

	"os/exec"
	"net/http"
	"io/ioutil"
	"encoding/json"
	
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

type Reputations struct {
        Address string `json:"address"`
}

type Blocked struct {
        Address    string  `json:"address"`
        Name       string  `json:"name"`
        Blocking    []string `json:"blocking"`
}

type BlockedConcurrent struct {
        Result []struct {
                Address    string  `json:"address"`
                Name       string  `json:"name"`
                Blocking    []string `json:"blocking"`
        } `json:"result"`
}

type Payload struct {
        Jsonrpc string   `json:"jsonrpc"`
        ID      string   `json:"id"`
        Method  string   `json:"method"`
        Params  []string `json:"params"`
}


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
var blockedc []BlockedConcurrent
var uname string
var pass string

func GetUserProfileConcurrent(addr string, wg *sync.WaitGroup) {
        var params []string
        var blocking BlockedConcurrent

        defer wg.Done()

        params = append(params, addr)

        data := Payload {
                Jsonrpc: "1.0",
                ID: "curltext",
                Method: "getuserprofile",
                Params: params,
        }

        payloadBytes, err := json.Marshal(data)
        if err != nil {
                fmt.Println(err.Error())
                return
        }
        body := bytes.NewReader(payloadBytes)

        req, err := http.NewRequest("POST", "http://localhost:37071", body)
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

        err = json.Unmarshal(jsn, &blocking)
        if err != nil {
                fmt.Println(err.Error())
                return
        }

        blockedc = append(blockedc, blocking)
        return
}


func CleanURL(url *string) {
	parts := strings.Split(*url, "=")
	raw := strings.Split(parts[1], "&")
	*url = raw[0]
}

func GetUserAddress(name string) (string, error) {
        var addr []Addr

        cmd := exec.Command(cli, "getuseraddress", name)
        ajsn, err := cmd.Output()
        if err != nil {
                return "", err
        }

        err = json.Unmarshal(ajsn, &addr)
        if err != nil {
                return "", err
        }

        return addr[0].Address, nil
}

func GetAllAddresses() ([]string, error) {
        var rep []Reputations
        var retval []string

        cmd := exec.Command(cli, "getreputations")
        jsn, err := cmd.Output()
        if err != nil {
                fmt.Println(err.Error())
                return retval, err
        }

        err = json.Unmarshal(jsn, &rep)
        if err != nil {
                fmt.Println(err.Error())
                return retval, err
        }

        for _, r := range rep {
                retval = append(retval, r.Address)
        }

        return retval, nil
}

func handleGetBlocks(w http.ResponseWriter, r *http.Request) {
	var tosend string
        var names []string

	user := r.FormValue("user")

        uaddr, err := GetUserAddress(user)
        if err != nil {
                fmt.Fprintf(w, "failed," + err.Error())
                return
        }

        _ = uaddr

        addrs, err := GetAllAddresses()
        if err != nil {
                fmt.Fprintf(w, "failed," + err.Error())
                return
        }

        wg := sync.WaitGroup{}
	completed := 0

        for i := 0; i < len(addrs); i += 5 {
		if len(addrs) - completed >= 5 {
                        wg.Add(5)

                        addr := addrs[i]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+1]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+2]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+3]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+4]
                        go GetUserProfileConcurrent(addr, &wg)

                        completed += 5

		} else if len(addrs) - completed >= 4 {
                        wg.Add(4)

                        addr := addrs[i]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+1]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+2]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+3]
                        go GetUserProfileConcurrent(addr, &wg)

                        completed += 4
                } else if len(addrs) - completed >= 3 {
                        wg.Add(3)

                        addr := addrs[i]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+1]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+2]
                        go GetUserProfileConcurrent(addr, &wg)

                        completed += 3
                } else if len(addrs) - completed >= 2 {
                        wg.Add(2)

                        addr := addrs[i]
                        go GetUserProfileConcurrent(addr, &wg)

                        addr = addrs[i+1]
                        go GetUserProfileConcurrent(addr, &wg)

                        completed += 2
                } else {
                        wg.Add(1)

                        addr := addrs[i]
                        go GetUserProfileConcurrent(addr, &wg)

                        completed += 1
                }

                wg.Wait()
        }

        for _, b := range blockedc {
                flag := false
                for _, bl := range b.Result[0].Blocking {
                        if uaddr == bl {
                                flag = true
                        }
                }

                if flag {
                        names = append(names, b.Result[0].Name)

                }
        }

        for _, name := range names {
		tosend += name + "<br/>"
        }

	fmt.Fprintf(w, tosend)
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


func handleProcessAll(w http.ResponseWriter, r *http.Request) {
	var rates []Rates
        var addr []Addr
	var response string
	var posts Posts

	user := r.FormValue("user")
        burl := r.FormValue("burl")

	if len(user) == 0 {
		fmt.Fprintf(w, "Missing 'user' Parameter")
		return
	}

        if len(burl) == 0 {
                fmt.Fprintf(w, "Missing 'burl' Parameter")
                return
        }

	CleanURL(&burl)

        cmd := exec.Command(cli, "getuseraddress", user)
        ajsn, err := cmd.Output()
        if err != nil {
                fmt.Fprintf(w, "failed," + err.Error())
                return
        }

        err = json.Unmarshal(ajsn, &addr)
        if err != nil {
                fmt.Fprintf(w, "failed," + err.Error())
                return
        }

        cmd = exec.Command(cli, "getusercontents", addr[0].Address, "0", burl, "20000")
        pjsn, err := cmd.Output()
        if err != nil {
                fmt.Fprintf(w, "failed," + err.Error())
                return
        }

        err = json.Unmarshal(pjsn, &posts)
        if err != nil {
                fmt.Fprintf(w, "failed," + err.Error())
                return
        }

        for _, post := range posts.Contents {
                cmd = exec.Command(cli, "getpostscores", post.Txid)
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

		response += post.Txid + "<br>"
		response += "<table>"
                for _, rate := range rates {
			response += "<tr>"
			response += "<td>" + rate.Name + "</td><td>" + rate.Value + "</td>"
			response += "</tr>"
                }
		response += "</table>"
		response += "<br/><br/>"
        }

	fmt.Fprintf(w, response)
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	html := `
		<html>
			<head>
				<title>Tattle</title>
				<script type="text/javascript">
				function flip_uname() {
					cb = document.getElementById('bulk').checked;
					if (cb == true) {
						document.getElementById('uname').style = "display: block;";	
					} else {
						document.getElementById('uname').style = "display: none;";	
					}
				}

				function flip_mode() {
					cb = document.getElementById('bulk').checked;
					bc = document.getElementById('blocked').checked;

					if (bc == true) {
						document.getElementById('uname').style = "display: block;";
						document.getElementById('burl').style = "display: none;";
						document.getElementById('bulk').checked = false;
						document.getElementById('clicker').innerText = "Who Blocked Me?";
						document.getElementById('labeltxt').innerText = "Bastyon Name";
					} else {
						document.getElementById('uname').style = "display: none;";
						document.getElementById('burl').style = "display: block;";
						document.getElementById('bulk').checked = false;
						document.getElementById('clicker').innerText = "Get Rates";
						document.getElementById('labeltxt').innerText = "Bastyon URL";
					}
				}

				function handle_click() {
					cb = document.getElementById('bulk').checked;
					bc = document.getElementById('blocked').checked;
					url = document.getElementById('burl').value;
					btn = document.getElementById('clicker');
					waiting = document.getElementById('waiting');

					if (bc == true) {
						user = document.getElementById('uname').value;
						xhttp = new XMLHttpRequest();
						xhttp.onreadystatechange = function() {
							if (this.readyState == 4 && this.status == 200) {
								btn.style = 'display: block;';
								waiting.style = 'display: none;';
								cdiv = document.getElementById('rates');
								cdiv.innerHTML = this.responseText;
								return;
							}
						};

						btn.style = 'display: none';
						waiting.style = 'display: block;';

						encuser = encodeURIComponent(user);
						
						xhttp.open("POST", "/getblocks", true);
						xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");

						xhttp.send('user=' + encuser);

						return;
					}

					if (url == "") {
						alert("Missing URL");
						return;
					}


					if (cb == true) {
						user = document.getElementById('uname').value;
						xhttp = new XMLHttpRequest();
						xhttp.onreadystatechange = function() {
							if (this.readyState == 4 && this.status == 200) {
								cdiv = document.getElementById('rates');
								cdiv.innerHTML = this.responseText;
								return;
							}
						};

						encuser = encodeURIComponent(user);
						encurl = encodeURIComponent(url);
						
						xhttp.open("POST", "/processall", true);
						xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");

						xhttp.send('burl=' + encurl + '&user=' + encuser);
					} else {
						xhttp = new XMLHttpRequest();
						xhttp.onreadystatechange = function() {
							if (this.readyState == 4 && this.status == 200) {
								cdiv = document.getElementById('rates');
								cdiv.innerHTML = this.responseText;
								return;
							}
						};

						encurl = encodeURIComponent(url);
						
						xhttp.open("POST", "/process", true);
						xhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");

						xhttp.send('burl=' + encurl);
					}
				}

				</script>
			</head>
			<body>
				<center>
					<h1>Tattle Tale</h1>
					<input type="checkbox" id="bulk" name="bulk" value="All" onclick="javascript:flip_uname();">
					<label for="bulk">Bulk</label>
					<br/>
					<input type="checkbox" id="blocked" name="blocked" value="All" onclick="javascript:flip_mode();">
					<label for="bulk">Blocking Me</label>
					<hr/><br/>
					<b id="labeltxt">Bastyon URL</b><br/>
					<br/>
					<input type="text" id="uname" name="uname" placeholder="seditioussov" style="display: none;" /><br/>
					<input type="text" id="burl" name="burl" placeholder="https://bastyon.com/index?s=dfef303ac7c4e8c8754450bd37ad26ed7dfcecb25e844ed7b4efb49e25284b01&mpost=true" />
					<br/>
					<br/>
					<button id="clicker" onclick="javascript:handle_click()">Get Rates</button>
					<div id="waiting" style="display: none;">
						<img src="https://c.tenor.com/rec5dlPBK2cAAAAd/mr-bean-waiting.gif" />
						<br/><br/>
						This can take a while.  We are searching, and sorting the entire blockchain<br/>
						Roughly ~20s
					</div>
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

        cookie, err := ioutil.ReadFile("/home/pnet/.pocketcoin/.cookie")
        if err != nil {
                fmt.Println(err.Error())
                return
        }

        parts := strings.Split(string(cookie), ":")
        uname = parts[0]
        pass = parts[1]

        router := mux.NewRouter()
	router.HandleFunc("/process", handleProcess)
	router.HandleFunc("/processall", handleProcessAll)
	router.HandleFunc("/getblocks", handleGetBlocks)
        router.HandleFunc("/", handleMain)

        corsObj := handlers.AllowedOrigins([]string{"*"})
        err = http.ListenAndServe(":80", handlers.CORS(corsObj)(router))
        if err != nil {
                fmt.Println("ListenAndServe: " + err.Error())
        }
}
