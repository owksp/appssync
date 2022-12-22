package internal

import (
	"log"
	"os"
	"path/filepath"

	"github.com/txsvc/stdlib/v2"
	"golang.org/x/oauth2/google"

	"github.com/txsvc/apikit/config"
	"github.com/txsvc/apikit/helpers"
	"github.com/txsvc/apikit/settings"
)

// the below version numbers should match the git release tags,
// i.e. there should be a version 'v0.1.0' on branch main !
const (
	majorVersion = 0
	minorVersion = 1
	fixVersion   = 0
)

type (
	AppsScriptConfig struct {
		// the interface to implement
		config.ConfigProvider

		// app info
		info *config.Info
		// path to configuration settings
		rootDir   string // the current working dir
		assetsDir string // assets dir (default: ./appsscript)
		confDir   string // the fully qualified path to the conf dir
		cred      string // the path to the credentials file (default: ./.config/credentials.json)
		// cached settings
		ds *settings.DialSettings
	}
)

func NewConfigProvider() config.ConfigProvider {

	info := config.NewAppInfo("Simple Apps Script CLI", "simpleas",
		"Copyright 2022, transformative.services, https://txs.vc",
		"A simpler Google Apps Script CLI",
		majorVersion, minorVersion, fixVersion)

	// get the current working dir. abort on error
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	c := &AppsScriptConfig{
		rootDir:   dir,
		assetsDir: "",
		confDir:   "",
		cred:      "",
		info:      &info,
	}

	return c
}

func (c *AppsScriptConfig) Info() *config.Info {
	return c.info
}

// ConfigLocation returns the config location that was set using SetConfigLocation().
// If no location is defined, GetConfigLocation looks for ENV['CONFIG_LOCATION'] or
// returns DefaultConfigLocation() if no environment variable was set.
func (c *AppsScriptConfig) ConfigLocation() string {
	if len(c.confDir) == 0 {
		return stdlib.GetString(config.ConfigDirLocationENV, config.DefaultConfigLocation)
	}
	return c.confDir
}

func (c *AppsScriptConfig) SetConfigLocation(loc string) {
	c.confDir = loc
	if c.ds != nil {
		c.ds = nil // force a reload the next time GetSettings() is called ...
	}
}

func (c *AppsScriptConfig) SetCredentials(cred string) {
	c.cred = cred
	if c.ds != nil {
		c.ds = nil // force a reload the next time GetSettings() is called ...
	}
}

func (c *AppsScriptConfig) AssetsLocation() string {
	if len(c.assetsDir) == 0 {
		return filepath.Join(c.rootDir, DefaultAssetLocation)
	}
	return c.assetsDir
}

func (c *AppsScriptConfig) SetAssetsLocation(loc string) {
	c.assetsDir = loc
}

func (c *AppsScriptConfig) Settings() *settings.DialSettings {
	if c.ds != nil {
		return c.ds
	}

	// try to load the dial settings
	pathToFile := filepath.Join(c.ConfigLocation(), config.DefaultConfigName)
	cfg, err := helpers.ReadDialSettings(pathToFile)
	if err != nil {
		cfg = c.defaultSettings()

		// read the ClientID & ClientSecret from the credentials.json if it exists
		path := filepath.Join(c.ConfigLocation(), credentialsFile)
		if len(c.cred) != 0 {
			path = c.cred
		}
		if b, err := os.ReadFile(path); err == nil {
			cred, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/script.projects")
			if err == nil {
				cfg.Credentials.UserID = cred.ClientID    // ClientId
				cfg.Credentials.Token = cred.ClientSecret // ClientSecret
			}
		}

		// save to the default location
		if err = helpers.WriteDialSettings(cfg, pathToFile); err != nil {
			log.Fatal(err)
		}
	}

	// patch values from ENV, if available
	cfg.Endpoint = stdlib.GetString(config.APIEndpointENV, cfg.Endpoint)
	cfg.Credentials.UserID = stdlib.GetString(ENV_GOOGLE_CLIENT_ID, cfg.Credentials.UserID)
	cfg.Credentials.Token = stdlib.GetString(ENV_GOOGLE_CLIENT_SECRET, cfg.Credentials.Token)

	// make it available for future calls
	c.ds = cfg
	return c.ds
}

func (c *AppsScriptConfig) defaultSettings() *settings.DialSettings {
	return &settings.DialSettings{
		Endpoint:      config.DefaultEndpoint,
		DefaultScopes: defaultScopes(),
		Credentials:   &settings.Credentials{}, // add this to avoid NPEs further down
	}
}

func defaultScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/script.projects",
		"https://www.googleapis.com/auth/script.projects.readonly",
	}
}
