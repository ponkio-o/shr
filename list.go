package shr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/olekukonko/tablewriter"
	"github.com/sagikazarmark/slog-shim"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

type ListActionOpts struct {
	Labels  []string
	Scope   string
	Owner   string
	Busy    bool
	EntName string
	IsEnt   bool
	Output  string
}

type Runner struct {
	Name   string
	Labels []string
	Busy   bool
	Status string
}

var opts ListActionOpts

func parseListActionOpts(c *cli.Context) {
	// labels
	if c.IsSet("labels") {
		opts.Labels = c.StringSlice("labels")
	}

	// Scope
	if c.IsSet("scope") {
		opts.Scope = c.String("scope")
	} else {
		opts.Scope = viper.GetString("scope")
	}

	// Is Enterprise
	if opts.Scope == "ent" {
		opts.IsEnt = true
		opts.EntName = viper.GetString("enterprise.name")
	}

	if c.IsSet("owner") {
		opts.Owner = c.String("owner")
	} else {
		opts.Owner = viper.GetString("owner")
	}

	// Outputs
	if c.IsSet("output") {
		opts.Output = c.String("output")
	} else {
		if viper.IsSet("output") {
			opts.Output = viper.GetString("output")
		} else {
			opts.Output = c.String("output")
		}
	}

	// busy
	if c.IsSet("busy") {
		opts.Busy = c.Bool("busy")
	}
}

func ListAction(c *cli.Context) error {
	parseListActionOpts(c)

	var runners *github.Runners
	var err error
	if opts.IsEnt {
		slog.Debug("list enterpirse runners")
		runners, _, err = App.GClient.Enterprise.ListRunners(context.Background(), opts.EntName, &github.ListOptions{})
		if err != nil {
			return err
		}
	} else {
		slog.Debug("list organization runners")
		runners, _, err = App.GClient.Actions.ListOrganizationRunners(context.Background(), opts.Owner, &github.ListOptions{})
		if err != nil {
			return err
		}
	}

	var results []*Runner
	for _, v := range runners.Runners {
		results = append(results, &Runner{
			Name:   v.GetName(),
			Labels: flattenLabelNames(v.Labels),
			Busy:   v.GetBusy(),
			Status: v.GetStatus(),
		})
	}

	filtered, err := runnerFilter(results, opts)
	if err != nil {
		return err
	}

	if err := showRunnersResult(filtered, opts); err != nil {
		return err
	}

	return nil
}

func runnerFilter(runners []*Runner, opts ListActionOpts) ([]*Runner, error) {
	var filteredRunners []*Runner
	var isBusyRunners []*Runner
	var hasLabelsRunners []*Runner

	// default
	// show all runners when not defined --busy and --labels flags
	if !opts.Busy && len(opts.Labels) == 0 {
		return runners, nil
	}

	// filtered by busy
	if opts.Busy {
		for _, runner := range runners {
			if runner.Busy {
				isBusyRunners = append(isBusyRunners, runner)
			}
		}
	}

	// filtered by labels
	if len(opts.Labels) != 0 {
		if opts.Busy {
			filteredRunners = filterHasLabels(isBusyRunners, opts.Labels)
		} else {
			filteredRunners = filterHasLabels(runners, opts.Labels)
		}
	}

	filteredRunners = append(filteredRunners, isBusyRunners...)
	filteredRunners = append(filteredRunners, hasLabelsRunners...)

	return filteredRunners, nil
}

func showRunnersResult(result []*Runner, opts ListActionOpts) error {
	if opts.Output == "json" {
		v, err := json.Marshal(result)
		if err != nil {
			return err
		}

		if string(v) == "null" {
			fmt.Println("{}")
			return nil
		}
		fmt.Println(string(v))

	} else if opts.Output == "text" {
		table := tablewriter.NewWriter(os.Stdout)

		keys := []string{"Name", "Labels", "Busy", "Status"}
		table.SetHeader(keys)

		var data [][]string
		for _, v := range result {
			var row []string
			row = append(row, v.Name)
			row = append(row, strings.Join(v.Labels, ","))
			row = append(row, fmt.Sprintf("%t", v.Busy))
			row = append(row, v.Status)
			data = append(data, row)
		}

		for _, v := range data {
			table.Append(v)
		}

		table.Render()
	} else {
		return fmt.Errorf("%v is not defined in --output option (json or text)", opts.Output)
	}

	return nil
}
