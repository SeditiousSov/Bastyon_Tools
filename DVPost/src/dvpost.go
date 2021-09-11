package main

import (
	"os"
	"fmt"
	"time"
	"strings"
	"context"

	"os/signal"
	"io/ioutil"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/page"
)

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
	var burls []string

	if len(os.Args) < 2 {
		fmt.Println("Usage: " + os.Args[0] + " <bastyon_url> <bastyon_url>")
		fmt.Println("Usage: " + os.Args[0] + " <file_with_urls>")
		fmt.Println("ie: " + os.Args[0] + " https://bastyon.com/post?s=d9c1033f9d727fba2854980364712ca2223f47cab929da1b666e7d4fae7e6d34")
		fmt.Println("ie: " + os.Args[0] + " /home/user/list.txt")
		return
	}

        loginkey = os.Getenv("BASTYON_KEY") 

        signalChannel := make(chan os.Signal, 2)
        signal.Notify(signalChannel, os.Interrupt)
        go func() {
                for {
                        <-signalChannel
                        os.Exit(1)
                }
        }()

	if _, err := os.Stat(os.Args[1]); err == nil {
		content, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			burls = append(burls, line)
		}
	} else {
		burls = os.Args[1:]
	}

	err := DVPost(burls)
	if err != nil {
		fmt.Println(err.Error())
	}
}
