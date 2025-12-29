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

// Track represents a song track/instrument lane.
type Track struct {
	ID         string    `json:"id"`
	SongID     string    `json:"song_id"`
	Name       string    `json:"name"`
	Instrument string    `json:"instrument"`
	Channel    *int      `json:"channel,omitempty"`
	Color      string    `json:"color"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateTrack inserts a new track and returns it.
func CreateTrack(songID, name, instrument string, channel *int, color string) (*Track, error) {
	loadEnv()

	if songID == "" || name == "" {
		return nil, fmt.Errorf("song_id and name are required")
	}

	payload := map[string]interface{}{
		"song_id":    songID,
		"name":       name,
		"instrument": instrument,
		"color":      color,
	}
	if channel != nil {
		payload["channel"] = *channel
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal track payload: %w", err)
	}

	req, _ := http.NewRequest("POST", supabaseURL+"/rest/v1/tracks", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create track: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create track failed (status %d): %s", resp.StatusCode, respBody)
	}

	var rows []Track
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode track response: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("create track returned no rows")
	}

	return &rows[0], nil
}

// DeleteTrack removes a track by id (and optional song guard).
func DeleteTrack(trackID, songID string) error {
	loadEnv()

	if trackID == "" {
		return fmt.Errorf("track_id is required")
	}

	url := fmt.Sprintf("%s/rest/v1/tracks?id=eq.%s", supabaseURL, trackID)
	if songID != "" {
		url += fmt.Sprintf("&song_id=eq.%s", songID)
	}

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete track: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete track failed (status %d): %s", resp.StatusCode, respBody)
	}

	return nil
}

// ListTracksBySong returns all tracks for a song ordered by created_at.
func ListTracksBySong(songID string) ([]Track, error) {
	loadEnv()

	if songID == "" {
		return nil, fmt.Errorf("song_id is required")
	}

	q := url.Values{}
	q.Set("song_id", "eq."+songID)
	q.Set("select", "id,song_id,name,instrument,channel,color,created_at")
	q.Set("order", "created_at.asc")

	endpoint := fmt.Sprintf("%s/rest/v1/tracks?%s", supabaseURL, q.Encode())
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch tracks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch tracks failed (status %d): %s", resp.StatusCode, respBody)
	}

	var tracks []Track
	if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
		return nil, fmt.Errorf("decode tracks: %w", err)
	}

	return tracks, nil
}
