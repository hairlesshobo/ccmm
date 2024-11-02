// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
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
package discovery

import (
	"ccmm/model"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// StartBroadcastUDP Sending a Broadcast Message
func StartBroadcastUDP(config model.LocalSendConfig, message model.BroadcastMessage) {
	// Set the multicast address and port
	multicastAddr := &net.UDPAddr{
		IP:   net.ParseIP(config.UdpBroadcastAddress),
		Port: config.UdpBroadcastPort,
	}

	data, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	// Create a local address and bind it to the configure address
	localAddr := &net.UDPAddr{
		IP:   net.ParseIP(config.ListenAddress),
		Port: 0,
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		return
	}
	defer conn.Close()
	for {
		_, err := conn.WriteToUDP(data, multicastAddr)
		if err != nil {
			fmt.Println("Failed to send message:", err)
			panic(err)
		}
		// fmt.Println(num, "bytes write to multicastAddr")
		//log
		// fmt.Println("UDP Broadcast message sent!")
		time.Sleep(5 * time.Second) // Send a broadcast message every 5 seconds
	}
}
