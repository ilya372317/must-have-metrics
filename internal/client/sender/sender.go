package sender

import (
	"bytes"
	"net/http"

	"github.com/ilya372317/must-have-metrics/internal/logger"
	"github.com/ilya372317/must-have-metrics/internal/utils/compress"
)

const failedSaveDataErrPattern = "failed to save data on server: %v\n"

type ReportSender func(requestURL, body string)

func SendReport(requestURL, body string) {
	compressedData, errCompress := compress.Do([]byte(body))
	if errCompress != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, errCompress)
	}
	request, errRequest := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(compressedData))
	if errRequest != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, errRequest)
		return
	}
	request.Header.Set("Content-Encoding", "gzip")

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		logger.Log.Errorf(failedSaveDataErrPattern, err)
		return
	}
	_ = res.Body.Close()
}
