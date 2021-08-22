package main

import (
	"fmt"
	"strings"
	"strconv"

	"os/exec"
	"net/http"
	"io/ioutil"
	"encoding/json"

)

type PeerInfo struct {
	Addr string `json:"addr"`
	Subver string `json:"subver"`
	BanScore int `json:"banscore"`
	SyncedHeaders int `json:"synced_headers"`
	SyncedBlocks int `json:"synced_blocks"`
}

type Version struct {
	Major string
	Minor string
	Patch string

	IMajor int
	IMinor int
	IPatch int
}

var cli string

func GetLatestRelease() (Version, error) {
	var version Version

	client := &http.Client {
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get("https://github.com/pocketnetteam/pocketnet.core/releases/latest")
	if err != nil {
		return version, err
	}
	defer resp.Body.Close()

	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return version, err
	}

	parts := strings.Split(string(bod), "=")
	uparts := strings.Split(parts[1], ">")
	vparts := strings.Split(uparts[0], "/")

	ver := vparts[len(vparts) - 1]
	ver = strings.Replace(ver, "\"", "", -1)
	ver = strings.Replace(ver, "v", "", -1)

	vp := strings.Split(ver, ".")

	version.Major = vp[0]
	version.Minor = vp[1]
	version.Patch = vp[2]

	version.IMajor, err = strconv.Atoi(vp[0])
	if err != nil {
		return version, err
	}

	version.IMinor, err = strconv.Atoi(vp[1])
	if err != nil {
		return version, err
	}

	version.IPatch, err = strconv.Atoi(vp[2])
	if err != nil {
		return version, err
	}

	return version, nil
}

func GetCurrentRelease() (Version, error) {
	var version Version

	cmd := exec.Command(cli, "--version")
	raw, err := cmd.Output()
	if err != nil {
		return version, err
	}

	parts := strings.Split(string(raw), " ")
	vraw := parts[len(parts) - 1]
	vparts := strings.Split(vraw, "-")
	
	ver := vparts[0]
	ver = strings.Replace(ver, "v", "", -1)
	bts := strings.Split(ver, ".")

	version.Major = bts[0]
	version.Minor = bts[1]
	version.Patch = bts[2]

	version.IMajor, err = strconv.Atoi(bts[0])
	if err != nil {
		return version, err
	}

	version.IMinor, err = strconv.Atoi(bts[1])
	if err != nil {
		return version, err
	}

	version.IPatch, err = strconv.Atoi(bts[2])
	if err != nil {
		return version, err
	}

	return version, nil
}


func main() {
	var peerinfo []PeerInfo

	versions_behind := 2
	ban_time := "604800"

	cli = `/usr/local/bin/pocketcoin-cli`

	cmd := exec.Command(cli, "getpeerinfo")
	jsn, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = json.Unmarshal(jsn, &peerinfo)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ver, err := GetLatestRelease()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	_ = ver

	for _, pinfo := range peerinfo {
		tmpver := Version{}
		pinfo.Subver = strings.Replace(pinfo.Subver, "/", "", -1)
		parts := strings.Split(pinfo.Subver, ":")
		verraw := parts[1]
		vparts := strings.Split(verraw, ".")

		tmpver.Major = vparts[0]
		tmpver.Minor = vparts[1]
		tmpver.Patch = vparts[2]

		tmpver.IMajor, err = strconv.Atoi(vparts[0])
		if err != nil {
			return 
		}

		tmpver.IMinor, err = strconv.Atoi(vparts[1])
		if err != nil {
			return 
		}

		tmpver.IPatch, err = strconv.Atoi(vparts[2])
		if err != nil {
			return 
		}


		if tmpver.IMajor < ver.IMajor {
			cmd := exec.Command(cli, "setban", pinfo.Addr, "add", ban_time)
			output, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			continue
		}

		if tmpver.IMinor < ver.IMinor {
			cmd := exec.Command(cli, "setban", pinfo.Addr, "add", ban_time)
			output, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			continue
		}

		if tmpver.IPatch < (ver.IPatch - versions_behind) {
			cmd := exec.Command(cli, "setban", pinfo.Addr, "add", ban_time)
			output, err := cmd.Output()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			continue
		}
		
	}
}
