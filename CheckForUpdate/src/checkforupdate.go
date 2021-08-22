package main

import (
	"os"
	"fmt"
	"strings"

	"os/exec"
	"net/http"
	"io/ioutil"
	"crypto/tls"
	"encoding/json"

        "gopkg.in/yaml.v2"
	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	gomail "gopkg.in/mail.v2"
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
}

type Config struct {
	Use_Twilio bool `yaml:"use_twilio"`
	Twilio_Account_Sid string `yaml:"twilio_account_sid"`
	Twilio_Auth_Token string `yaml:"twilio_auth_token"`
	Twilio_Phone_Number string `yaml:"twilio_phone_number"`
	Destination_Phone_Number string `yaml:"destination_phone_number"`

	Use_Email bool `yaml:"use_email"`
	Email_Server string `yaml:"email_server"`
	Email_Port int `yaml:"email_port"`
	Email_From string `yaml:"email_from"`
	Email_To string `yaml:"email_to"`
	Email_Server_Password string `yaml:"email_server_password"`

	BastyonCLI string `yaml:"bastyon_cli"`
	Use_Action bool `yaml:"use_action"`
	Action string `yaml:"action"`
	Action_Params []string `yaml:"action_params"`
}

var config Config
var cli string
var notification string

func ForwardSMSTwilio(phonenumber string, bdy string) {
	os.Setenv("TWILIO_ACCOUNT_SID", config.Twilio_Account_Sid)
	os.Setenv("TWILIO_AUTH_TOKEN", config.Twilio_Auth_Token)
	os.Setenv("TWILIO_PHONE_NUMBER", config.Twilio_Phone_Number)

	client := twilio.NewRestClient()

	params := &openapi.CreateMessageParams{}
	params.SetTo(phonenumber)
	params.SetFrom(config.Twilio_Phone_Number)
	params.SetBody(bdy)

	_, err := client.ApiV2010.CreateMessage(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func SendEmail() (error) {
	m := gomail.NewMessage()

	m.SetHeader("From", config.Email_From)
	m.SetHeader("To", config.Email_To)
	m.SetHeader("Subject", "New Bastyon Node Version Available")
	
	m.SetBody("text/plain", notification)

	d := gomail.NewDialer(config.Email_Server, config.Email_Port, config.Email_From, config.Email_Server_Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

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

	return version, nil
}


func main() {
	var peerinfo []PeerInfo

	// Load Config
        etcflag := false
        b, err := ioutil.ReadFile("/etc/bastyon/checkforupdate.yml")
        if err != nil {
                etcflag = true
        }

        if etcflag == true {
                b, err = ioutil.ReadFile("checkforupdate.yml")
                if err != nil {
                        fmt.Println("Error Reading Config File From Two Locations (/etc/bastyon/checkforupdate.yml) And (Current Directory): " + err.Error())
                        return
                }
        }

        if len(b) == 0 {
                fmt.Println("Failed To Load A Config!")
                return
        }

        yml := string(b)

        err = yaml.Unmarshal([]byte(yml), &config)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cli = config.BastyonCLI

	cver, err := GetCurrentRelease()
	if err != nil {
		fmt.Println(err.Error())
		return 
	}

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

	notification = `A new version of Bastyon Node is available (` + ver.Major + `.` + ver.Minor + `.` + ver.Patch + `).  Your version is: ` + cver.Major + `.` + cver.Minor + `.` + cver.Patch 

	if cver.Major < ver.Major {
		if config.Use_Twilio {
			ForwardSMSTwilio(config.Destination_Phone_Number, notification)
		}

		if config.Use_Email {
			SendEmail()
		}

		if config.Use_Action {
			if _, err := os.Stat(config.Action); err == nil {
				cmd := exec.Command(config.Action, config.Action_Params...)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				fmt.Println(output)
				return
			}
		}

		return
	}

	if cver.Minor < ver.Minor {
		if config.Use_Twilio {
			ForwardSMSTwilio(config.Destination_Phone_Number, notification)
		}

		if config.Use_Email {
			SendEmail()
		}

		if config.Use_Action {
			if _, err := os.Stat(config.Action); err == nil {
				cmd := exec.Command(config.Action, config.Action_Params...)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				fmt.Println(output)
				return
			}
		}

		return
	}

	if cver.Patch < ver.Patch {
		if config.Use_Twilio {
			ForwardSMSTwilio(config.Destination_Phone_Number, notification)
		}

		if config.Use_Email {
			SendEmail()
		}

		if config.Use_Action {
			if _, err := os.Stat(config.Action); err == nil {
				cmd := exec.Command(config.Action, config.Action_Params...)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				fmt.Println(output)
				return
			}
		}

		return
	}
}
