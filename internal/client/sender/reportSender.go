package sender

import (
	"log"
	"net/http"
)

type ReportSender func(requestURL string)

func SendReport(requestURL string) {
	res, err := http.Post(requestURL, "text/plain", nil)
	defer res.Body.Close()
	if err != nil {
		log.Printf("failed to save data on server: %v\n", err)
	}

}
