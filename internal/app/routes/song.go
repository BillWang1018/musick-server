package routes

import (
	"encoding/json"
	"log"

	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

type SongListRequest struct {
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
}

type SongListResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Songs   []services.Song `json:"songs,omitempty"`
}

type CreateSongRequest struct {
	UserID string `json:"user_id"`
	RoomID string `json:"room_id"`
	Title  string `json:"title"`
	BPM    int    `json:"bpm"`
	Steps  int    `json:"steps"`
}

type CreateSongResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Song    *services.Song `json:"song,omitempty"`
}

type UpdateSongRequest struct {
	UserID string  `json:"user_id"`
	RoomID string  `json:"room_id"`
	SongID string  `json:"song_id"`
	Title  *string `json:"title,omitempty"`
	BPM    *int    `json:"bpm,omitempty"`
	Steps  *int    `json:"steps,omitempty"`
}

type UpdateSongResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Song    *services.Song `json:"song,omitempty"`
}

// RegisterSongRoutes wires song-related handlers.
func RegisterSongRoutes(s *easytcp.Server) {
	s.AddRoute(501, handleCreateSong)
	s.AddRoute(510, handleListSongs)
	s.AddRoute(511, handleUpdateSong)
}

func handleListSongs(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("510 list songs: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendSongListError(ctx, "not authenticated")
		return
	}

	var listReq SongListRequest
	if err := json.Unmarshal(req.Data(), &listReq); err != nil {
		sendSongListError(ctx, "invalid request format")
		return
	}

	if listReq.UserID == "" || listReq.RoomID == "" {
		sendSongListError(ctx, "user_id and room_id are required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != listReq.UserID {
		sendSongListError(ctx, "user_id mismatch")
		return
	}

	songs, err := services.ListSongsByRoom(listReq.RoomID)
	if err != nil {
		log.Printf("failed to list songs: %v", err)
		sendSongListError(ctx, "failed to list songs")
		return
	}

	resp := SongListResponse{
		Success: true,
		Message: "songs fetched",
		Songs:   songs,
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendSongListError(ctx easytcp.Context, msg string) {
	resp := SongListResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleCreateSong(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("501 create song: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendSongCreateError(ctx, "not authenticated")
		return
	}

	var createReq CreateSongRequest
	if err := json.Unmarshal(req.Data(), &createReq); err != nil {
		sendSongCreateError(ctx, "invalid request format")
		return
	}

	if createReq.UserID == "" || createReq.RoomID == "" || createReq.Title == "" {
		sendSongCreateError(ctx, "user_id, room_id, and title are required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != createReq.UserID {
		sendSongCreateError(ctx, "user_id mismatch")
		return
	}

	song, err := services.CreateSong(createReq.RoomID, createReq.Title, createReq.BPM, createReq.Steps, createReq.UserID)
	if err != nil {
		log.Printf("failed to create song: %v", err)
		sendSongCreateError(ctx, "failed to create song")
		return
	}

	resp := CreateSongResponse{
		Success: true,
		Message: "song created",
		Song:    song,
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendSongCreateError(ctx easytcp.Context, msg string) {
	resp := CreateSongResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleUpdateSong(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("511 update song: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendSongUpdateError(ctx, "not authenticated")
		return
	}

	var upReq UpdateSongRequest
	if err := json.Unmarshal(req.Data(), &upReq); err != nil {
		sendSongUpdateError(ctx, "invalid request format")
		return
	}

	if upReq.UserID == "" || upReq.RoomID == "" || upReq.SongID == "" {
		sendSongUpdateError(ctx, "user_id, room_id, and song_id are required")
		return
	}

	if upReq.Title == nil && upReq.BPM == nil && upReq.Steps == nil {
		sendSongUpdateError(ctx, "no fields to update")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != upReq.UserID {
		sendSongUpdateError(ctx, "user_id mismatch")
		return
	}

	updated, err := services.UpdateSong(upReq.SongID, upReq.Title, upReq.BPM, upReq.Steps)
	if err != nil {
		log.Printf("failed to update song: %v", err)
		sendSongUpdateError(ctx, "failed to update song")
		return
	}

	resp := UpdateSongResponse{
		Success: true,
		Message: "song updated",
		Song:    updated,
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendSongUpdateError(ctx easytcp.Context, msg string) {
	resp := UpdateSongResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}
