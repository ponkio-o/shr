package shr

import (
	"github.com/google/go-github/v57/github"
)

func flattenLabelNames(labels []*github.RunnerLabels) []string {
	var names []string
	for _, v := range labels {
		names = append(names, v.GetName())
	}

	return names
}
