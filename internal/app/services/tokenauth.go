package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var supabaseURL = os.Getenv("SUPABASE_URL")
var supabaseAnonKey = os.Getenv("SUPABASE_ANON_KEY")

type SupabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// VerifyToken validates JWT with Supabase and returns user info.
func VerifyToken(token string) (*SupabaseUser, error) {
	req, _ := http.NewRequest("GET", supabaseURL+"/auth/v1/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", supabaseAnonKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase auth failed: %s", body)
	}

	var user SupabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
