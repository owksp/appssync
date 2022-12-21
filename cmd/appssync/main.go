package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mickume/appssync/internal"
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
	"github.com/txsvc/stdlib/v2"
)

const (
	ENV_GOOGLE_CLIENT_ID     = "GOOGLE_CLIENT_ID"
	ENV_GOOGLE_CLIENT_SECRET = "GOOGLE_CLIENT_SECRET"
	ENV_APPS_SCRIPT_ID       = "APPS_SCRIPT_ID"

	googleOAuthStart    = "/start"
	googleOAuthRedirect = "/a/1/auth"

	baseUrl           = "http://localhost:8080"
	tokenFile         = "token.json"
	credentialsFile   = "credentials.json"
	defaultTargetPath = "appsscript"
	defaultConfigPath = ".config"
)

var (
	appsScriptID *string
	clientID     string
	clientSecret string

	randState  string
	config     *oauth2.Config
	configPath *string

	appsScripsScopes = []string{
		"https://www.googleapis.com/auth/script.projects",
		"https://www.googleapis.com/auth/script.projects.readonly",
	}

	svc *apikit.App
)

func main() {
	// example of using a cmd line flag for configuration
	appsScriptID = flag.String("id", "", "Google Apps Script Id")
	cliClientID := flag.String("clientid", "", "Google Client Id")
	cliClientSecret := flag.String("clientsecret", "", "Google Client Secret")
	targetPath := flag.String("path", defaultTargetPath, "Apps Script assets path")
	configPath = flag.String("config", defaultConfigPath, "Config path")
	flag.Parse()

	// read the ClientID & ClientSecret from the credentials.json if it exists
	b, err := os.ReadFile(filepath.Join(*configPath, credentialsFile))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err = google.ConfigFromJSON(b, "https://www.googleapis.com/auth/script.projects")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	if *cliClientID == "" {
		if config.ClientID != "" {
			clientID = config.ClientID
		} else {
			clientID = stdlib.GetString(ENV_GOOGLE_CLIENT_ID, "")
		}
	}

	if *cliClientSecret == "" {
		if config.ClientSecret != "" {
			clientSecret = config.ClientSecret
		} else {
			clientSecret = stdlib.GetString(ENV_GOOGLE_CLIENT_ID, "")
		}
	}

	// check if the tool has a valid access token
	token, err := internal.LoadToken(filepath.Join(*configPath, tokenFile))

	if err != nil || token.AccessToken == "" {
		// start the authorization process
		fmt.Printf("Go to the following link in your browser and authorize the tool first:\n%s%s\n", baseUrl, googleOAuthStart)

		svc, err = apikit.New(setup, shutdown)
		if err != nil {
			log.Fatal(err)
		}
		svc.Listen("")
	} else {
		err := internal.PullAppsScript(*appsScriptID, *targetPath, config, token)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("Pulled Apps Script assets for scriptId '%s'.\n", *appsScriptID)
		}
	}
}

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
	cfg := internal.GetOAuthConfig(clientID, clientSecret, appsScripsScopes)

	randState = fmt.Sprintf("st%d", time.Now().UnixNano())
	cfg.RedirectURL = fmt.Sprintf("%s%s", baseUrl, googleOAuthRedirect)
	authURL := cfg.AuthCodeURL(randState, oauth2.AccessTypeOffline)
	config = &cfg

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
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return err
	}

	// store the token
	if err := internal.StoreToken(filepath.Join(*configPath, tokenFile), config, token); err != nil {
		return err
	}

	// clean-up
	randState = ""
	config = nil

	// delayed shutdown ...
	go func() {
		time.Sleep(time.Duration(5) * time.Second)
		svc.Stop()
	}()

	return api.StandardResponse(c, http.StatusOK, nil)
}
