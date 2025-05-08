package main

import (
	"encoding/json"
	"fmt"
	"fund-manager/config"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

func init() {
	config.LoadEnv()
}

func main() {
	var (
		clientID     = os.Getenv("UPSTOX_API_KEY")
		clientSecret = os.Getenv("UPSTOX_API_SECRET")
		redirectURI  = "http://localhost:3000/callback"
		authURL      = "https://api.upstox.com/v2/login/authorization/dialog"
		tokenURL     = "https://api.upstox.com/v2/login/authorization/token"
		authCode     string
	)
	fmt.Println(clientID)
	if clientID == "" || clientSecret == "" {
		log.Fatal("Missing UPSTOX_CLIENT_ID or UPSTOX_CLIENT_SECRET in environment")
	}

	// Step 1: Start a local HTTP server to capture the code
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}
		authCode = code
		fmt.Fprintln(w, "Authorization successful! You can close this window.")
	})

	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	go http.Serve(ln, nil)

	// Step 2: Open the authorization URL in the browser
	authQuery := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&state=abc123", authURL, url.QueryEscape(clientID), url.QueryEscape(redirectURI))
	fmt.Println("Opening browser for authentication...")
	exec.Command("open", authQuery).Start() // Use "xdg-open" on Linux or "start" on Windows

	// Wait for the user to complete auth
	for authCode == "" {
		// Busy wait for callback to hit
	}

	// Step 3: Exchange code for token
	data := url.Values{}
	data.Set("code", authCode)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		log.Fatal("Failed to exchange code for token:", err)
	}
	defer resp.Body.Close()

	var tokenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		log.Fatal("Failed to decode token response:", err)
	}

	fmt.Println("Access Token Response:")
	prettyJSON, _ := json.MarshalIndent(tokenResp, "", "  ")
	fmt.Println(string(prettyJSON))
}
