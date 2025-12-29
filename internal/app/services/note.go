package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Note represents a grid note tied to a song/track.
type Note struct {
	ID          string    `json:"id"`
	SongID      string    `json:"song_id"`
	TrackID     string    `json:"track_id"`
	Step        int       `json:"step"`
	Pitch       int       `json:"pitch"`
	Velocity    int       `json:"velocity"`
	LengthSteps int       `json:"length_steps"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateNote inserts a new note row and returns it.
func CreateNote(songID, trackID string, step, pitch, velocity, lengthSteps int, userID string) (*Note, error) {
	loadEnv()

	if velocity <= 0 {
		velocity = 100
	}
	if lengthSteps <= 0 {
		lengthSteps = 1
	}
	if step < 0 {
		return nil, fmt.Errorf("step must be non-negative")
	}
	if pitch <= 0 {
		return nil, fmt.Errorf("pitch must be positive")
	}

	payload := map[string]interface{}{
		"song_id":      songID,
		"track_id":     trackID,
		"step":         step,
		"pitch":        pitch,
		"velocity":     velocity,
		"length_steps": lengthSteps,
		"created_by":   userID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal note payload: %w", err)
	}

	req, _ := http.NewRequest("POST", supabaseURL+"/rest/v1/notes", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create note failed (status %d): %s", resp.StatusCode, respBody)
	}

	var rows []Note
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode note response: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("create note returned no rows")
	}

	return &rows[0], nil
}

// DeleteNote removes a note by unique coordinates.
func DeleteNote(songID, trackID string, step, pitch int) error {
	loadEnv()

	if step < 0 {
		return fmt.Errorf("step must be non-negative")
	}
	if pitch <= 0 {
		return fmt.Errorf("pitch must be positive")
	}

	url := fmt.Sprintf("%s/rest/v1/notes?song_id=eq.%s&track_id=eq.%s&step=eq.%d&pitch=eq.%d", supabaseURL, songID, trackID, step, pitch)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete note failed (status %d): %s", resp.StatusCode, respBody)
	}

	return nil
}

// ListNotesBySong fetches all notes for a song (optionally filtered by track).
func ListNotesBySong(songID, trackID string) ([]Note, error) {
	loadEnv()

	q := url.Values{}
	q.Set("song_id", "eq."+songID)
	q.Set("select", "id,song_id,track_id,step,pitch,velocity,length_steps,created_by,created_at")
	q.Set("order", "step.asc,pitch.asc")
	if trackID != "" {
		q.Set("track_id", "eq."+trackID)
	}

	endpoint := fmt.Sprintf("%s/rest/v1/notes?%s", supabaseURL, q.Encode())
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch notes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch notes failed (status %d): %s", resp.StatusCode, respBody)
	}

	var notes []Note
	if err := json.NewDecoder(resp.Body).Decode(&notes); err != nil {
		return nil, fmt.Errorf("decode notes: %w", err)
	}

	return notes, nil
}
