package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/google/go-github/v57/github"
	"github.com/ponkio-o/shr"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

var (
	version = ""
	commit  = ""
)

const (
	COMMAND_NAME     = "shr"
	CONFIG_DIR_NAME  = "shr"
	CONFIG_FILE_NAME = "config"
	CONFIG_FILE_EXT  = "yaml"
)

func init() {
	logLevel := new(slog.LevelVar)
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		logLevel.Set(slog.LevelDebug)
	}
	ops := slog.HandlerOptions{
		Level: logLevel,
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &ops)))
}

func main() {
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("v%s (%s)\n", cCtx.App.Version, commit)
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "token",
			Usage:   "GitHub Token",
			EnvVars: []string{"GITHUB_TOKEN"},
		},
	}

	app := &cli.App{
		Name:    COMMAND_NAME,
		Usage:   fmt.Sprintf("%s is a cli tool that manage GitHub Actions self-hosted runner", COMMAND_NAME),
		Version: version,
		Before:  beforeFunc,
		Flags:   flags,
	}

	app.Commands = []*cli.Command{
		{
			Name:   "list",
			Usage:  "List self-hosted runners",
			Action: shr.ListAction,
			Before: listBefore,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "scope",
					Usage: "Scope that show runners [ent|org]",
					Value: "org",
					Action: func(c *cli.Context, v string) error {
						if v != "org" && v != "ent" {
							return errors.New("scope must be 'ent' or 'org'")
						}
						return nil
					},
				},
				&cli.StringFlag{
					Name:    "owner",
					Aliases: []string{"own"},
					Usage:   "The origanization name",
				},
				&cli.StringFlag{
					Name:  "output",
					Value: "json",
					Usage: "The output options [json|text]",
				},
				&cli.StringSliceFlag{
					Name:  "labels",
					Usage: "filterd by labels of runner",
				},
				&cli.BoolFlag{
					Name:  "busy",
					Usage: "status of runner",
				},
			},
		},
		{
			Name:   "init",
			Usage:  "Initializing the configuration",
			Action: shr.InitAction,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func listBefore(c *cli.Context) error {
	return nil
}

func beforeFunc(c *cli.Context) error {
	viper.AddConfigPath(fmt.Sprintf("%s/%s", xdg.ConfigHome, CONFIG_DIR_NAME))
	viper.SetConfigName(CONFIG_FILE_NAME)
	viper.SetConfigType(CONFIG_FILE_EXT)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	url := viper.GetString("enterprise.url")
	if url != "" {
		shr.App.GClient, err = github.NewClient(nil).WithEnterpriseURLs(url, url)
		if err != nil {
			return err
		}
	} else {
		shr.App.GClient = github.NewClient(nil)
		if err != nil {
			return err
		}
	}

	// GitHub token
	var token string
	if c.IsSet("token") {
		token = c.String("token")
	} else {
		token = viper.GetString("github_token")
	}
	shr.App.GClient = shr.App.GClient.WithAuthToken(token)

	return nil
}
