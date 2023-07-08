package db

import (
	"database/sql"
	"encoding/json"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type JudgementLink struct {
	Link string `json:"link"`
	Date string `json:"date"`
	Lang string `json:"lang"`
}

type Judgement struct {
	DiaryNumber        string
	CaseNumber         string
	PetitionerName     string
	RespondentName     string
	PetitionerAdvocate string
	RespondentAdvocate string
	Bench              string
	JudgementBy        string
	JudgementLinks     string
}

func Open() (*sql.DB, error) {
	db, dbErr := sql.Open("sqlite3", "judgements.db")
	if dbErr != nil {
		return nil, dbErr
	}

	createTableSql := `
	CREATE TABLE IF NOT EXISTS judgements (
		DiaryNumber TEXT,
		CaseNumber TEXT,
		PetitionerName TEXT,
		RespondentName TEXT,
		PetitionerAdvocate TEXT,
		RespondentAdvocate TEXT,
		Bench TEXT,
		JudgementBy TEXT,
		JudgementLinks TEXT
	)
	`
	if _, execErr := db.Exec(createTableSql); execErr != nil {
		return nil, execErr
	}

	return db, nil
}

func Close(db *sql.DB) error {
	return db.Close()
}

func (j *Judgement) Insert(db *sql.DB, txn *sql.Tx) error {
	stmt, prepErr := txn.Prepare("INSERT INTO judgements(DiaryNumber, CaseNumber, PetitionerName, RespondentName, PetitionerAdvocate, RespondentAdvocate, Bench, JudgementBy, JudgementLinks) VALUES (?,?,?,?,?,?,?,?,?)")
	if prepErr != nil {
		return prepErr
	}
	defer stmt.Close()

	jLinkByte, _ := json.Marshal(j.JudgementLinks)

	_, execErr := stmt.Exec(j.DiaryNumber, j.CaseNumber, j.PetitionerName,
		j.RespondentName, j.PetitionerAdvocate, j.RespondentAdvocate, j.Bench,
		j.JudgementBy, string(jLinkByte))
	if execErr != nil {
		return prepErr
	}
	return nil
}

func (j *Judgement) ReadAll(db *sql.DB) ([]*Judgement, error) {
	selectQuery := "SELECT * FROM judgements"
	rows, qErr := db.Query(selectQuery)
	if qErr != nil {
		return nil, qErr
	}
	defer rows.Close()

	var judgements []*Judgement
	var readErr error

	for rows.Next() {
		var judgement Judgement
		err := rows.Scan(&judgement.DiaryNumber, &judgement.CaseNumber, &judgement.PetitionerName, &judgement.RespondentName, &judgement.PetitionerAdvocate, &judgement.RespondentAdvocate, &judgement.Bench, &judgement.JudgementBy, &judgement.JudgementLinks)

		readErr = errors.Join(readErr, err)
		judgements = append(judgements, &judgement)

	}
	return judgements, readErr
}
