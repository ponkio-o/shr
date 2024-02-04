package shr

import "github.com/google/go-github/v57/github"

type Shr struct {
	GClient *github.Client
}

var App Shr
