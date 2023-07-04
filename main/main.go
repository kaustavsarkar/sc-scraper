package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"
)

const url = "https://main.sci.gov.in/judgments"

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var data string

	if err := chromedp.Run(ctx, chromedp.Navigate(url),
		readCaptccha(&data)); err != nil {
		log.Fatal(err)
	}

	fmt.Print(data)

}

func readCaptccha(captcha *string) chromedp.Action {
	return chromedp.Text("p#cap > font", captcha, chromedp.NodeVisible, chromedp.ByQuery)
}
