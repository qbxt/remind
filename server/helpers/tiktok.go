package helpers

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qbxt/gologger"
	"time"
)

type DBRow struct {
	ID int64
	MessageSID string
	Link string
	SendAt int64
	Done bool
}

func CheckForTikToksToSend() {
	gologger.Info("starting tiktok check", nil)
	db, err:= sql.Open("sqlite3", "./remind.db")
	if err != nil {
		gologger.Error("could not open db", err, nil)
		return
	}
	defer db.Close()

	now := time.Now().Unix()
	readQuery := `SELECT * FROM messages WHERE NOT(done) AND send_at < ?`

	rows, err := db.Query(readQuery, now)
	if err != nil {
		gologger.Error("could not read from db", err, nil)
		return
	}

	var completed []int64

	for rows.Next() {
		item := &DBRow{}
		if err := rows.Scan(&item.ID, &item.MessageSID, &item.Link, &item.SendAt, &item.Done); err != nil {
			gologger.Error("could not scan row", err, nil)
			continue
		}

		completed = append(completed, item.ID)

		go SendTikTok(item)
	}

	markAsDone := `UPDATE messages SET done = TRUE WHERE id = ?`

	for _, completedItem := range completed {
		if _, err := db.Exec(markAsDone, completedItem); err != nil {
			gologger.Error("could not update finished items", err, nil)
		}
	}
}
