// =================================================================================
//
//		gim - https://www.foxhollow.cc/projects/gim/
//
//	 go-import-media, aka gim, is a tool for automatically importing media
//	 from removable disks into a predefined folder structure automatically.
//
//		Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>
//
//	This file was originally part of the localsend-go project created by
//	MeowRain. It was adapted for use by Steve Cross in the go-import-media
//	project on 2024-10-30.
//
//	    Copyright (c) 2024 MeowRain
//	    localsend-go - https://github.com/meowrain/localsend-go
//	    License: MIT (for complete text, see LICENSE-MIT file in localsend folder)
//
// =================================================================================
package localsend

import (
	"fmt"
	"gim/localsend/discovery"
	"gim/localsend/handler"
	"gim/model"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/google/uuid"
)

func RunServer(config model.LocalSendConfig, sessionCompleteCallback func(string)) {
	message := model.BroadcastMessage{
		Alias:       config.Alias,
		Version:     "2.0",
		DeviceModel: runtime.GOOS,
		DeviceType:  "server",
		Fingerprint: fmt.Sprintf("%s_%d", uuid.New().String(), config.ListenPort),
		Port:        config.ListenPort,
		Protocol:    "http",
		Download:    false,
		Announce:    true,
	}

	// Enable broadcast and monitoring functions
	go discovery.StartBroadcastUDP(config, message)
	go discovery.StartBroadcastHTTP(config, message)

	// Start HTTP Server
	httpServer := http.NewServeMux()

	/*Send and receive part*/
	httpServer.HandleFunc("/api/localsend/v2/prepare-upload", func(w http.ResponseWriter, r *http.Request) {
		handler.PrepareReceive(config, message, w, r)
	})
	httpServer.HandleFunc("/api/localsend/v2/upload", func(w http.ResponseWriter, r *http.Request) {
		handler.ReceiveHandler(config, message, w, r, sessionCompleteCallback)
	})
	httpServer.HandleFunc("/api/localsend/v2/info", func(w http.ResponseWriter, r *http.Request) {
		handler.RegisterHandler(config, message, w, r)
	})
	httpServer.HandleFunc("/api/localsend/v2/register", func(w http.ResponseWriter, r *http.Request) {
		handler.RegisterHandler(config, message, w, r)
	})

	go func() {
		slog.Info(fmt.Sprintf("Localsend server '%s' started at %s:%d", config.Alias, config.ListenAddress, config.ListenPort))

		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.ListenAddress, config.ListenPort), httpServer); err != nil {
			slog.Error(fmt.Sprintf("Localsend Server '%s' failed: %v", config.Alias, err))
			return
		}
	}()

	// TODO: need graceful shutdown?
	select {} // Blocking program waiting to receive file
}
