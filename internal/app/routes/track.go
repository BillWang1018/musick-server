package routes

import (
	"encoding/json"
	"log"

	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

type CreateTrackRequest struct {
	UserID     string `json:"user_id"`
	RoomID     string `json:"room_id"`
	SongID     string `json:"song_id"`
	Name       string `json:"name"`
	Instrument string `json:"instrument"`
	Channel    *int   `json:"channel,omitempty"`
	Color      string `json:"color"`
}

type CreateTrackResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Track   *services.Track `json:"track,omitempty"`
}

type DeleteTrackRequest struct {
	UserID  string `json:"user_id"`
	RoomID  string `json:"room_id"`
	SongID  string `json:"song_id"`
	TrackID string `json:"track_id"`
}

type DeleteTrackResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type TrackBroadcast struct {
	Action  string          `json:"action"` // "on" for add, "off" for delete
	Track   *services.Track `json:"track,omitempty"`
	TrackID string          `json:"track_id,omitempty"`
	SongID  string          `json:"song_id,omitempty"`
}

// RegisterTrackRoutes wires track-related handlers.
func RegisterTrackRoutes(s *easytcp.Server) {
	s.AddRoute(604, handleCreateTrack)
	s.AddRoute(605, handleDeleteTrack)
}

func handleCreateTrack(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("604 create track: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendCreateTrackError(ctx, "not authenticated")
		return
	}

	var tReq CreateTrackRequest
	if err := json.Unmarshal(req.Data(), &tReq); err != nil {
		sendCreateTrackError(ctx, "invalid request format")
		return
	}

	if tReq.UserID == "" || tReq.RoomID == "" || tReq.SongID == "" || tReq.Name == "" {
		sendCreateTrackError(ctx, "user_id, room_id, song_id, and name are required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != tReq.UserID {
		sendCreateTrackError(ctx, "user_id mismatch")
		return
	}

	track, err := services.CreateTrack(tReq.SongID, tReq.Name, tReq.Instrument, tReq.Channel, tReq.Color)
	if err != nil {
		log.Printf("failed to create track: %v", err)
		sendCreateTrackError(ctx, "failed to create track")
		return
	}

	// Track membership for broadcasts.
	services.AddSessionToRoom(tReq.RoomID, ctx.Session())

	resp := CreateTrackResponse{Success: true, Message: "track created", Track: track}
	data, _ := json.Marshal(resp)

	// Broadcast add on route 606.
	bcast := TrackBroadcast{Action: "on", Track: track, TrackID: track.ID, SongID: track.SongID}
	if b, err := json.Marshal(bcast); err == nil {
		services.BroadcastToRoom(tReq.RoomID, easytcp.NewMessage(606, b), nil)
	}

	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func handleDeleteTrack(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("605 delete track: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendDeleteTrackError(ctx, "not authenticated")
		return
	}

	var dReq DeleteTrackRequest
	if err := json.Unmarshal(req.Data(), &dReq); err != nil {
		sendDeleteTrackError(ctx, "invalid request format")
		return
	}

	if dReq.UserID == "" || dReq.RoomID == "" || dReq.SongID == "" || dReq.TrackID == "" {
		sendDeleteTrackError(ctx, "user_id, room_id, song_id, and track_id are required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != dReq.UserID {
		sendDeleteTrackError(ctx, "user_id mismatch")
		return
	}

	if err := services.DeleteTrack(dReq.TrackID, dReq.SongID); err != nil {
		log.Printf("failed to delete track: %v", err)
		sendDeleteTrackError(ctx, "failed to delete track")
		return
	}

	services.AddSessionToRoom(dReq.RoomID, ctx.Session())

	resp := DeleteTrackResponse{Success: true, Message: "track deleted"}
	data, _ := json.Marshal(resp)

	bcast := TrackBroadcast{Action: "off", TrackID: dReq.TrackID, SongID: dReq.SongID}
	if b, err := json.Marshal(bcast); err == nil {
		services.BroadcastToRoom(dReq.RoomID, easytcp.NewMessage(606, b), nil)
	}

	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendCreateTrackError(ctx easytcp.Context, msg string) {
	resp := CreateTrackResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func sendDeleteTrackError(ctx easytcp.Context, msg string) {
	resp := DeleteTrackResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}
