package shr

import "golang.org/x/exp/slices"

func filterHasLabels(runners []*Runner, labels []string) []*Runner {
	var hasLabelsRunners []*Runner
	for _, runner := range runners {
		flag := false
		for _, label := range opts.Labels {
			if slices.Contains(runner.Labels, label) {
				flag = true
			} else {
				flag = false
				break
			}
		}
		if flag {
			hasLabelsRunners = append(hasLabelsRunners, runner)
		}
	}

	return hasLabelsRunners
}
