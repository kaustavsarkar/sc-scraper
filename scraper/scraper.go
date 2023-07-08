package scraper

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	url                      = "https://main.sci.gov.in/judgments"
	judgementDayTabSel       = "div#tabbed-nav > ul.z-tabs-desktop > li[data-link=\"tab3\"]"
	activeJudgementDayTabSel = "li[data-link=\"tab3\"][class=\"z-tab z-active\"]"
	captchaTextSel           = "p#cap > font"
	captchaInputSel          = "input#ansCaptcha"
	fromDateSel              = "input#JBJfrom_date"
	toDateSel                = "input#JBJto_date"
	submitBtnSel             = "input#v_getJBJ"
	loadingSel               = "img[title=\"Loading...\"]"
	judgementTableSel        = "div#JBJ > table"
	contentDivSel            = "div#JBJ"
	basePath                 = "/home/kaustav/work/ain/sc-scraper/output"
)

type dates struct {
	startDate string
	endDate   string
}

func ScrapeAll() {
	fromDate := time.Date(1950, time.January, 01, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(1950, time.December, 31, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)
	retryListTmp := make([]*dates, 0)
	var retryList []*dates

	for !toDate.Equal(endDate) {
		fromDateStr := fromDate.Format("02-01-2006")
		toDateStr := toDate.Format("02-01-2006")
		log.Printf("startdate %s endDate %s", fromDateStr, toDateStr)
		fromDate = fromDate.AddDate(1, 0, 0)
		toDate = toDate.AddDate(1, 0, 0)
		html, err := run(fromDateStr, toDateStr)

		if err != nil {
			retryListTmp = append(retryListTmp, &dates{startDate: fromDateStr, endDate: toDateStr})
			log.Printf("there has been an error %v", err)
		}
		save(html, fromDateStr)
	}

	for len(retryListTmp) > 0 {
		log.Printf("next retry size %v", len(retryListTmp))
		retryList = make([]*dates, len(retryListTmp))
		retryList = copy(retryListTmp)
		retryListTmp = retryListTmp[:0]

		for _, date := range retryList {
			fromDateStr := date.startDate
			toDateStr := date.endDate
			log.Printf("Retry startdate %s endDate %s", fromDateStr, toDateStr)

			html, err := run(fromDateStr, toDateStr)

			if err != nil {
				retryListTmp = append(retryListTmp, &dates{startDate: fromDateStr, endDate: toDateStr})
				log.Printf("there has been an error %v", err)
			}
			save(html, fromDateStr)
		}

	}
}

func run(startDate, endDate string) (string, error) {
	// nctx, ncancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
	// defer ncancel()
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	var captcha string
	// startDate := "01-01-1950"
	// endDate := "31-12-1950"
	var judgementData string
	var buf []byte

	log.Printf("starting date %s end date %s", startDate, endDate)

	if err := chromedp.Run(ctxTimeout, chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Click(judgementDayTabSel, chromedp.ByQuery),
		chromedp.WaitVisible(activeJudgementDayTabSel)); err != nil {
		return "",
			fmt.Errorf("errored while selecting the judgement date from years %s . error: %v", startDate, err)
	}

	if err := chromedp.Run(ctxTimeout,
		chromedp.WaitVisible(captchaInputSel, chromedp.ByQuery),
		chromedp.Text(captchaTextSel, &captcha, chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.Focus(captchaInputSel)); err != nil {
		return "",
			fmt.Errorf("errored while reading captcha from years %s . error: %v", startDate, err)
	}

	if err := chromedp.Run(ctxTimeout,
		chromedp.SendKeys(captchaInputSel, captcha, chromedp.ByQuery),
		chromedp.Clear(fromDateSel, chromedp.ByQuery),
		chromedp.SendKeys(fromDateSel, startDate, chromedp.ByQuery),
		chromedp.Clear(toDateSel, chromedp.ByQuery),
		chromedp.SendKeys(toDateSel, endDate, chromedp.ByQuery)); err != nil {
		return "",
			fmt.Errorf("errored while setting captcha and dates from years %s . error: %v", startDate, err)
	}

	if err := chromedp.Run(ctxTimeout,
		chromedp.Click(submitBtnSel, chromedp.ByQuery)); err != nil {
		return "",
			fmt.Errorf("errored while clicking on submit button from years %s . error: %v", startDate, err)
	}

	if err := chromedp.Run(ctxTimeout,
		chromedp.WaitVisible(loadingSel, chromedp.ByQuery),
		chromedp.WaitVisible(judgementTableSel, chromedp.ByQuery)); err != nil {
		log.Fatal(err)
		return "",
			fmt.Errorf("errored while waiting for the data to load from years %s . error: %v", startDate, err)
	}

	if err := chromedp.Run(ctxTimeout,
		chromedp.InnerHTML(contentDivSel, &judgementData, chromedp.NodeVisible, chromedp.ByQuery)); err != nil {
		log.Printf("buffer %v", buf)
		errWr := os.WriteFile("fullScreenshot.png", buf, 0o644)
		err = errors.Join(err, errWr)
		return "",
			fmt.Errorf("errored while copying data from years %s . error: %v", startDate, err)
	}
	return judgementData, nil
}

func save(judgementData, date string) {
	folderPath := basePath + "/" + date
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		log.Printf("failed to save html for %s", date)
		return
	}
	filePath := folderPath + "/" + "raw"

	if err := ioutil.WriteFile(filePath, []byte(judgementData), 0644); // Save the file
	err != nil {
		log.Printf("Failed to save file: %v", err)
		return
	}

	log.Printf("File saved successfully: %s", filePath)
}

func copy(src []*dates) []*dates {
	dest := make([]*dates, 0)
	for _, date := range src {
		dest = append(dest, date)
	}
	log.Printf("ogt dest %v", dest)
	return dest
}
