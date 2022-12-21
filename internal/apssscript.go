package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/script/v1"
)

func PullAppsScript(scriptId, path string, cfg *oauth2.Config, token *oauth2.Token) error {
	ctx := context.Background()

	client := cfg.Client(ctx, token)
	svc, err := script.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Apps Script client: %v", err)
	}

	content, err := svc.Projects.GetContent(scriptId).Do()
	if err != nil {
		return err
	}

	// pull the files
	for _, f := range content.Files {
		err := pullFile(f, path)
		if err != nil {
			log.Fatalf("Unable to pull Apps Script assets: %v", err)
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
