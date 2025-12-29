package routes

import (
	"encoding/json"
	"log"

	"musick-server/internal/app/services"

	"github.com/DarthPestilane/easytcp"
)

type CreateNoteRequest struct {
	UserID      string `json:"user_id"`
	RoomID      string `json:"room_id"`
	SongID      string `json:"song_id"`
	TrackID     string `json:"track_id"`
	Step        int    `json:"step"`
	Pitch       int    `json:"pitch"`
	Velocity    int    `json:"velocity"`
	LengthSteps int    `json:"length_steps"`
}

type CreateNoteResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Note    *services.Note `json:"note,omitempty"`
}

type DeleteNoteRequest struct {
	UserID  string `json:"user_id"`
	RoomID  string `json:"room_id"`
	SongID  string `json:"song_id"`
	TrackID string `json:"track_id"`
	Step    int    `json:"step"`
	Pitch   int    `json:"pitch"`
}

type DeleteNoteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListNotesRequest struct {
	UserID  string `json:"user_id"`
	RoomID  string `json:"room_id"`
	SongID  string `json:"song_id"`
	TrackID string `json:"track_id"`
}

type ListNotesResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Notes   []services.Note  `json:"notes,omitempty"`
	Tracks  []services.Track `json:"tracks,omitempty"`
}

// NoteBroadcast is the unified payload for route 603 broadcasts.
type NoteBroadcast struct {
	Action  string         `json:"action"` // "on" for create, "off" for delete
	SongID  string         `json:"song_id"`
	TrackID string         `json:"track_id"`
	Step    int            `json:"step"`
	Pitch   int            `json:"pitch"`
	Note    *services.Note `json:"note,omitempty"`
}

// RegisterNoteRoutes wires note-related handlers.
func RegisterNoteRoutes(s *easytcp.Server) {
	s.AddRoute(601, handleCreateNote)
	s.AddRoute(602, handleDeleteNote)
	s.AddRoute(610, handleListNotes)
}

func handleCreateNote(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("601 create note: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendNoteCreateError(ctx, "not authenticated")
		return
	}

	var createReq CreateNoteRequest
	if err := json.Unmarshal(req.Data(), &createReq); err != nil {
		sendNoteCreateError(ctx, "invalid request format")
		return
	}

	if createReq.UserID == "" || createReq.RoomID == "" || createReq.SongID == "" || createReq.TrackID == "" {
		sendNoteCreateError(ctx, "user_id, room_id, song_id, and track_id are required")
		return
	}
	if createReq.Step < 0 || createReq.Pitch <= 0 {
		sendNoteCreateError(ctx, "step must be >= 0 and pitch must be > 0")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != createReq.UserID {
		sendNoteCreateError(ctx, "user_id mismatch")
		return
	}

	note, err := services.CreateNote(createReq.SongID, createReq.TrackID, createReq.Step, createReq.Pitch, createReq.Velocity, createReq.LengthSteps, createReq.UserID)
	if err != nil {
		log.Printf("failed to create note: %v", err)
		sendNoteCreateError(ctx, "failed to create note")
		return
	}

	// Ensure this session is tracked in the room for broadcasts.
	services.AddSessionToRoom(createReq.RoomID, ctx.Session())

	resp := CreateNoteResponse{
		Success: true,
		Message: "note created",
		Note:    note,
	}

	data, _ := json.Marshal(resp)

	// Broadcast the new note to all sessions in the room on route 603 with unified payload.
	bcast := NoteBroadcast{
		Action:  "on",
		SongID:  createReq.SongID,
		TrackID: createReq.TrackID,
		Step:    createReq.Step,
		Pitch:   createReq.Pitch,
		Note:    note,
	}
	if b, err := json.Marshal(bcast); err == nil {
		services.BroadcastToRoom(createReq.RoomID, easytcp.NewMessage(603, b), nil)
	}

	// Send direct response on 601.
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendNoteCreateError(ctx easytcp.Context, msg string) {
	resp := CreateNoteResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleDeleteNote(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("602 delete note: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendNoteDeleteError(ctx, "not authenticated")
		return
	}

	var delReq DeleteNoteRequest
	if err := json.Unmarshal(req.Data(), &delReq); err != nil {
		sendNoteDeleteError(ctx, "invalid request format")
		return
	}

	if delReq.UserID == "" || delReq.RoomID == "" || delReq.SongID == "" || delReq.TrackID == "" {
		sendNoteDeleteError(ctx, "user_id, room_id, song_id, and track_id are required")
		return
	}
	if delReq.Step < 0 || delReq.Pitch <= 0 {
		sendNoteDeleteError(ctx, "step must be >= 0 and pitch must be > 0")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != delReq.UserID {
		sendNoteDeleteError(ctx, "user_id mismatch")
		return
	}

	if err := services.DeleteNote(delReq.SongID, delReq.TrackID, delReq.Step, delReq.Pitch); err != nil {
		log.Printf("failed to delete note: %v", err)
		sendNoteDeleteError(ctx, "failed to delete note")
		return
	}

	// Ensure this session is tracked in the room for broadcasts.
	services.AddSessionToRoom(delReq.RoomID, ctx.Session())

	resp := DeleteNoteResponse{Success: true, Message: "note deleted"}
	data, _ := json.Marshal(resp)

	// Broadcast deletion to room on route 603 with unified payload.
	bcast := NoteBroadcast{
		Action:  "off",
		SongID:  delReq.SongID,
		TrackID: delReq.TrackID,
		Step:    delReq.Step,
		Pitch:   delReq.Pitch,
	}
	if b, err := json.Marshal(bcast); err == nil {
		services.BroadcastToRoom(delReq.RoomID, easytcp.NewMessage(603, b), nil)
	}

	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendNoteDeleteError(ctx easytcp.Context, msg string) {
	resp := DeleteNoteResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}

func handleListNotes(ctx easytcp.Context) {
	req := ctx.Request()
	log.Printf("610 list notes: id=%d bytes=%d", req.ID(), len(req.Data()))

	if !services.IsAuthenticated(ctx.Session()) {
		sendListNotesError(ctx, "not authenticated")
		return
	}

	var lnReq ListNotesRequest
	if err := json.Unmarshal(req.Data(), &lnReq); err != nil {
		sendListNotesError(ctx, "invalid request format")
		return
	}

	if lnReq.UserID == "" || lnReq.RoomID == "" || lnReq.SongID == "" {
		sendListNotesError(ctx, "user_id, room_id, and song_id are required")
		return
	}

	session := services.GetSession(ctx.Session())
	if session == nil || session.UserID != lnReq.UserID {
		sendListNotesError(ctx, "user_id mismatch")
		return
	}

	notes, err := services.ListNotesBySong(lnReq.SongID, lnReq.TrackID)
	if err != nil {
		log.Printf("failed to list notes: %v", err)
		sendListNotesError(ctx, "failed to list notes")
		return
	}

	tracks, err := services.ListTracksBySong(lnReq.SongID)
	if err != nil {
		log.Printf("failed to list tracks: %v", err)
		sendListNotesError(ctx, "failed to list tracks")
		return
	}

	// Track membership for future broadcasts.
	services.AddSessionToRoom(lnReq.RoomID, ctx.Session())

	resp := ListNotesResponse{
		Success: true,
		Message: "notes fetched",
		Notes:   notes,
		Tracks:  tracks,
	}

	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))
}

func sendListNotesError(ctx easytcp.Context, msg string) {
	resp := ListNotesResponse{Success: false, Message: msg}
	data, _ := json.Marshal(resp)
	ctx.SetResponseMessage(easytcp.NewMessage(ctx.Request().ID(), data))
}
