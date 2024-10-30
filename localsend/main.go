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
//	    License: MIT (for complete text, see LICENSE file in localsend folder)
//
// =================================================================================
package main

import (
	"flag"
	"fmt"
	"gim/localsend/discovery"
	"gim/localsend/discovery/shared"
	"gim/localsend/handler"
	"gim/localsend/model"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/google/uuid"
)

func main() {
	port := flag.Int("port", 53317, "Port to listen on")
	udp_broadcast_address := flag.String("udp_address", "224.0.0.167", "Address to use for UDP multicast. Must match clients on network.")
	udp_broadcast_port := flag.Int("udp_port", 53317, "Port to use for UDP multicast. Must match clients on network.")
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Failed to get hostname, error: ", err)
		os.Exit(1)
	}

	config := model.NewConfig()
	config.Alias = fmt.Sprintf("%s__%d", hostname, *port)
	config.ListenPort = *port
	config.UdpBroadcastAddress = *udp_broadcast_address
	config.UdpBroadcastPort = *udp_broadcast_port

	RunServer(config)
}

func RunServer(config model.ConfigModel) {
	model.Config = config

	shared.Messsage = model.BroadcastMessage{
		Alias:       model.Config.Alias,
		Version:     "2.0",
		DeviceModel: runtime.GOOS,
		DeviceType:  "server",
		Fingerprint: uuid.New().String(),
		Port:        model.Config.ListenPort,
		Protocol:    "http",
		Download:    true,
		Announce:    true,
	}

	// Enable broadcast and monitoring functions
	go discovery.StartBroadcast()
	go discovery.StartHTTPBroadcast()

	// Start HTTP Server
	httpServer := http.NewServeMux()

	/*Send and receive part*/
	httpServer.HandleFunc("/api/localsend/v2/prepare-upload", handler.PrepareReceive)
	httpServer.HandleFunc("/api/localsend/v2/upload", handler.ReceiveHandler)
	httpServer.HandleFunc("/api/localsend/v2/info", handler.RegisterHandler)
	httpServer.HandleFunc("/api/localsend/v2/register", handler.RegisterHandler)
	httpServer.HandleFunc("/receive", handler.DownloadRequestHandler) // Download Handler

	go func() {
		log.Printf("Server '%s' started at :%d\n", shared.Messsage.Alias, model.Config.ListenPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", model.Config.ListenPort), httpServer); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	fmt.Println("Waiting to receive files...")
	// TODO: need graceful shutdown?
	select {} // Blocking program waiting to receive file
}
