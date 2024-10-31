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
package handler

import (
	"encoding/json"
	"fmt"
	"gim/model"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	sessionMutex     sync.Mutex
	sessionIDCounter = 0
	sessions         = make(map[string]*model.ReceiveSession)
)

func PrepareReceive(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request) {
	var req model.PrepareReceiveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	slog.Info(fmt.Sprintf("Received request: %v", req))

	sessionMutex.Lock()
	sessionIDCounter++
	sessionID := fmt.Sprintf("session-%d", sessionIDCounter)
	sessionMutex.Unlock()

	sessions[sessionID] = &model.ReceiveSession{
		Alias:          req.Info.Alias,
		TotalFiles:     0,
		CompletedFiles: 0,
		Files:          make(map[string]model.FileInfo),
	}

	files := make(map[string]string)
	for fileID, fileInfo := range req.Files {
		token := fmt.Sprintf("token-%s", fileID)
		files[fileID] = token

		// Save file name
		sessions[sessionID].Files[fileID] = fileInfo
		sessions[sessionID].TotalFiles += 1
	}

	resp := model.PrepareReceiveResponse{
		SessionID: sessionID,
		Files:     files,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func ReceiveHandler(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	fileID := r.URL.Query().Get("fileId")
	token := r.URL.Query().Get("token")

	// Verify request parameters
	if sessionID == "" || fileID == "" || token == "" {
		slog.Warn("Upload missing parameters",
			slog.String("SessionID", sessionID),
			slog.String("FileID", fileID),
			slog.String("Token", token),
		)
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	session, ok := sessions[sessionID]
	if !ok {
		slog.Warn("Invalid session ID",
			slog.String("SessionID", sessionID),
			slog.String("FileID", fileID),
			slog.String("Token", token),
		)
		http.Error(w, "Couldn't find sessionID "+sessionID, http.StatusBadRequest)
	}

	fmt.Printf("%+v\n", session)
	// Use fileID to get the file name
	fileEntry, ok := session.Files[fileID]
	if !ok {
		slog.Warn("Invalid file ID",
			slog.String("SessionID", sessionID),
			slog.String("FileID", fileID),
			slog.String("Token", token),
		)
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}
	fileName := fileEntry.FileName

	slog.Info("Beginning to receive file",
		slog.String("SessionID", sessionID),
		slog.String("FileID", fileID),
		slog.String("Token", token),
		slog.String("FileName", fileName),
	)

	// TODO: fix this assumption
	// Generate file paths, preserving file extensions
	filePath := filepath.Join("uploads", session.Alias, fileName)

	// Create the folder if it does not exist
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		fmt.Println("Error creating directory:", err)
		return
	}

	// Create a file
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 2*1024*1024) // 2MB buffer
	for {
		n, err := r.Body.Read(buffer)
		if err != nil && err != io.EOF {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			fmt.Println("Error reading file:", err)
			return
		}
		if n == 0 {
			break
		}

		_, err = file.Write(buffer[:n])
		if err != nil {
			http.Error(w, "Failed to write file", http.StatusInternalServerError)
			fmt.Println("Error writing file:", err)
			return

		}
	}

	session.CompletedFiles += 1

	fmt.Println("Saved file:", filePath)
	w.WriteHeader(http.StatusOK)

	if session.CompletedFiles == session.TotalFiles {
		slog.Info(fmt.Sprintf("Session '%s' complete, transferred %d file(s)", sessionID, session.CompletedFiles))
		delete(sessions, sessionID)
		// TODO: take action on completed session
	}

}

// ReceiveHandler Handling file download requests
func DownloadRequestHandler(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		http.Error(w, "File parameter is required", http.StatusBadRequest)
		return
	}

	// TODO: fix this assumption
	// Assuming the files are stored in the "uploads" directory
	filePath := filepath.Join("uploads", fileName)
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not open file: %v", err), http.StatusNotFound)
		return
	}
	defer file.Close()

	// Setting the response header
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Write the file contents to the response
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not write file to response: %v", err), http.StatusInternalServerError)
		return
	}
}
