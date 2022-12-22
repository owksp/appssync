package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/txsvc/apikit"
	"github.com/txsvc/apikit/api"
	"github.com/txsvc/apikit/config"
)

func setup() *echo.Echo {
	// create a new router instance
	e := echo.New()

	// add and configure any middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	// OAuth stuff
	e.GET(googleOAuthStart, startEndpoint)
	e.GET(googleOAuthRedirect, redirectEndpoint)

	// done
	return e
}

func shutdown(ctx context.Context, a *apikit.App) error {
	return nil
}

// startEndpoint starts the OAuth 2.0 flow
func startEndpoint(c echo.Context) error {

	// FIXME secure this by expecting the clientId as part of the request
	cfg := GetOAuthConfig(config.GetConfig().Settings().Credentials.UserID, config.GetConfig().Settings().Credentials.Token, config.GetConfig().Settings().DefaultScopes)

	randState = fmt.Sprintf("st%d", time.Now().UnixNano())
	cfg.RedirectURL = fmt.Sprintf("%s%s", config.GetConfig().Settings().Endpoint, googleOAuthRedirect)
	authURL := cfg.AuthCodeURL(randState, oauth2.AccessTypeOffline)
	_config = &cfg

	fmt.Printf("%v\n", cfg)
	fmt.Printf("%v\n", authURL)

	// hand-over to Google for authentication
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// redirectEndpoint handles the OAuth call-back and creates a token on success
func redirectEndpoint(c echo.Context) error {

	state := c.Request().FormValue("state")
	code := c.Request().FormValue("code")

	// FIXME handle cancellation !

	// FIXME validation !!

	fmt.Printf("%s -> %s\n", state, code)

	ctx := c.Request().Context()
	token, err := _config.Exchange(ctx, code)
	if err != nil {
		return err
	}

	// store the token
	f, err := os.OpenFile(filepath.Join(config.GetConfig().ConfigLocation(), tokenFile), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to store oauth token: %v", err)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Fatalf("Unable to store oauth token: %v", err)
	}

	// clean-up
	randState = ""
	_config = nil

	// delayed shutdown ...
	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		svc.Stop()
	}()

	return api.StandardResponse(c, http.StatusOK, nil)
}

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
