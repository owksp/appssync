package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli/v2"

	kit "github.com/txsvc/apikit/cli"
	"github.com/txsvc/apikit/config"

	"mickume/appssync/internal"
)

func init() {
	config.SetProvider(internal.NewConfigProvider())
}

func main() {
	// initialize the CLI
	cfg := config.GetConfig()
	app := &cli.App{
		Name:      cfg.Info().ShortName(),
		Version:   cfg.Info().VersionString(),
		Usage:     cfg.Info().About(),
		Copyright: cfg.Info().Copyright(),
		Commands:  setupCommands(),
		Flags:     setupFlags(),
		Before: func(c *cli.Context) error {
			// handle global config
			if path := c.String("config"); path != "" {
				config.SetConfigLocation(path)
			}
			if path := c.String("cred"); path != "" {
				config.GetConfig().(*internal.AppsScriptConfig).SetCredentials(path)
			}
			return nil
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	// run the CLI
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

//
// CLI commands and flags
//

// setupCommands returns all custom CLI commands and the default ones
func setupCommands() []*cli.Command {
	cmds := []*cli.Command{
		{
			Name:   "auth",
			Usage:  "Authorize the CLI",
			Action: internal.CmdAuth,
		},
		{
			Name:      "pull",
			Usage:     "Pull remote Apps Script assets",
			UsageText: fmt.Sprintf("%s pull <scriptId>", config.GetConfig().Info().ShortName()),
			Action:    internal.CmdPull,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "dir",
					Usage:       "Local assets location",
					DefaultText: internal.DefaultAssetLocation,
					Value:       internal.DefaultAssetLocation,
				},
			},
		},
	}

	return cmds
}

// setupCommands returns all global CLI flags and some default ones
func setupFlags() []cli.Flag {
	credFile := fmt.Sprintf("%s/credentials.json", config.DefaultConfigLocation)

	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
		},
		&cli.StringFlag{
			Name:        "cred",
			Usage:       "credentials file to use",
			DefaultText: credFile,
			Value:       credFile,
		},
	}

	// merge with global flags
	return kit.MergeFlags(flags, kit.WithGlobalFlags())
}
