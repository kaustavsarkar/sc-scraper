package main

import (
	"log"

	"sc-scraper.com/db"
	"sc-scraper.com/filereader"
	"sc-scraper.com/scraper"
)

const root = "/home/kaustav/work/ain/sc-scraper/output"

func main() {
	// scraper.ScrapeAll()
	// crawlAndParse()
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
