package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

const url = "https://main.sci.gov.in/judgments"

func main() {
	nctx, ncancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
	defer ncancel()
	ctx, cancel := chromedp.NewContext(nctx)
	defer cancel()
	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	var captcha string
	startDate := "01-01-1950"
	endDate := "31-12-1950"
	// var judgementData string
	var buf []byte

	log.Printf("starting date %s end date %s", startDate, endDate)

	if err := chromedp.Run(ctxTimeout, chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		clickOnJudgementDay(),
		chromedp.WaitVisible("li[data-link=\"tab3\"][class=\"z-tab z-active\"]")); err != nil {
		log.Fatal(err)
	}

	if err := chromedp.Run(ctxTimeout, chromedp.WaitVisible("input#ansCaptcha", chromedp.ByQuery),
		readCaptccha(&captcha),
		chromedp.Focus("input#ansCaptcha")); err != nil {
		log.Printf("buffer %v", buf)
		if err := os.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	if err := chromedp.Run(ctxTimeout, provideCaptcha(&captcha)); err != nil {
		log.Printf("buffer %v", buf)
		if err := os.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
			log.Fatal(err)
		}
		log.Fatal(err)
	}

	// chromedp.Clear("input#JBJfrom_date", chromedp.ByQuery),
	// selectStartDate(&startDate),
	// chromedp.Clear("input#JBJto_date", chromedp.ByQuery),
	// selectEndDate(&endDate),
	// waitForLoadingToAppear(),
	// waitForLoadingToDisappear(),
	// chromedp.CaptureScreenshot(&buf),
	// copyJudgementDiv(&judgementData)

	fmt.Print(captcha)

}

func clickOnJudgementDay() chromedp.Action {
	log.Print("Click on judgement day")
	return chromedp.Click("div#tabbed-nav > ul.z-tabs-desktop > li[data-link=\"tab3\"]", chromedp.ByQuery)
}

func readCaptccha(captcha *string) chromedp.Action {
	log.Print("Reading captcha")
	return chromedp.Text("p#cap > font", captcha, chromedp.NodeVisible, chromedp.ByQuery)
}

func provideCaptcha(captcha *string) chromedp.Action {
	log.Printf("Read captcha %s", *captcha)
	defer time.Sleep(20 * time.Second)
	return chromedp.SendKeys("input#ansCaptcha", *captcha, chromedp.ByQuery)
}

func selectStartDate(date *string) chromedp.Action {
	log.Printf("Select start time %s", *date)
	return chromedp.SendKeys("input#JBJfrom_date", *date, chromedp.ByQuery)
}

func selectEndDate(date *string) chromedp.Action {
	log.Printf("Select end time %s", *date)
	return chromedp.SendKeys("input#JBJto_date", *date, chromedp.ByQuery)
}

func clickOnSubmit() chromedp.Action {
	log.Printf("Click on submit")
	return chromedp.Click("input#v_getJBJ", chromedp.ByQuery)
}

func waitForLoadingToAppear() chromedp.Action {
	log.Print("Wait for loading to start")
	return chromedp.WaitVisible("img[title=\"Loading...\"]", chromedp.ByQuery)
}

func waitForLoadingToDisappear() chromedp.Action {
	log.Print("Wait for loading to end")
	return chromedp.WaitNotVisible("img[title=\"Loading...\"]", chromedp.ByQuery)
}

func copyJudgementDiv(judgementData *string) chromedp.Action {
	log.Print("Copying judgements")
	return chromedp.InnerHTML("div#JBJ", judgementData, chromedp.NodeVisible, chromedp.ByQuery)
}
