package main

import (
	"os"
	"fmt"
	"context"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/page"
)

func PostMessage(msg string, tags []string) (error) {
        loginkey := os.Getenv("BASTYON_KEY") 

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
                            chromedp.Navigate("https://bastyon.com/authorization"),
			    chromedp.Sleep(3 * time.Second),
                            chromedp.WaitVisible(`#authorization > div.scrollmaketpart > div.maket.leftPart > div > div > div > div > div.form > div.formh > div:nth-child(1) > div > div:nth-child(1) > div > input`, chromedp.ByQuery),
                            chromedp.SendKeys(`#authorization > div.scrollmaketpart > div.maket.leftPart > div > div > div > div > div.form > div.formh > div:nth-child(1) > div > div:nth-child(1) > div > input`, loginkey, chromedp.ByQuery),
                            chromedp.Click(`#authorization > div.scrollmaketpart > div.maket.leftPart > div > div > div > div > div.form > div.formh > div.item.e.buttons > button`, chromedp.NodeVisible),
                            chromedp.Sleep(3 * time.Second),
			    chromedp.WaitVisible(`#sortableBody > div:nth-child(1) > div > div.draggablecontent > div.txt > div.emojionearea.message > div.emojionearea-editor.pastable`, chromedp.ByQuery),
			    chromedp.SendKeys(`#sortableBody > div:nth-child(1) > div > div.draggablecontent > div.txt > div.emojionearea.message > div.emojionearea-editor.pastable`, msg, chromedp.ByQuery),
			    chromedp.Sleep(1 * time.Second),
			    chromedp.ActionFunc(func(c context.Context) error {
				    for _, tag := range tags {
					err := chromedp.SendKeys(`#taginput > div > div > div.tagsInputWrapper > div > div > div > div.searchInput > div.searchInputIcon > div.searchInputWrapper > input`, tag + " ", chromedp.NodeVisible).Do(c)
					if err != nil {
						return err   
					}

					chromedp.Sleep(500 * time.Millisecond).Do(c)
				    }

				    return nil
			    }),
			    chromedp.Sleep(1 * time.Second),
			    chromedp.Click(`#share > div.absolutetxwrapper > div:nth-child(4) > div > div.postlineWrapper > div > div.postWrapper > div > div > button`, chromedp.NodeVisible),
			    chromedp.Sleep(10 * time.Second),
        )
        
        if err != nil {
                return err
        }

	return nil
}

func main() {
	tags := os.Args[2:]
	err := PostMessage(os.Args[1], tags)
	if err != nil {
		fmt.Println(err.Error())
	}
}
