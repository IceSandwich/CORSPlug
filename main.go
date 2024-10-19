package main

import (
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	log "github.com/sirupsen/logrus"
)

func main3() {
	var mw *walk.MainWindow
	declarative.MainWindow{
		AssignTo: &mw,
		Title:    "Close window automatically",
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			declarative.Label{
				Text: "This window will be closed in five seconds.",
			},
			declarative.PushButton{
				Text: "Close instantly",
				OnClicked: func() {
					mw.Close()
				},
			},
		},
	}.Create()

	go func() {
		time.Sleep(5 * time.Second)
		mw.Synchronize(func() {
			mw.Close()
		})
	}()

	mw.Run()
}

func main2() {
	var window *walk.MainWindow
	var err error

	if window, err = walk.NewMainWindow(); err != nil {
		log.WithError(err).Fatal("Failed to create main window")
	}

	go func() {
		time.Sleep(time.Second * time.Duration(3))

		dialog := NewRequestPermissionDialog(nil, "33333", "22222")
		if dialog.Run() == 0 {
			log.Infof("Permission denied for %s", "22222")
		}

		walk.App().Exit(0)
	}()

	window.Run()
}

func main() {
	app, err := NewApplication(11451)
	if err != nil {
		log.WithError(err).Fatal("Failed to create application")
	}

	app.Run()
	log.Info("Exiting")
}
