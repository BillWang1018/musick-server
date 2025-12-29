package routes

import (
	"encoding/json"
	"log"

	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

type CreateRoomRequest struct {
	UserID    string `json:"user_id"`
	RoomName  string `json:"room_name"`
	IsPrivate bool   `json:"is_private"`
}

type CreateRoomResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id,omitempty"`
	RoomCode  string `json:"room_code,omitempty"`
	RoomName  string `json:"room_name,omitempty"`
	IsPrivate bool   `json:"is_private,omitempty"`
}

type ListRoomsRequest struct {
	UserID string `json:"user_id"`
}

type ListRoomsResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Rooms   []services.Room `json:"rooms,omitempty"`
}

type FindPublicRoomsRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type FindPublicRoomsResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Rooms   []services.Room `json:"rooms,omitempty"`
}

func RegisterRoomRoutes(s *easytcp.Server) {
	s.AddRoute(201, handleCreateRoom)
	s.AddRoute(210, handleListRooms)
	s.AddRoute(211, handleFindPublicRooms)
}

func handleCreateRoom(ctx easytcp.Context) {
	req := ctx.Request()

	// Check authentication
	if !services.IsAuthenticated(ctx.Session()) {
		sendRoomError(ctx, "not authenticated")
		return
	}

	var createReq CreateRoomRequest
	if err := json.Unmarshal(req.Data(), &createReq); err != nil {
		sendRoomError(ctx, "invalid request format")
		return
	}

	// Validate required fields
	if createReq.RoomName == "" {
		sendRoomError(ctx, "room_name is required")
		return
	}

	if createReq.UserID == "" {
		sendRoomError(ctx, "user_id is required")
		return
	}

	// Verify the user_id matches the authenticated session
	userSession := services.GetSession(ctx.Session())
	if userSession == nil || userSession.UserID != createReq.UserID {
		sendRoomError(ctx, "user_id mismatch")
		return
	}

	// Create room in database
	room, err := services.CreateRoom(createReq.UserID, createReq.RoomName, createReq.IsPrivate)
	if err != nil {
		log.Printf("failed to create room: %v", err)
		sendRoomError(ctx, "failed to create room")
		return
	}

	log.Printf("room created: %s (code: %s) by user %s", room.Title, room.Code, room.OwnerID)

	resp := CreateRoomResponse{
		Success:   true,
		Message:   "room created successfully",
		RoomID:    room.ID,
		RoomCode:  room.Code,
		RoomName:  room.Title,
		IsPrivate: room.IsPrivate,
	}

	respData, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), respData))
}

func sendRoomError(ctx easytcp.Context, msg string) {
	resp := CreateRoomResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleListRooms(ctx easytcp.Context) {
	req := ctx.Request()

	if !services.IsAuthenticated(ctx.Session()) {
		sendListRoomsError(ctx, "not authenticated")
		return
	}

	var listReq ListRoomsRequest
	if err := json.Unmarshal(req.Data(), &listReq); err != nil {
		sendListRoomsError(ctx, "invalid request format")
		return
	}

	if listReq.UserID == "" {
		sendListRoomsError(ctx, "user_id is required")
		return
	}

	userSession := services.GetSession(ctx.Session())
	if userSession == nil || userSession.UserID != listReq.UserID {
		sendListRoomsError(ctx, "user_id mismatch")
		return
	}

	rooms, err := services.ListRoomsByUser(listReq.UserID)
	if err != nil {
		log.Printf("failed to list rooms: %v", err)
		sendListRoomsError(ctx, "failed to list rooms")
		return
	}

	resp := ListRoomsResponse{
		Success: true,
		Message: "rooms fetched",
		Rooms:   rooms,
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendListRoomsError(ctx easytcp.Context, msg string) {
	resp := ListRoomsResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleFindPublicRooms(ctx easytcp.Context) {
	req := ctx.Request()

	if !services.IsAuthenticated(ctx.Session()) {
		sendFindPublicRoomsError(ctx, "not authenticated")
		return
	}

	var findReq FindPublicRoomsRequest
	if err := json.Unmarshal(req.Data(), &findReq); err != nil {
		sendFindPublicRoomsError(ctx, "invalid request format")
		return
	}

	if findReq.UserID == "" {
		sendFindPublicRoomsError(ctx, "user_id is required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != findReq.UserID {
		sendFindPublicRoomsError(ctx, "user_id mismatch")
		return
	}

	rooms, err := services.FindPublicRooms(findReq.Name)
	if err != nil {
		log.Printf("failed to find public rooms: %v", err)
		sendFindPublicRoomsError(ctx, "failed to find public rooms")
		return
	}

	resp := FindPublicRoomsResponse{
		Success: true,
		Message: "rooms fetched",
		Rooms:   rooms,
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendFindPublicRoomsError(ctx easytcp.Context, msg string) {
	resp := FindPublicRoomsResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}
