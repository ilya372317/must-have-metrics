package sender

import (
	"log"
	"net/http"
)

type ReportSender func(requestURL string)

func SendReport(requestURL string) {
	if _, err := http.Post(requestURL, "text/plain", nil); err != nil {
		log.Printf("failed to save data on server: %v\n", err)
	}
}
