package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GetOAuthConfig(clientId, clientSecret string, scopes []string) oauth2.Config {
	cfg := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       scopes,
	}
	cfg.Endpoint.AuthStyle = oauth2.AuthStyleAutoDetect
	return cfg
}

func NewOAuthClient(ctx context.Context, path string, cfg *oauth2.Config) (*http.Client, error) {
	token, err := LoadToken(path)
	if err != nil {
		return nil, err
	}
	return cfg.Client(ctx, token), nil
}

func StoreToken(path string, cfg *oauth2.Config, token *oauth2.Token) error {

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to store oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func LoadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}
