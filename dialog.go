package main

import (
	"fmt"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
)

type RequestPermissionDialog struct {
	Dialog *walk.Dialog

	srcAddr     *walk.Label
	dstAddr     *walk.Label
	allowButton *walk.PushButton

	isDialogRunning bool
}

func NewRequestPermissionDialog(parent walk.Form, srcAddr, dstAddr string) *RequestPermissionDialog {
	// const imgFilename = "xxx"
	// imgFile, err := walk.NewImageFromFileForDPI(imgFilename, 300)
	// if err != nil {
	// 	log.WithError(err).Errorf("Failed to load image: {}", imgFilename)
	// }

	ret := &RequestPermissionDialog{}
	declarative.Dialog{
		AssignTo: &ret.Dialog,
		Title:    "CORSPlug - Connection request",
		Size: declarative.Size{
			Width:  600,
			Height: 400,
		},
		FixedSize: true,
		Layout:    declarative.VBox{},
		Children: []declarative.Widget{
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.ImageView{
						// Image: imgFile,
					},
					declarative.Composite{
						Layout: declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Composite{
								Layout: declarative.HBox{},
								Children: []declarative.Widget{
									declarative.Label{
										AssignTo: &ret.srcAddr,
										Text:     srcAddr,
										Font: declarative.Font{
											Bold: true,
										},
									},
									declarative.Label{
										Text: "wants to connect",
									},
									declarative.Label{
										AssignTo: &ret.dstAddr,
										Text:     dstAddr,
										Font: declarative.Font{
											Bold: true,
										},
									},
								},
							},
							declarative.Composite{
								Layout:    declarative.HBox{},
								Alignment: declarative.Alignment2D(walk.AlignNear),
								Children: []declarative.Widget{
									declarative.Label{
										Text: "Do you allow it?",
									},
								},
							},
						},
					},
				},
			},
			declarative.Composite{
				Layout: declarative.HBox{},
				Children: []declarative.Widget{
					declarative.PushButton{
						AssignTo:  &ret.allowButton,
						Text:      "Allow",
						Enabled:   false,
						OnClicked: ret.onClickYes,
					},
					declarative.PushButton{
						Text:      "No",
						OnClicked: ret.onClickNo,
					},
				},
			},
		},
	}.Create(parent)
	ret.Dialog.Starting().Attach(ret.starting)
	return ret
}

func (ins *RequestPermissionDialog) onClickYes() {
	ins.isDialogRunning = false
	ins.Dialog.Close(1)
}

func (ins *RequestPermissionDialog) onClickNo() {
	ins.isDialogRunning = false
	ins.Dialog.Close(0)
}

func (ins *RequestPermissionDialog) starting() {
	go func() {
		txt := ins.allowButton.Text()
		for waitTime := 5; waitTime >= 0 && ins.isDialogRunning; waitTime-- {
			newText := fmt.Sprintf("%s (%ds)", txt, waitTime)
			ins.allowButton.SetText(newText)
			time.Sleep(time.Second * 1)
		}
		ins.allowButton.SetEnabled(true)
		ins.allowButton.SetText(txt)
	}()

	go func() {
		txt := ins.Dialog.Title()
		for waitTime := 20; waitTime >= 0 && ins.isDialogRunning; waitTime-- {
			newText := fmt.Sprintf("%s (%ds)", txt, waitTime)
			ins.Dialog.SetTitle(newText)
			time.Sleep(time.Second * 1)
		}
		ins.isDialogRunning = false
		ins.Dialog.Synchronize(func() {
			ins.Dialog.Close(0)
		})
	}()
}

func (ins *RequestPermissionDialog) Run() int {
	ins.isDialogRunning = true
	ins.Dialog.Run()
	return ins.Dialog.Result()
}
