package services

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
)

func RecognizeSong(base64Wav string) (string, error) {
	loadEnv()
	apiKey := os.Getenv("RAPIDAPI_KEY")
	apiHost := os.Getenv("RAPIDAPI_HOST")

	if apiKey == "" || apiHost == "" {
		return "", fmt.Errorf("RapidAPI keys missing in .env")
	}

	// 1. 解碼 Base64
	wavBytes, err := base64.StdEncoding.DecodeString(base64Wav)
	if err != nil {
		return "", fmt.Errorf("Base64 解碼失敗: %v", err)
	}

	// 2. 解析 WAV Header
	var pcmBytes []byte
	numChannels := 1

	if len(wavBytes) > 44 && string(wavBytes[0:4]) == "RIFF" && string(wavBytes[8:12]) == "WAVE" {
		// 讀取聲道數
		numChannels = int(binary.LittleEndian.Uint16(wavBytes[22:24]))
		fmt.Printf("檢測到音訊聲道數: %d\n", numChannels)

		// 尋找 data chunk
		offset := 12
		found := false
		for offset+8 < len(wavBytes) {
			chunkID := string(wavBytes[offset : offset+4])
			chunkSize := binary.LittleEndian.Uint32(wavBytes[offset+4 : offset+8])

			if chunkID == "data" {
				start := offset + 8
				end := start + int(chunkSize)
				if end > len(wavBytes) {
					end = len(wavBytes)
				}
				pcmBytes = wavBytes[start:end]
				found = true
				break
			}
			offset += 8 + int(chunkSize)
		}
		if !found {
			pcmBytes = wavBytes[44:]
		}
	} else {
		pcmBytes = wavBytes
	}

	// 3. 聲道轉換邏輯優化
	if numChannels == 2 {
		fmt.Println("使用【平均法】將立體聲混合為單聲道 (提升準確度)...")

		// 預備單聲道 buffer (長度減半)
		monoData := make([]byte, len(pcmBytes)/2)

		// 遍歷 16-bit 樣本 (每 4 bytes 是一組 L+R)
		// L(low), L(high), R(low), R(high)
		for i, j := 0, 0; i+4 <= len(pcmBytes); i, j = i+4, j+2 {
			// 讀取左聲道 (Little Endian int16)
			left := int16(binary.LittleEndian.Uint16(pcmBytes[i : i+2]))
			// 讀取右聲道
			right := int16(binary.LittleEndian.Uint16(pcmBytes[i+2 : i+4]))

			// 計算平均值: (L + R) / 2
			avg := (int(left) + int(right)) / 2

			// 防爆音截斷 (雖然平均通常不會爆)
			if avg > math.MaxInt16 {
				avg = math.MaxInt16
			} else if avg < math.MinInt16 {
				avg = math.MinInt16
			}

			// 寫入單聲道 buffer
			binary.LittleEndian.PutUint16(monoData[j:j+2], uint16(avg))
		}
		pcmBytes = monoData
	}

	// 4. 重新編碼與發送
	finalPayload := base64.StdEncoding.EncodeToString(pcmBytes)

	// 嘗試使用 v1 detect 接口 (有時候對原曲辨識較準，若失敗可改回 v2)
	url := "https://shazam.p.rapidapi.com/songs/detect"

	req, err := http.NewRequest("POST", url, strings.NewReader(finalPayload))
	if err != nil {
		return "", err
	}

	req.Header.Set("content-type", "text/plain")
	req.Header.Set("X-RapidAPI-Key", apiKey)
	req.Header.Set("X-RapidAPI-Host", apiHost)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == 204 {
		return `{"matches": []}`, nil
	}

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", res.StatusCode, string(body))
	}

	return string(body), nil
}
