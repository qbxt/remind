package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/qbxt/gologger"
	"math/rand"
	"os"
	"time"
)

type TwilioSMS struct {
	MessageSid string `json:"messageSid"`
	Sender string `json:"from"`
	Receiver string `json:"to"`
	Body string `json:"body"`
	DateCreated string `json:"date_created"`
}

func SendConfirmation(smsObj *TwilioSMS) {
	twilioSID := os.Getenv("REMIND_TWILIO_SID")
	twilioAuthToken := os.Getenv("REMIND_TWILIO_AUTH_TOKEN")

	_, _, _ = gorequest.New().SetBasicAuth(twilioSID, twilioAuthToken).
		Post(fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", twilioSID)).
		Type(gorequest.TypeUrlencoded).
		Send(fmt.Sprintf("From=%s", smsObj.Receiver)).
		Send(fmt.Sprintf("To=%s", smsObj.Sender)).
		Send("Body=👍 Got it! I'll remind you about this soon!").
		End()
}

func SendError(smsObj *TwilioSMS, badMessage string) {
	twilioSID := os.Getenv("REMIND_TWILIO_SID")
	twilioAuthToken := os.Getenv("REMIND_TWILIO_AUTH_TOKEN")

	_, _, _ = gorequest.New().SetBasicAuth(twilioSID, twilioAuthToken).
		Post(fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", twilioSID)).
		Type(gorequest.TypeUrlencoded).
		Send(fmt.Sprintf("From=%s", smsObj.Receiver)).
		Send(fmt.Sprintf("To=%s", smsObj.Sender)).
		Send(fmt.Sprintf("Body=☠️ Uh-oh! Something went wrong over here (%s). \n\nYou can try again in a few minutes, or email remind@queue.bot if the problem persists", badMessage)).
		End()
}

func SendTikTok(row *DBRow) {
	gologger.Info("sending tiktok!", nil)
	twilioSID := os.Getenv("REMIND_TWILIO_SID")
	twilioAuthToken := os.Getenv("REMIND_TWILIO_AUTH_TOKEN")

	_, body, errs := gorequest.New().SetBasicAuth(twilioSID, twilioAuthToken).
		Get(fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages/%s.json", twilioSID, row.MessageSID)).
		End()
	if len(errs) != 0 {
		for _, err := range errs {
			gologger.Error("could not get twilio message", err, nil)
		}
		return
	}

	sms := &TwilioSMS{}
	if err := json.Unmarshal([]byte(body), sms); err != nil {
		gologger.Error("could not encode to struct", err, nil)
		return
	}

	sentAt, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", sms.DateCreated)
	if err != nil {
		gologger.Error("could not parse time", err, nil)
		return
	}

	sinceSent := time.Since(sentAt)
	sinceSentString := shortString(sinceSent)

	greetings := []string{"Hi", "Heya", "Hello", "Greetings", "Howdy", "Yo", "Ahoy"}
	greeting := greetings[rand.Intn(len(greetings))]

	_, _, _ = gorequest.New().SetBasicAuth(twilioSID, twilioAuthToken).
		Post(fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", twilioSID)).
		Type(gorequest.TypeUrlencoded).
		Send(fmt.Sprintf("From=%s", sms.Receiver)).
		Send(fmt.Sprintf("To=%s", sms.Sender)).
		Send(fmt.Sprintf("Body=👋 %s! I'm reminding you about this TikTok from %s ago: %s\n\n", greeting, sinceSentString, row.Link)).
		End()
}

func shortString(d time.Duration) string {
	day := 24 * time.Hour
	week := 7 * day
	if d.Hours() >= week.Hours() {
		dHours := int(d.Hours())
		weekHours := int(week.Hours())
		count := dHours % weekHours
		descriptor := ""
		if count > 1 {
			descriptor = "week"
		} else {
			descriptor = "weeks"
		}
		return fmt.Sprintf("%d %s", count, descriptor)
	}

	if d.Hours() >= day.Hours() {
		dHours := int(d.Hours())
		dayHours := int(day.Hours())
		count := dHours % dayHours
		descriptor := ""
		if count > 1 {
			descriptor = "hour"
		} else {
			descriptor = "hours"
		}
		return fmt.Sprintf("%d %s", count, descriptor)
	}

	if d.Hours() >= time.Hour.Hours() {
		count := int(d.Hours())
		descriptor := ""
		if count > 1 {
			descriptor = "hour"
		} else {
			descriptor = "hours"
		}
		return fmt.Sprintf("%d %s", count, descriptor)
	}

	return "less than an hour"
}
