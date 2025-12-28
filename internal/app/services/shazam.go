package services

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// RecognizeSong 呼叫 RapidAPI 的 Shazam 接口
func RecognizeSong(base64Audio string) (string, error) {
	loadEnv()
	apiKey := os.Getenv("RAPIDAPI_KEY")
	apiHost := os.Getenv("RAPIDAPI_HOST")

	url := "https://shazam.p.rapidapi.com/songs/v2/detect"

	// 這裡假設客戶端傳來的是處理好的音訊字串
	payload := strings.NewReader(base64Audio)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "text/plain")
	req.Header.Add("X-RapidAPI-Key", apiKey)
	req.Header.Add("X-RapidAPI-Host", apiHost)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	return string(body), nil
}
