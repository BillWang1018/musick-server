package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseAnon := os.Getenv("SUPABASE_ANON_KEY")
	supabaseAPI := os.Getenv("SUPABASE_API_KEY")
	supabaseService := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	accountID := flag.String("account", "", "account_id to filter room_members")
	keyOverride := flag.String("key", "", "override key (Bearer/apikey)")
	useService := flag.Bool("service", false, "use SUPABASE_SERVICE_ROLE_KEY if set")
	flag.Parse()

	key := supabaseAPI
	keyName := "api"
	if key == "" {
		key = supabaseAnon
		keyName = "anon"
	}
	if *useService && supabaseService != "" {
		key = supabaseService
		keyName = "service"
	}
	if *keyOverride != "" {
		key = *keyOverride
		keyName = "override"
	}

	if supabaseURL == "" || key == "" {
		log.Fatal("SUPABASE_URL or API key is missing")
	}

	target := "rooms"
	q := url.Values{}
	if *accountID != "" {
		// Fetch rooms joined by this account via inner join on room_members.
		target = "rooms"
		q.Set("select", "id,code,owner_id,title,is_private,created_at,room_members!inner(role,account_id)")
		q.Set("room_members.account_id", "eq."+*accountID)
	} else {
		q.Set("select", "id")
	}

	endpoint := fmt.Sprintf("%s/rest/v1/%s?%s", supabaseURL, target, q.Encode())
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("endpoint=%s", endpoint)
	log.Printf("using key=%s status=%d body=%s", keyName, resp.StatusCode, string(body))
}
