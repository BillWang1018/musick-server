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

// CommunityPost models the community_posts row.
type CommunityPost struct {
	ID         string    `json:"id"`
	AuthorID   string    `json:"author_id"`
	AuthorName string    `json:"author_name,omitempty"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CommunityAttachment models community_post_attachments rows.
type CommunityAttachment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	FilePath  string    `json:"file_path"`
	FileType  string    `json:"file_type"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

// CommunityPostWithAttachments keeps optional attachments for a post.
type CommunityPostWithAttachments struct {
	CommunityPost
	Attachments []CommunityAttachment `json:"community_post_attachments,omitempty"`
}

// CreateCommunityPost inserts a new post authored by user.
func CreateCommunityPost(authorID, title, body string) (*CommunityPost, error) {
	loadEnv()

	payload := map[string]interface{}{
		"author_id": authorID,
		"title":     title,
		"body":      body,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal post payload: %w", err)
	}

	req, _ := http.NewRequest("POST", supabaseURL+"/rest/v1/community_posts", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create post failed (status %d): %s", resp.StatusCode, bodyBytes)
	}

	var rows []CommunityPost
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode post response: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("create post returned no rows")
	}

	post := &rows[0]
	if name, err := fetchSenderName(authorID); err == nil {
		post.AuthorName = name
	}

	return post, nil
}

// DeleteCommunityPost deletes a post by id scoped to author.
func DeleteCommunityPost(postID, authorID string) error {
	loadEnv()

	url := fmt.Sprintf("%s/rest/v1/community_posts?id=eq.%s&author_id=eq.%s", supabaseURL, postID, authorID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete post failed (status %d): %s", resp.StatusCode, bodyBytes)
	}

	return nil
}

// UpdateCommunityPost updates title/body and returns the updated row.
func UpdateCommunityPost(postID, authorID string, title *string, body *string) (*CommunityPost, error) {
	loadEnv()

	payload := map[string]interface{}{}
	if title != nil {
		payload["title"] = *title
	}
	if body != nil {
		payload["body"] = *body
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Touch updated_at on update.
	payload["updated_at"] = time.Now().UTC()

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal update payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/community_posts?id=eq.%s&author_id=eq.%s", supabaseURL, postID, authorID)
	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update post failed (status %d): %s", resp.StatusCode, bodyBytes)
	}

	var rows []CommunityPost
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decode update response: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("update post returned no rows")
	}

	post := &rows[0]
	if name, err := fetchSenderName(authorID); err == nil {
		post.AuthorName = name
	}

	return post, nil
}

// ListCommunityPosts returns posts in reverse chronological order with optional attachments.
func ListCommunityPosts(before string, limit int, includeAttachments bool) ([]CommunityPostWithAttachments, bool, error) {
	loadEnv()

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	selectClause := "id,author_id,title,body,created_at,updated_at"
	if includeAttachments {
		selectClause += ",community_post_attachments(id,post_id,file_path,file_type,mime_type,created_at)"
	}

	q := url.Values{}
	q.Set("select", selectClause)
	q.Set("order", "created_at.desc")
	q.Set("limit", fmt.Sprintf("%d", limit+1))
	if before != "" {
		q.Set("created_at", "lt."+before)
	}

	endpoint := fmt.Sprintf("%s/rest/v1/community_posts?%s", supabaseURL, q.Encode())
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseAPIKey)
	req.Header.Set("apikey", supabaseAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("list posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("list posts failed (status %d): %s", resp.StatusCode, bodyBytes)
	}

	var rows []CommunityPostWithAttachments
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, false, fmt.Errorf("decode posts: %w", err)
	}

	nameCache := make(map[string]string)
	for i := range rows {
		authorID := rows[i].AuthorID
		if cached, ok := nameCache[authorID]; ok {
			rows[i].AuthorName = cached
			continue
		}
		if name, err := fetchSenderName(authorID); err == nil {
			rows[i].AuthorName = name
			nameCache[authorID] = name
		}
	}

	hasMore := false
	if len(rows) > limit {
		hasMore = true
		rows = rows[:limit]
	}

	return rows, hasMore, nil
}
