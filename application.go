package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lxn/walk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Session struct {
	Origin     string
	TargetHost string
}

type Application struct {
	MainLoop *walk.MainWindow
	Icon     *walk.Icon
	Notify   *walk.NotifyIcon
	Server   *http.Server

	// session id to session instance
	Sessions map[string]Session

	// origin, targetHost hash to session id
	revMappingSessions map[string]string
}

func NewApplication(port int) (_ *Application, err error) {
	ins := new(Application)
	ins.Sessions = make(map[string]Session)
	ins.revMappingSessions = make(map[string]string)

	if ins.MainLoop, err = walk.NewMainWindow(); err != nil {
		return nil, errors.Wrapf(err, "Failed to create main window")
	}

	if ins.Icon, err = walk.Resources.Icon("syncthing.ico"); err != nil {
		return nil, errors.Wrapf(err, "Failed to load icon")
	}

	if ins.Notify, err = walk.NewNotifyIcon(ins.MainLoop); err != nil {
		return nil, errors.Wrapf(err, "Failed to create notify icon")
	}
	defer func() {
		if err != nil {
			ins.Notify.Dispose()
		}
	}()
	if err = ins.Notify.SetIcon(ins.Icon); err != nil {
		return nil, errors.Wrapf(err, "Failed to set icon for notify icon")
	}
	if err = ins.Notify.SetVisible(true); err != nil {
		return nil, errors.Wrapf(err, "Failed to set visible for notify icon")
	}

	ins.setupTrayMenu()

	mux := http.NewServeMux()
	mux.HandleFunc("/require_permission", ins.httpRequirePermission)
	mux.HandleFunc("/", ins.httpHandler)
	ins.Server = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
		// Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 	mux.ServeHTTP(w, r)
		// 	if w.Header().Get("Content-Type") == "" {
		// 		ins.httpHandler(w, r)
		// 	}
		// }),
	}
	return ins, nil
}

type ResultCode int

const (
	ResultCodeSuccess          ResultCode = 0
	ResultCodePermissionDenied ResultCode = 100 + iota
	ResultCodeRequireOrigin
	ResultCodeRequireHost
	ResultCodeInValidSession
)

type JsonResult struct {
	Code ResultCode  `json:"code"`
	Msg  interface{} `json:"msg"`
}

func (rc ResultCode) String() string {
	switch rc {
	case ResultCodeSuccess:
		return "success"
	case ResultCodePermissionDenied:
		return "permission denied"
	case ResultCodeRequireOrigin:
		return "require origin in header"
	case ResultCodeRequireHost:
		return "require host in form"
	case ResultCodeInValidSession:
		return "invalid session"
	}
	return "unknown error code"
}

func writeJson(w http.ResponseWriter, result interface{}) error {
	msg, err := json.Marshal(result)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal json result")
	}
	_, err = w.Write(msg)
	if err != nil {
		return errors.Wrapf(err, "Failed to write json result")
	}
	return nil
}

func writeError(w http.ResponseWriter, code ResultCode) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	return writeJson(w, JsonResult{
		Code: code,
		Msg:  code.String(),
	})
}

func writeSuccess(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return writeJson(w, JsonResult{
		Code: ResultCodeSuccess,
		Msg:  data,
	})
}

func (ins *Application) calcHash(origin string, targetHost string) string {
	return origin + "_CORSPLUG_SPLIT_" + targetHost
}

func (ins *Application) httpRequirePermission(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if origin == "" {
		writeError(w, ResultCodeRequireOrigin)
		return
	}

	targetHost := r.FormValue("host")
	if targetHost == "" {
		writeError(w, ResultCodeRequireHost)
		return
	}

	if sid, ok := ins.revMappingSessions[ins.calcHash(origin, targetHost)]; ok {
		writeSuccess(w, sid)
		log.Infof("Use existed session: %s", sid)
		return
	}

	dialog := NewRequestPermissionDialog(nil, origin, targetHost)
	if dialog.Run() == 0 {
		log.Infof("Permission denied for %s", targetHost)
		writeError(w, ResultCodePermissionDenied)
		return
	}

	session := Session{
		Origin:     origin,
		TargetHost: targetHost,
	}
	sessionID := strings.ReplaceAll(uuid.New().String(), "-", "")
	ins.Sessions[sessionID] = session
	ins.revMappingSessions[ins.calcHash(origin, targetHost)] = sessionID
	writeSuccess(w, sessionID)
	log.Infof("New session: %s", sessionID)
}

func (ins *Application) httpHandler(w http.ResponseWriter, r *http.Request) {
	// allow error msg to pass through
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := strings.TrimPrefix(r.URL.Path, "/")
	requestPath := ""
	if idx := strings.Index(sessionID, "/"); idx != -1 {
		requestPath = sessionID[idx+1:]
		sessionID = sessionID[:idx]
	}
	if r.URL.RawQuery != "" {
		requestPath += "?" + r.URL.RawQuery
	}
	// log.Info("Request session: ", sessionID)
	// log.Info("Request path: ", requestPath)

	session, hasSessionID := ins.Sessions[sessionID]
	if !hasSessionID {
		writeError(w, ResultCodeInValidSession)
		return
	}

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent) // 返回 204
		return
	}

	targetURL := "http://" + session.TargetHost + "/" + requestPath
	log.Info("Proxy: /", requestPath, " --> ", targetURL)

	// 创建一个新的请求
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	// 复制原请求的头部信息
	corsplugHeaders := make(map[string][]string)
	for key, value := range r.Header {
		if strings.HasPrefix(key, "CORSPlug-") {
			corsplugHeaders[strings.TrimPrefix(key, "CORSPlug-")] = value
		} else {
			req.Header[key] = value
		}
	}

	if removeHeaders, hasRemoveHeaders := corsplugHeaders["RemoveHeaders"]; hasRemoveHeaders {
		for _, headerName := range strings.Split(strings.ReplaceAll(removeHeaders[0], " ", ""), ",") {
			req.Header.Del(headerName)
		}
	}

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error forwarding request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 读取目标服务器的响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	// 将目标服务器的响应写回原请求的响应
	w.WriteHeader(resp.StatusCode)
	w.Header().Set("Access-Control-Allow-Origin", session.Origin)
	w.Write(body)
}

func (ins *Application) setupQuitTrayAction() error {
	action := walk.NewAction()
	if err := action.SetText("Quit CORSPlug"); err != nil {
		return errors.Wrapf(err, "Failed to set text for quit action")
	}
	action.Triggered().Attach(func() {
		ins.Destroy()
	})
	if err := ins.Notify.ContextMenu().Actions().Add(action); err != nil {
		return errors.Wrapf(err, "Failed to add action for quit")
	}

	return nil
}

func (ins *Application) setupAutoStartupTrayAction() error {
	action := walk.NewAction()
	if err := action.SetText("Auto startup"); err != nil {
		return errors.Wrapf(err, "Failed to set text for auto startup action")
	}
	if err := action.SetCheckable(true); err != nil {
		return errors.Wrapf(err, "Failed to set checkable for auto startup action")
	}
	action.Triggered().Attach(func() {
		if action.Checked() {
			fmt.Println("checked")
		} else {
			fmt.Println("unchecked")
		}
	})
	if err := ins.Notify.ContextMenu().Actions().Add(action); err != nil {
		return errors.Wrapf(err, "Failed to add action for auto startup")
	}

	return nil
}

func (ins *Application) setupTrayMenu() error {
	if err := ins.setupAutoStartupTrayAction(); err != nil {
		return err
	}
	if err := ins.setupQuitTrayAction(); err != nil {
		return err
	}
	return nil
}

func (ins *Application) Run() {
	go func() {
		log.Infof("Listening on %v", ins.Server.Addr)
		if err := ins.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Server error] %+v\n", err)
			ins.Destroy()
		}
	}()
	ins.MainLoop.Run()
}

func (ins *Application) Destroy() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ins.Server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Failed to shutdown server")
	}

	walk.App().Exit(0)
	ins.Notify.Dispose()
}
