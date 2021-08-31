package main

import (
	"os"
	"fmt"
	"sync"
	"bytes"
	"strings"

	"os/exec"
	"net/http"
	"io/ioutil"
	"encoding/json"
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

type Addr struct {
        Name string `json:"name"`
        Address string `json:"address"`
}

type Payload struct {
	Jsonrpc string   `json:"jsonrpc"`
	ID      string   `json:"id"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

var cli string
var blockedc []BlockedConcurrent 
var uname string
var pass string

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

func GetUserProfile(address string) ([]Blocked, error) {
	var blocking []Blocked

	cmd := exec.Command(cli, "getuserprofile", `"` + address + `"`)
	jsn, err := cmd.Output()
	if err != nil {
		return blocking, err 
	}

	err = json.Unmarshal(jsn, &blocking)
	if err != nil {
		return blocking, err
	}

	return blocking, nil
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

func main() {
	var names []string

	wg := sync.WaitGroup{}

	cli = `/usr/local/bin/pocketcoin-cli`

	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " \"<bastyon_name>\"")
		return
	}

	uaddr, err := GetUserAddress(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	_ = uaddr

	addrs, err := GetAllAddresses() 
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cookie, err := ioutil.ReadFile("/home/pnet/.pocketcoin/.cookie")
	if err != nil {
		fmt.Println(err.Error())
		return 
	}
	
	parts := strings.Split(string(cookie), ":")
	uname = parts[0]
	pass = parts[1]
	completed := 0

	for i := 0; i < len(addrs); i += 4 {
		if len(addrs) - completed >= 4 {
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
		fmt.Println(name)
	}
}
