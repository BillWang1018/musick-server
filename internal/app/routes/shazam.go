package routes

import (
	"encoding/json"
	"log"
	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

type ShazamRequest struct {
	AudioData string `json:"audio_data"` // 假設客戶端傳送 Base64 或原始數據
}

type ShazamResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Result  interface{} `json:"result,omitempty"`
}

func RegisterShazamRoutes(s *easytcp.Server) {
	s.AddRoute(401, handleShazamDetect)
}

func handleShazamDetect(ctx easytcp.Context) {
	req := ctx.Request()

	// 1. 檢查是否登入 (按照現有專案慣例)
	if !services.IsAuthenticated(ctx.Session()) {
		sendShazamError(ctx, "未驗證身份")
		return
	}

	// 2. 解析 JSON 請求
	var sReq ShazamRequest
	if err := json.Unmarshal(req.Data(), &sReq); err != nil {
		sendShazamError(ctx, "無效的請求格式")
		return
	}

	// 3. 呼叫服務層向 RapidAPI 請求
	resultJson, err := services.RecognizeSong(sReq.AudioData)
	if err != nil {
		log.Printf("Shazam API 錯誤: %v", err)
		sendShazamError(ctx, "辨識失敗")
		return
	}

	// 4. 回傳成功結果
	var rawResult map[string]interface{}
	json.Unmarshal([]byte(resultJson), &rawResult)

	resp := ShazamResponse{
		Success: true,
		Message: "辨識成功",
		Result:  rawResult,
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendShazamError(ctx easytcp.Context, msg string) {
	resp := ShazamResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}
