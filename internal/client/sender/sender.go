package sender

import (
	"log"
	"net/http"
	"strings"
)

type ReportSender func(requestURL, body string)

func SendReport(requestURL, body string) {
	res, err := http.Post(requestURL, "text/plain", strings.NewReader(body))

	if err != nil {
		log.Printf("failed to save data on server: %v\n", err)
		return
	}
	_ = res.Body.Close()
}
