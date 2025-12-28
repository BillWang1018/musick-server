package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

func main() {
	// 1. 連線到伺服器
	conn, err := net.Dial("tcp", "127.0.0.1:5896")
	if err != nil {
		fmt.Printf("【失敗】無法連線到伺服器: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("【成功】已連線至伺服器 127.0.0.1:5896")

	// 2. 準備 401 請求數據
	requestBody := map[string]string{
		"audio_data": "這是測試用的音訊數據",
	}
	jsonData, _ := json.Marshal(requestBody)

	// 3. 封裝 easytcp 格式 (Size: 4 bytes | ID: 4 bytes | Data)
	id := uint32(401)
	size := uint32(len(jsonData))

	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], size)
	binary.LittleEndian.PutUint32(header[4:8], id)

	// 4. 發送請求
	conn.Write(header)
	conn.Write(jsonData)
	fmt.Println("【發送】已送出 401 辨識請求，等待伺服器回傳中...")

	// 5. 讀取回傳 (設定 10 秒超時，避免 API 跑太久沒反應)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 先讀取回傳的標頭 (8 bytes)
	respHeader := make([]byte, 8)
	_, err = conn.Read(respHeader)
	if err != nil {
		fmt.Printf("【錯誤】讀取回傳標頭失敗: %v\n", err)
		return
	}

	// 解析回傳的資料長度
	respSize := binary.LittleEndian.Uint32(respHeader[0:4])

	// 根據長度讀取實際的 JSON 內容
	respBody := make([]byte, respSize)
	_, err = conn.Read(respBody)
	if err != nil {
		fmt.Printf("【錯誤】讀取回傳內容失敗: %v\n", err)
		return
	}

	fmt.Printf("\n===== 收到伺服器回傳結果 =====\n%s\n==============================\n", string(respBody))
}
