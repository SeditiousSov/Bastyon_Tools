package main

import (
	"os"
	"fmt"
	"bytes"
	"strings"
	"strconv"

	"net/http"
	"io/ioutil"
	"encoding/json"
)

/*
This is ten percent luck, twenty percent skill
Fifteen percent concentrated power of will
*/

type NodeList struct {
	Result []struct {
		Time     int    `json:"time"`
		Services int    `json:"services"`
		Address  string `json:"address"`
		Port     int    `json:"port"`
	} `json:"result"`
	Error interface{} `json:"error"`
	ID    string      `json:"id"`
}

type PeerInfo struct {
	Result []struct {
		ID              int           `json:"id"`
		Addr            string        `json:"addr"`
		Addrlocal       string        `json:"addrlocal"`
		Addrbind        string        `json:"addrbind"`
		Services        string        `json:"services"`
		Relaytxes       bool          `json:"relaytxes"`
		Lastsend        int           `json:"lastsend"`
		Lastrecv        int           `json:"lastrecv"`
		Bytessent       int           `json:"bytessent"`
		Bytesrecv       int           `json:"bytesrecv"`
		Conntime        int           `json:"conntime"`
		Timeoffset      int           `json:"timeoffset"`
		Pingtime        float64       `json:"pingtime"`
		Minping         float64       `json:"minping"`
		Version         int           `json:"version"`
		Subver          string        `json:"subver"`
		Inbound         bool          `json:"inbound"`
		Addnode         bool          `json:"addnode"`
		Startingheight  int           `json:"startingheight"`
		Banscore        int           `json:"banscore"`
		SyncedHeaders   int           `json:"synced_headers"`
		SyncedBlocks    int           `json:"synced_blocks"`
		Inflight        []interface{} `json:"inflight"`
		Whitelisted     bool          `json:"whitelisted"`
		Minfeefilter    float64       `json:"minfeefilter"`
	} `json:"result"`
	Error interface{} `json:"error"`
	ID    string      `json:"id"`
}


type PostsEX struct {
	Result string `json:"result"`
	Data   struct {
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
				Image  string      `json:"image"`
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
		} `json:"contents"`
		Videos struct {
		} `json:"videos"`
		Users []struct {
			Address    string  `json:"address"`
			Name       string  `json:"name"`
			ID         int     `json:"id"`
			I          string  `json:"i"`
			B          string  `json:"b"`
			R          string  `json:"r"`
			Reputation float64 `json:"reputation"`
			Postcnt    int     `json:"postcnt"`
			Rc         int     `json:"rc"`
		} `json:"users"`
	} `json:"data"`
}

var uname string
var pass string

func GetAllNodes() ([]string, error) {
	var nodes NodeList
	var retnodes []string

	jstring := `{"jsonrpc": "1.0", "id":"curltest", "method":"getnodeaddresses", "params": [10000]}`

	body := bytes.NewReader([]byte(jstring))
	req, err := http.NewRequest("POST", "http://localhost:37071", body)
	if err != nil {
		return retnodes, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
        req.SetBasicAuth(uname, pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return retnodes, err
	}
	defer resp.Body.Close()

	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return retnodes, err
	}

	err = json.Unmarshal(bod, &nodes)
	if err != nil {
		return retnodes, err
	}

	for _, node := range nodes.Result {
		strport := strconv.Itoa(node.Port)
		con := node.Address + ":" + strport
		retnodes = append(retnodes, con)
	}

	return retnodes, nil
}

func GetPeerInfo() ([]string, error) {
	var nodes PeerInfo 
	var retnodes []string

	jstring := `{"jsonrpc": "1.0", "id":"curltest", "method":"getpeerinfo", "params": []}`

	body := bytes.NewReader([]byte(jstring))
	req, err := http.NewRequest("POST", "http://localhost:37071", body)
	if err != nil {
		return retnodes, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
        req.SetBasicAuth(uname, pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return retnodes, err
	}
	defer resp.Body.Close()

	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return retnodes, err
	}

	err = json.Unmarshal(bod, &nodes)
	if err != nil {
		return retnodes, err
	}

	for _, node := range nodes.Result {
		retnodes = append(retnodes, node.Addr)
	}

	return retnodes, nil
}

func GetHierarchialStrip(node string) (PostsEX, error) {
	var posts PostsEX

	jstring := `{"parameters":[0,"",10,"en",[],""],"method":"gethierarchicalstrip","cachehash":"f9cf2d7a34b4423af82c9dc2672ae06b","options":{"ex":true,"node":"` + node + `"}}`

	body := bytes.NewReader([]byte(jstring))
	req, err := http.NewRequest("POST", "https://pocketnet.app:8899/rpc-ex/gethierarchialstrip/", body)
	if err != nil {
		return posts, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return posts, err
	}
	defer resp.Body.Close()

	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return posts, err
	}

	err = json.Unmarshal(bod, &posts)
	if err != nil {
		return posts, err
	}

	return posts, nil
}

func main() {
	var ptxid string
	var pheight string
	var nodedata = make(map[string]PostsEX)
	var mctxid = make(map[string]int)
	var mcheight = make(map[string]int)
	var assclowns = make(map[string]string)


	// Delete From Here (If You Defined rcpuser and password in the conf)
        cookie, err := ioutil.ReadFile("/home/pnet/.pocketcoin/.cookie")
        if err != nil {
                fmt.Println(err.Error())
                return
        }

        parts := strings.Split(string(cookie), ":")
        uname = parts[0]
        pass = parts[1]
	// To Here

	// Uncomment Below If you have an rpc user and password
	// uname = "your_rpc_user"
	// pass = "your_rpc_password"




	// Get As Many Node IPs As Possible
	peerlist, err := GetPeerInfo()
	chkerror(err)

	nodelist, err := GetAllNodes()
	chkerror(err)



	// Get Historical Strip For Height And TXID Of Remote Nodes
	for _, node := range nodelist {
		pex, err := GetHierarchialStrip(node)
		if err != nil {
			continue
		}

		nodedata[node] = pex
	}

	for _, peer := range peerlist {
		pex, err := GetHierarchialStrip(peer)
		if err != nil {
			continue
		}

		nodedata[peer] = pex
	}




	// Get Most Common Height and Txid
	for _, v := range nodedata {
		h := strconv.Itoa(v.Data.Height)
		mcheight[h]++
		mctxid[v.Data.Contents[0].Txid]++
	}

	// Everyone Has The Same Height, And Txid
	if len(mcheight) == 1 && len(mctxid) == 1 {
		fmt.Println("All Scanned Nodes Are Good")
		return
	}

	// Find The Height That Most Nodes Agree Is The Current Height
	for _, v := range mcheight {
		for key, val := range mcheight {
			if val >= pheight {
				pheight = key
			}
		}
	}

	// Find The TXID That Most Nodes Agree Is The Current TXID
	for _, v := range mctxid {
		for key, val := range mctxid {
			if val >= pheight {
				pheight = key
			}
		}
	}

	iHeight, err := strconv.Atoi(pheight)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Check For Any Rogue Or Unagreeable Nodes
	for k, v := range nodedata {
		if v.Data.Height < iHeight || v.Data.Contents[0].Txid != ptxid {
			sHeight := strconv.Itoa(v.Data.Height)
			stringval := `Node ` + k + ` Is Suspect.  Height: (` + sHeight + `/` + pheight + `) Txid: ` + v.Data.Contents[0].Txid + `/` + ptxid
			assclowns[k] = stringval
		}
	}


	// If AssClowns Is 0, Then The Node That Was Different On Height Or Txid 
	// Probably Has More Recent Info Then The Consensus
	if len(assclowns) == 0 {
		fmt.Println("All Scanned Nodes Are Good")
		return
	}


	// We've Got Suspect Nodes.  Probably Check Just How Far Off They Are
	for _, v := range assclowns {
		fmt.Println(v)
	}

	return
	for k, v := range nodedata {
		fmt.Print(k + " - ")
		fmt.Print(v.Data.Height)
		fmt.Print(" - ")
		fmt.Print(v.Data.Contents[0].Txid)
		fmt.Print(" - ")
		fmt.Println(v.Data.Contents[0].Time)
	}
}

func chkerror(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
