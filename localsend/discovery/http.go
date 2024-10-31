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
package discovery

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gim/model"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

// getLocalIP Get the local IP address
func getLocalIP() ([]net.IP, error) {
	ips := make([]net.IP, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil && !v.IP.IsLoopback() {
					ips = append(ips, v.IP)
				}
			}
		}
	}
	return ips, nil
}

// pingScan Use ICMP ping to scan all active devices on the LAN
func pingScan() ([]string, error) {
	var ips []string

	ipGroup, err := getLocalIP()
	if err != nil {
		return nil, err
	}

	for _, i := range ipGroup {
		// TODO: fix this assumption
		ip := i.Mask(net.IPv4Mask(255, 255, 255, 0)) // Assume the subnet mask is 24
		ip4 := ip.To4()
		if ip4 == nil {
			return nil, fmt.Errorf("invalid IPv4 address")
		}

		var wg sync.WaitGroup
		var mu sync.Mutex

		for i := 1; i < 255; i++ {
			ip4[3] = byte(i)
			targetIP := ip4.String()

			wg.Add(1)
			go func(ip string) {
				defer wg.Done()
				pinger, err := probing.NewPinger(ip)
				if err != nil {
					slog.Error(fmt.Sprintf("Failed to create pinger: %s", err.Error()))
					return
				}
				pinger.SetPrivileged(true)
				pinger.Count = 1
				pinger.Timeout = time.Second * 1

				pinger.OnRecv = func(pkt *probing.Packet) {
					mu.Lock()
					ips = append(ips, ip)
					mu.Unlock()
				}
				err = pinger.Run()
				if err != nil {
					//Ignore ping failures
					return
				}
			}(targetIP)
		}

		wg.Wait()
	}

	return ips, nil
}

// StartBroadcastHTTP Send HTTP requests to all IPs in the LAN
func StartBroadcastHTTP(config model.LocalSendConfig, message model.BroadcastMessage) {

	for {
		data, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}

		ips, err := pingScan()
		if err != nil {
			slog.Warn(fmt.Sprintf("Failed to discover devices via ping scan: %s", err.Error()))
			return
		}

		var wg sync.WaitGroup
		for _, ip := range ips {
			wg.Add(1)
			go func(ip string) {
				defer wg.Done()
				ctx := context.Background()
				registerWithHttp(config, ctx, ip, data)
			}(ip)
		}

		wg.Wait()

		time.Sleep(5 * time.Second) // Send HTTP broadcast message every 5 seconds
	}
}

// registerWithHttp Sending HTTP Requests
func registerWithHttp(config model.LocalSendConfig, ctx context.Context, ip string, data []byte) {
	url := fmt.Sprintf("https://%s:%d/api/localsend/v2/register", ip, config.ListenPort)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create HTTP request: %s", err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to read HTTP response body: %s", err.Error()))
		return
	}
	var response model.BroadcastMessage
	err = json.Unmarshal(body, &response)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to parse HTTP response from %s: %v", ip, err.Error()))
		return
	}
}
