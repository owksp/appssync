package internal

import (
	"fmt"
	"log"
	"path/filepath"

	"golang.org/x/oauth2"

	"github.com/urfave/cli/v2"

	"github.com/txsvc/apikit"
	kit "github.com/txsvc/apikit/cli"
	"github.com/txsvc/apikit/config"
)

const (
	ENV_GOOGLE_CLIENT_ID     = "GOOGLE_CLIENT_ID"
	ENV_GOOGLE_CLIENT_SECRET = "GOOGLE_CLIENT_SECRET"

	DefaultAssetLocation = "./appsscript"

	googleOAuthStart    = "/start"
	googleOAuthRedirect = "/a/1/auth"

	tokenFile       = "token.json"
	credentialsFile = "credentials.json"
)

var (
	svc       *apikit.App
	randState string
	_config   *oauth2.Config
)

func CmdAuth(c *cli.Context) error {
	//logger := logger.New()

	if c.NArg() > 0 {
		return kit.ErrInvalidNumArguments
	}

	// check if the tool has a valid access token
	token, err := LoadToken(filepath.Join(config.GetConfig().ConfigLocation(), tokenFile))

	if err != nil || token.AccessToken == "" {
		// start the authorization process
		fmt.Printf("Go to the following link in your browser and authorize the cli first:\n\n%s%s\n\n", config.GetConfig().Settings().Endpoint, googleOAuthStart)

		svc, err = apikit.New(setup, shutdown)
		if err != nil {
			log.Fatal(err)
		}
		svc.Listen("")
	} else {
		fmt.Println("Already authenticated.")
	}

	return nil
}

func CmdPull(c *cli.Context) error {
	//logger := logger.New()

	if path := c.String("dir"); path != "" {
		config.GetConfig().(*AppsScriptConfig).SetAssetsLocation(path)
	}

	if c.NArg() != 1 {
		return kit.ErrInvalidNumArguments
	}
	scriptId := c.Args().First()

	// check if the tool has a valid access token
	token, err := LoadToken(filepath.Join(config.GetConfig().ConfigLocation(), tokenFile))

	if err != nil || token.AccessToken == "" {
		fmt.Println("Not authenticated.")
		return nil
	}

	fmt.Printf("Pulling assets for scriptId '%s'.\n", scriptId)

	err = pullAppsScript(scriptId)
	return err
}
