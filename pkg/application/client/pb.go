package client

import "github.com/leorolland/sortir.in/pkg/application"

type PBClient interface {
	CreateEvent(event application.Event) error
}
