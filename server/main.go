package main

import (
	"./routes"
	"./helpers"
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qbxt/gologger"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	gologger.Init()
	rand.Seed(time.Now().UnixNano())
	db, err := sql.Open("sqlite3", "./remind.db")
	if err != nil {
		gologger.Fatal("error opening db", err, nil)
	}

	statement := `CREATE TABLE IF NOT EXISTS messages (
    	id INTEGER PRIMARY KEY, 
    	messageSid TEXT, 
    	link TEXT, 
    	send_at INTEGER,
    	done BOOLEAN
	);`

	if _, err := db.Exec(statement); err != nil {
		gologger.Error("could not create table", err, nil)
	}

	go helpers.CheckForTikToksToSend()
	c := cron.New()
	_, _ = c.AddFunc("@every 30s", helpers.CheckForTikToksToSend)
	c.Start()

	r := mux.NewRouter()

	r.HandleFunc("/sms", routes.Sms)

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	switch os.Getenv("REMIND_SERVER_TYPE") {
	case "PROD", "PRODUCTION":
		gologger.Info("server is running!", logrus.Fields{"env": "staging"})
		if err := srv.ListenAndServeTLS("./certs/cert.pem", "./certs/key.pem"); err != nil {
			gologger.Error("server error", err, nil)
		}
		break
	default:
		gologger.Info("server is running!", logrus.Fields{"env": "staging"})
		if err := srv.ListenAndServe(); err != nil {
			gologger.Error("server error", err, nil)
		}
		break
	}
}
