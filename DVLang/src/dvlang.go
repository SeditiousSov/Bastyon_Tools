package main

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"strings"
	"context"

	"net/url"
	"net/http"
	"os/signal"
	"io/ioutil"
	"encoding/json"

	"github.com/chromedp/chromedp"
	"github.com/abadojack/whatlanggo"
	"github.com/chromedp/cdproto/page"
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
var loginkey string

func DVPost(burls []string) (error) {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", true))...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	const script = `(function(w, n, wn) {
	  // Pass the Webdriver Test.
	  Object.defineProperty(n, 'webdriver', {
	    get: () => false,
	  });

	  // Pass the Plugins Length Test.
	  // Overwrite the plugins property to use a custom getter.
	  Object.defineProperty(n, 'plugins', {
	    // This just needs to have length > 0 for the current test,
	    // but we could mock the plugins too if necessary.
	    get: () => [1, 2, 3, 4, 5],
	  });

	  // Pass the Languages Test.
	  // Overwrite the plugins property to use a custom getter.
	  Object.defineProperty(n, 'languages', {
	    get: () => ['en-US', 'en'],
	  });

	  // Pass the Chrome Test.
	  // We can mock this in as much depth as we need for the test.
	  w.chrome = {
	    runtime: {},
	  };

	  // Pass the Permissions Test.
	  const originalQuery = wn.permissions.query;
	  return wn.permissions.query = (parameters) => (
	    parameters.name === 'notifications' ?
	      Promise.resolve({ state: Notification.permission }) :
	      originalQuery(parameters)
	  );

	})(window, navigator, window.navigator);`

	var scriptID page.ScriptIdentifier
        err := chromedp.Run(ctx, 
	    chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		scriptID, err = page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
		if err != nil {
			return err
		}

		return nil
	    }),

	    // Login
	    chromedp.Navigate("https://bastyon.com/authorization"),
	    chromedp.Sleep(3 * time.Second),
	    chromedp.WaitVisible(`#authorization > div.scrollmaketpart > div.maket.leftPart > div > div > div > div > div.form > div.formh > div:nth-child(1) > div > div:nth-child(1) > div > input`, chromedp.ByQuery),
	    chromedp.SendKeys(`#authorization > div.scrollmaketpart > div.maket.leftPart > div > div > div > div > div.form > div.formh > div:nth-child(1) > div > div:nth-child(1) > div > input`, loginkey, chromedp.ByQuery),
	    chromedp.Click(`#authorization > div.scrollmaketpart > div.maket.leftPart > div > div > div > div > div.form > div.formh > div.item.e.buttons > button`, chromedp.NodeVisible),

	    chromedp.ActionFunc(func(c context.Context) error {
		    for _, burl := range burls {
			    // The Down Vote
			    chromedp.Sleep(3 * time.Second).Do(c)
			    chromedp.Navigate(burl).Do(c)
			    chromedp.WaitVisible(`#post > div > div > div > div.work.unselectablemobile > div.shareTable.post.truerepost > div:nth-child(5) > div > div.panel.unselectablemobile > div.item.forstars > div > div.starswr > div > div.starsWrapper.starsWrapperM > div:nth-child(1) > i`, chromedp.ByQuery).Do(c)
			    chromedp.Sleep(3 * time.Second).Do(c)
			    chromedp.Click(`#post > div > div > div > div.work.unselectablemobile > div.shareTable.post.truerepost > div:nth-child(5) > div > div.panel.unselectablemobile > div.item.forstars > div > div.starswr > div > div.starsWrapper.starsWrapperM > div:nth-child(1)`, chromedp.ByQuery).Do(c)
			    chromedp.Click(`#post > div > div > div > div.work.unselectablemobile > div.shareTable.post.truerepost > div:nth-child(5) > div > div.panel.unselectablemobile > div.item.forstars > div > div.starswr > div > div.starsWrapper.starsWrapperM > div:nth-child(1) > i`, chromedp.ByQuery).Do(c)
			    chromedp.Sleep(8 * time.Second).Do(c)
		    }

		    return nil
	    }),

        )
        
        if err != nil {
                return err
        }

	return nil
}

func main() {
//	var burls []string
        var posts Post

        loginkey = os.Getenv("BASTYON_KEY") 
        cookie := os.Getenv("BASTYON_PASSWORD")
	rpcurl := os.Getenv("NODE_URL")

	if len(loginkey) == 0 {
		fmt.Println("Missing BASTYON_KEY Environment Variable")
		return
	}

	if len(cookie) == 0 {
		fmt.Println("Missing BASTYON_PASSWORD Environment Variable")
		return
	}

	if len(rpcurl) == 0 {
		fmt.Println("Missing NODE_URL Environment Variable")
		fmt.Println("ie: http://127.0.0.1:37071")
		return
	}

        parts := strings.Split(string(cookie), ":")
        uname = parts[0]
        pass = parts[1]

        signalChannel := make(chan os.Signal, 2)
        signal.Notify(signalChannel, os.Interrupt)
        go func() {
                for {
                        <-signalChannel
                        os.Exit(1)
                }
        }()

        rpccmd := `{"jsonrpc": "1.0", "id":"curltest", "method":"gethistoricalstrip", "params": [0, "", 300, "en"] }`
        body := bytes.NewReader([]byte(rpccmd))

        req, err := http.NewRequest("POST", rpcurl, body)
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

	var openlist []string

        for _, post := range posts.Result.Contents {
		decoded, err := url.QueryUnescape(post.M)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		if decoded == "" {
			continue
		}

		info := whatlanggo.Detect(decoded)
		if info.Lang.String() == "Russian" {
			dvurl := "https://bastyon.com/post?s=" + post.Txid
			openlist = append(openlist, dvurl)
		}
	}

	if len(openlist) > 0 {
		err := DVPost(openlist)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}
