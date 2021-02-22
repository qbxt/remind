package routes

import (
	"../helpers"
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/parnurzeal/gorequest"
	"github.com/qbxt/gologger"
	"github.com/sirupsen/logrus"
	str2duration "github.com/xhit/go-str2duration"
	"net/http"
	"strings"
	"time"
)

func Sms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		gologger.Info("received SMS", nil)
		if err := r.ParseForm(); err != nil {
			gologger.Error("could not parse form", err, nil)
			return
		}
		form := r.Form
		req := &helpers.TwilioSMS{
			MessageSid: form["MessageSid"][0],
			Sender: form["From"][0],
			Receiver: form["To"][0],
			Body: form["Body"][0],
		}

		/* if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			gologger.Error("could not decode request", err, nil)
			w.WriteHeader(http.StatusInternalServerError)
			helpers.SendError(req, "DECODE_REQ_ERR")
			return
		} */

		bodyArgs := strings.Split(req.Body, " ")
		if len(bodyArgs) != 2 { // URL, time string
			gologger.Warn("missing one or more arguments", nil, nil)
			helpers.SendError(req, "MISSING_ARGUMENTS")
			return
		}

		ok, err := resolveTikTokUrl(bodyArgs[0])
		if !ok {
			gologger.Warn("bad tiktok url", nil, logrus.Fields{"url": bodyArgs[0]})
			helpers.SendError(req, "INVALID_LINK")
			return
		} else if err != nil {
			gologger.Warn("error resolving TikTok URL", err, nil)
			helpers.SendError(req, "COULD_NOT_RESOLVE_LINK")
			return
		}

		delta, err := str2duration.ParseDuration(bodyArgs[1])
		if err != nil {
			gologger.Warn("could not parse time", err, nil)
			helpers.SendError(req, "COULD_NOT_PARSE_TIME")
			return
		}
		now := time.Now().Unix()
		nowAndDelta := now + int64(delta.Seconds())

		addItem := `INSERT OR REPLACE INTO messages (
			messageSid, link, send_at
		) VALUES(?, ?, ?)`

		db, err := sql.Open("sqlite3", "./remind.db")
		if err != nil {
			gologger.Error("could not open db", err, nil)
			helpers.SendError(req, "DB_ERROR")
			return
		}

		stmt, err := db.Prepare(addItem)
		if err != nil {
			gologger.Error("statement error", err, nil)
			helpers.SendError(req, "DB_ERROR")
			return
		}
		if _, err := stmt.Exec(req.MessageSid, bodyArgs[0], nowAndDelta); err != nil {
			gologger.Error("could not save to db", err, nil)
			helpers.SendError(req, "DB_ERROR")
			return
		}

		_ = stmt.Close()

		go helpers.SendConfirmation(req)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("Method not allowed"))
	}
}

func resolveTikTokUrl(urlLocation string) (bool, error) {
	if !strings.HasPrefix(urlLocation, "https://vm.tiktok.com/") {
		return false, nil
	}

	resp, _, errs := gorequest.New().Get(urlLocation).End()
	if len(errs) != 0 {
		for _, err := range errs {
			gologger.Warn("error fetching tiktok URL", err, logrus.Fields{"url": urlLocation})
		}
		return false, errors.New("could not fetch URL, please check console")
	}

	// valid vm.tiktok.com links return {"rip": "m.tiktok.com"} https://queue.bot/t/FD7mTXuP.png
	// invalid links return {"rip": "www.tiktok.com"} https://queue.bot/t/GR1iMtL2.png
	for _, h := range resp.Header {
		if h[0] == "m.tiktok.com" {
			return true, nil
		}
	}
	return false, nil
}
