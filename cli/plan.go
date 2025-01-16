package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
)

var PlanCommand = &cli.Command{
	Name:                  "plan",
	Aliases:               []string{"p"},
	Usage:                 "generate a build plan for a directory",
	ArgsUsage:             "DIRECTORY",
	EnableShellCompletion: true,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "env",
			Usage: "environment variables to set. format: KEY=VALUE",
		},
		&cli.StringFlag{
			Name:    "out",
			Aliases: []string{"o"},
			Usage:   "output file name",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "output format. one of: json",
			Value: "json",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		buildResult, _, err := GenerateBuildResultForCommand(cmd)
		if err != nil {
			return cli.Exit(err, 1)
		}

		serializedPlan, err := json.MarshalIndent(buildResult.Plan, "", "  ")
		if err != nil {
			return cli.Exit(err, 1)
		}

		output := cmd.String("out")
		if output == "" {
			// Write to stdout if no output file specified
			os.Stdout.Write(serializedPlan)
			os.Stdout.Write([]byte("\n"))
			return nil
		} else {
			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				return cli.Exit(err, 1)
			}

			err = os.WriteFile(output, serializedPlan, 0644)
			if err != nil {
				return cli.Exit(err, 1)
			}

			log.Infof("Plan written to %s", output)
		}

		return nil
	},
}
