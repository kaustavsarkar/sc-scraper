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
		chromedp.OuterHTML("html", &data, chromedp.ByQuery)); err != nil {
		log.Fatal(err)
	}

	fmt.Print(data)

}
