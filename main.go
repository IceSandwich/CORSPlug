package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	app, err := NewApplication(11451)
	if err != nil {
		log.WithError(err).Fatal("Failed to create application")
	}

	app.Run()
	log.Info("Exiting")
}
