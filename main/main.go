package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sc-scraper.com/db"
	"sc-scraper.com/filereader"
	"sc-scraper.com/scraper"
)

const root = "/home/kaustav/work/ain/sc-scraper/output"

func main() {
	// scraper.ScrapeAll()
	// crawlAndParse()
	downloadAndSave()
}

func crawlAndParse() {
	folders, _ := filereader.TraverseOutputDir(root)
	jdb, dbErr := db.Open()

	if dbErr != nil {
		log.Printf("error in opening db %v", dbErr)
		return
	}

	defer db.Close(jdb)
	// Set PRAGMA synchronous = OFF
	_, exErr := jdb.Exec("PRAGMA synchronous = OFF")
	if exErr != nil {
		log.Fatal(exErr)
	}

	// Begin a transaction
	tx, txErr := jdb.Begin()
	if txErr != nil {
		log.Fatal(txErr)
	}

	for _, folder := range folders {
		html, _ := filereader.GetHtml(root + "/" + folder + "/raw")
		judgements, _ := scraper.ParseHtml(html)

		for _, judgement := range judgements {
			err := judgement.Insert(jdb, tx)
			if err != nil {
				log.Fatalf("insert error %v", err)

			}
		}
	}

	// Commit the transaction
	err := tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func downloadAndSave() {
	jdb, dbErr := db.Open()

	if dbErr != nil {
		log.Printf("error in opening db %v", dbErr)
		return
	}
	defer db.Close(jdb)

	judgements, readErr := db.ReadAll(jdb)
	if readErr != nil {
		log.Fatal(readErr)
	}

	for _, judgement := range judgements {
		jsonString := strings.ReplaceAll(strings.Trim(judgement.JudgementLinks, `"`), "\\\"", "\"")
		var judgementLinks []db.JudgementLink
		unMarshalErr := json.Unmarshal([]byte(jsonString), &judgementLinks)
		if unMarshalErr != nil {
			log.Print([]byte(jsonString))
			log.Fatalf("err: %v string %v", unMarshalErr, jsonString)
			return
		}
		var date string
		for _, l := range judgementLinks {
			if len(l.Date) <= 0 {
				continue
			}

			hasSpace := strings.Contains(l.Date, " ")
			var dString string

			if hasSpace {
				dString = strings.Split(l.Date, " ")[0]
			} else {
				dString = l.Date
			}

			d, err := parseDate(dString)
			if err == nil {
				date = "01-01-" + fmt.Sprint(d.Year())
			}
		}

		for _, l := range judgementLinks {
			filePath := root + "/" + date + "/"
			log.Printf("%s", l.Link)
			scraper.DownloadPdf(l, filePath)
		}

	}
}

func parseDate(dateString string) (time.Time, error) {
	layout := "02-01-2006" // The expected date format, e.g., DD-MM-YYYY
	date, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
