package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"google.golang.org/api/option"
	"google.golang.org/api/script/v1"

	"github.com/txsvc/apikit/config"
)

func pullAppsScript(scriptId string) error {
	ctx := context.Background()

	token, err := LoadToken(filepath.Join(config.GetConfig().ConfigLocation(), tokenFile))
	if err != nil || token.AccessToken == "" {
		return fmt.Errorf("invalid token")
	}

	cfg := GetOAuthConfig(config.GetConfig().Settings().Credentials.UserID, config.GetConfig().Settings().Credentials.Token, config.GetConfig().Settings().DefaultScopes)
	client := cfg.Client(ctx, token)

	svc, err := script.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Apps Script client: %v", err)
	}

	// retrieve the list of files in the remote project
	content, err := svc.Projects.GetContent(scriptId).Do()
	if err != nil {
		return err
	}

	// pull the files
	assetLocation := config.GetConfig().(*AppsScriptConfig).AssetsLocation()
	for _, f := range content.Files {
		err := pullFile(f, assetLocation)
		if err != nil {
			return fmt.Errorf("unable to pull Apps Script assets: %v", err)
		}
	}

	return nil
}

func pullFile(f *script.File, path string) error {

	ext := "gs"
	if f.Type == "JSON" {
		ext = "json"
	} else if f.Type == "HTML" {
		ext = "html"
	}
	fullPath := filepath.Join(path, fmt.Sprintf("%s.%s", f.Name, ext))

	// check for sub-folders and create them if necessary
	relPath := filepath.Dir(fullPath)
	if _, err := os.Stat(relPath); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(relPath, os.ModePerm)
	}

	if err := os.WriteFile(fullPath, []byte(f.Source), 0644); err != nil {
		return err
	}

	fmt.Printf("Pulling -> %s.%s\n", f.Name, ext)
	return nil
}
