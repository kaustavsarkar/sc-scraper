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
		clickOnJudgementDay(),
		readCaptccha(&data)); err != nil {
		log.Fatal(err)
	}

	fmt.Print(data)

}

func clickOnJudgementDay() chromedp.Action {
	return chromedp.Click("div#tabbed-nav > ul.z-tabs-desktop > li[data-link=\"tab3\"]", chromedp.ByQuery)
}

func readCaptccha(captcha *string) chromedp.Action {
	return chromedp.Text("p#cap > font", captcha, chromedp.NodeVisible, chromedp.ByQuery)
}
