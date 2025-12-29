package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Song represents a song tied to a room.
type Song struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"room_id"`
	Title           string    `json:"title"`
	BPM             int       `json:"bpm"`
	Steps           int       `json:"steps"`
	BeatsPerMeasure int       `json:"beats_per_measure"`
	Scale           string    `json:"scale"`
	StartPitch      int       `json:"start_pitch"`
	OctaveRange     int       `json:"octave_range"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}

// ListSongsByRoom fetches songs for a given room from Supabase.
func ListSongsByRoom(roomID string) ([]Song, error) {
	loadEnv()

	q := url.Values{}
	q.Set("select", "id,room_id,title,bpm,steps,beats_per_measure,scale,start_pitch,octave_range,created_by,created_at")
	q.Set("room_id", "eq."+roomID)
	q.Set("order", "created_at.asc")

	endpoint := fmt.Sprintf("%s/rest/v1/songs?%s", supabaseURL, q.Encode())
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch songs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch songs failed (status %d): %s", resp.StatusCode, bodyBytes)
	}

	var songs []Song
	if err := json.NewDecoder(resp.Body).Decode(&songs); err != nil {
		return nil, fmt.Errorf("decode songs: %w", err)
	}

	return songs, nil
}

// CreateSong inserts a new song row and returns it.
func CreateSong(roomID, title string, bpm, steps int, userID string) (*Song, error) {
	loadEnv()

	if bpm <= 0 {
		bpm = 120
	}

	if steps <= 0 {
		steps = 64
	}

	payload := map[string]interface{}{
		"room_id":    roomID,
		"title":      title,
		"bpm":        bpm,
		"steps":      steps,
		"created_by": userID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal song payload: %w", err)
	}

	req, _ := http.NewRequest("POST", supabaseURL+"/rest/v1/songs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create song: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create song failed (status %d): %s", resp.StatusCode, respBody)
	}

	var rows []Song
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode song response: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("create song returned no rows")
	}

	return &rows[0], nil
}

// UpdateSong updates song metadata and returns the updated row.
func UpdateSong(songID string, title *string, bpm *int, steps *int, beatsPerMeasure *int, scale *string, startPitch *int, octaveRange *int) (*Song, error) {
	loadEnv()

	if songID == "" {
		return nil, fmt.Errorf("song_id is required")
	}

	payload := map[string]interface{}{}

	if title != nil {
		if *title == "" {
			return nil, fmt.Errorf("title cannot be empty")
		}
		payload["title"] = *title
	}

	if bpm != nil {
		if *bpm <= 0 {
			return nil, fmt.Errorf("bpm must be positive")
		}
		payload["bpm"] = *bpm
	}

	if steps != nil {
		if *steps <= 0 {
			return nil, fmt.Errorf("steps must be positive")
		}
		payload["steps"] = *steps
	}

	if beatsPerMeasure != nil {
		if *beatsPerMeasure <= 0 {
			return nil, fmt.Errorf("beats_per_measure must be positive")
		}
		payload["beats_per_measure"] = *beatsPerMeasure
	}

	if scale != nil {
		val := strings.ToLower(strings.TrimSpace(*scale))
		if val != "major" && val != "minor" {
			return nil, fmt.Errorf("scale must be 'major' or 'minor'")
		}
		payload["scale"] = val
	}

	if startPitch != nil {
		if *startPitch < 0 || *startPitch > 127 {
			return nil, fmt.Errorf("start_pitch must be between 0 and 127")
		}
		payload["start_pitch"] = *startPitch
	}

	if octaveRange != nil {
		if *octaveRange <= 0 {
			return nil, fmt.Errorf("octave_range must be positive")
		}
		payload["octave_range"] = *octaveRange
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal song update payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/songs?id=eq.%s", supabaseURL, songID)
	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update song: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update song failed (status %d): %s", resp.StatusCode, respBody)
	}

	var rows []Song
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode song update response: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("update song returned no rows")
	}

	return &rows[0], nil
}
