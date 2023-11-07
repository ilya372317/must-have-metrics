package sender

import (
	"log"
	"net/http"
)

type ReportSender func(requestURL string)

func SendReport(requestURL string) {
	res, err := http.Post(requestURL, "text/plain", nil)

	if err != nil {
		log.Printf("failed to save data on server: %v\n", err)
		return
	}
	_ = res.Body.Close()
}
