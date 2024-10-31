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
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	sessions = make(map[string]*model.ReceiveSession)
)

func PrepareReceive(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request) {
	// TODO: add ability to check for existing file with matching size and remove it from returned file list
	pin := r.URL.Query().Get("pin")

	if config.RequirePassword != "" && pin != config.RequirePassword {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req model.PrepareReceiveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	slog.Info(fmt.Sprintf("Received request: %v", req))

	sessionID := uuid.New().String()

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

func ReceiveHandler(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request, sessionCompleteCallback func(string)) {
	sessionID := r.URL.Query().Get("sessionId")
	fileID := r.URL.Query().Get("fileId")
	token := r.URL.Query().Get("token")

	// Verify request parameters
	if sessionID == "" || fileID == "" || token == "" {
		// TODO: create a new `sessionLogger` for this instead of repeating all this code each time below
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

	// verify access is allowed
	// this obviously isn't a very robust access system, but its mainly meant to
	// prevent accidental uploads
	if config.AllowedAliases[0] != "__ALL__" {
		allowed := false

		for _, allowedAlias := range config.AllowedAliases {
			if strings.EqualFold(session.Alias, allowedAlias) {
				allowed = true
				break
			}
		}

		if !allowed {
			http.Error(w, "Access Denied", http.StatusForbidden)
			return
		}
	}

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

	// Generate file paths, preserving file extensions
	outputDirectory, err := filepath.Abs(config.StoragePath)
	if err != nil {
		slog.Error("Failed to get absolute output directory: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	if config.AppendSenderAlias {
		outputDirectory = path.Join(outputDirectory, session.Alias)
	}

	filePath := path.Join(outputDirectory, fileName)

	// Create the folder if it does not exist
	err = os.MkdirAll(outputDirectory, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		slog.Warn("Error creating directory: "+err.Error(),
			slog.String("directory", outputDirectory),
			slog.String("SessionID", sessionID),
			slog.String("FileID", fileID),
			slog.String("Token", token),
			slog.String("FileName", fileName),
		)
		return
	}

	// Create a file
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		slog.Warn("Error creating file: "+err.Error(),
			slog.String("file", filePath),
			slog.String("SessionID", sessionID),
			slog.String("FileID", fileID),
			slog.String("Token", token),
			slog.String("OutputDirectory", outputDirectory),
			slog.String("FileName", fileName),
		)
		return
	}
	defer file.Close()

	buffer := make([]byte, 2*1024*1024) // 2MB buffer
	for {
		n, err := r.Body.Read(buffer)
		if err != nil && err != io.EOF {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			slog.Warn("Error reading file: "+err.Error(),
				slog.String("SessionID", sessionID),
				slog.String("FileID", fileID),
				slog.String("Token", token),
				slog.String("FileName", fileName),
			)
			return
		}
		if n == 0 {
			break
		}

		_, err = file.Write(buffer[:n])
		if err != nil {
			http.Error(w, "Failed to write file", http.StatusInternalServerError)
			slog.Warn("Error writing file: "+err.Error(),
				slog.String("SessionID", sessionID),
				slog.String("FileID", fileID),
				slog.String("Token", token),
				slog.String("OutputDirectory", outputDirectory),
				slog.String("FileName", fileName),
			)
			return

		}
	}

	session.CompletedFiles += 1

	slog.Info(fmt.Sprintf("Finished receiving file '%s' from '%s'", filePath, session.Alias))
	w.WriteHeader(http.StatusOK)

	if session.CompletedFiles == session.TotalFiles {
		slog.Info(fmt.Sprintf("Session '%s' complete, transferred %d file(s)", sessionID, session.CompletedFiles))
		delete(sessions, sessionID)

		// the session has finished, execute the callback
		sessionCompleteCallback(outputDirectory)
	}

}

// ReceiveHandler Handling file download requests
// func DownloadRequestHandler(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request) {
// 	fileName := r.URL.Query().Get("file")
// 	if fileName == "" {
// 		http.Error(w, "File parameter is required", http.StatusBadRequest)
// 		return
// 	}

// 	filePath := config.StoragePath

// 	if config.AppendSenderAlias {
// 		filePath = path.Join(filePath, session.Alias)
// 	}

// 	filePath = path.Join(filePath, fileName)

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not open file: %v", err), http.StatusNotFound)
// 		return
// 	}
// 	defer file.Close()

// 	// Setting the response header
// 	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
// 	w.Header().Set("Content-Type", "application/octet-stream")

// 	// Write the file contents to the response
// 	_, err = io.Copy(w, file)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not write file to response: %v", err), http.StatusInternalServerError)
// 		return
// 	}
// }
