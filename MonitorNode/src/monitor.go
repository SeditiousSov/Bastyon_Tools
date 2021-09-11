package main

import(
	"os"
	"fmt"
	"errors"
	"strings"

	"net/http"
	"io/ioutil"
)

func GetLatestBlock() (string, error) {
	resp, err := http.Get("http://pnet.codefaction.net:8181/getheight")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bod)), nil
}

func GetLastFileBlock(fpath string) (string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	buf := make([]byte, 128)
	stat, err := os.Stat(fpath)
	start := stat.Size() - 128 
	_, err = file.ReadAt(buf, start)
	if err != nil {
		return "", err
	} 

	parts := strings.Split(string(buf), " ")
	if strings.Contains(string(buf), "BestHeader:") {
		return parts[6], nil
	} else {
		return "", errors.New("Missing Best Header As Last Line In Debug.log")
	} 

	return "", nil

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " <bastion_debug.log>")
		return
	}

	if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
		fmt.Println("Can't Find File: " + os.Args[1])
		return
	}

	block, err := GetLatestBlock()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fblock, err := GetLastFileBlock(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if block == fblock {
		fmt.Println("all good")
	} else {
		fmt.Println("not good")
	}
}
