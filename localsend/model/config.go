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
package model

// var Config ConfigModel

var defaultConfig = ConfigModel{
	Alias:               "",
	ListenAddress:       "0.0.0.0",
	ListenPort:          53317,
	UdpBroadcastAddress: "224.0.0.167",
	UdpBroadcastPort:    53317,
}

type ConfigModel struct {
	Alias               string
	ListenAddress       string
	ListenPort          int
	UdpBroadcastAddress string
	UdpBroadcastPort    int
}

func NewConfig() ConfigModel {
	return defaultConfig
}
