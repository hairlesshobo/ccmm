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
	"ccmm/model"
	"ccmm/util"
	"encoding/json"
	"fmt"
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

func getOutputFilePath(config model.LocalSendConfig, alias string, fileName string) (string, string, error) {
	// Generate file paths, preserving file extensions
	outputDirectory, err := filepath.Abs(config.StoragePath)
	if err != nil {
		return "", "", err
	}

	if config.AppendSenderAlias {
		outputDirectory = path.Join(outputDirectory, alias)
	}

	return outputDirectory, path.Join(outputDirectory, fileName), nil
}

func PrepareReceive(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request) {
	// TODO: add ability to check for existing file with matching size and remove it from returned file list
	pin := r.URL.Query().Get("pin")

	var req model.PrepareReceiveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requestLogger := slog.Default().With(slog.String("Alias", req.Info.Alias))

	requestLogger.Info("Received upload request")
	requestLogger.Debug(fmt.Sprintf("Request details: %v", req))

	if config.RequirePassword != "" && pin != config.RequirePassword {
		w.WriteHeader(http.StatusUnauthorized)

		if pin != "" {
			requestLogger.Warn("Invalid PIN provided, access denied")
		}

		return
	}

	sessionID := uuid.New().String()

	sessions[sessionID] = &model.ReceiveSession{
		Alias:          req.Info.Alias,
		TotalFiles:     0,
		CompletedFiles: 0,
		Files:          make(map[string]model.FileInfo),
	}

	files := make(map[string]string)
	for fileID, fileInfo := range req.Files {
		// Get output file path to see if the file already exists and is the same size
		_, filePath, err := getOutputFilePath(config, req.Info.Alias, fileInfo.FileName)
		if err != nil {
			slog.Error("Failed to get absolute output directory: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		// GetFileSize returns -1 if the file doesn't exist
		if util.GetFileSize(filePath) == fileInfo.Size {
			slog.Debug("File already exists and is same size, telling the sender to skip it",
				slog.String("FileName", fileInfo.FileName),
				slog.Int64("FileSize", fileInfo.Size),
			)
			continue
		}

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

	requestLogger.Info("Done processing upload request, ready to receive files", slog.String("SessionID", sessionID))
}

func ReceiveHandler(config model.LocalSendConfig, message model.BroadcastMessage, w http.ResponseWriter, r *http.Request, sessionCompleteCallback func(string)) {
	sessionID := r.URL.Query().Get("sessionId")
	fileID := r.URL.Query().Get("fileId")
	token := r.URL.Query().Get("token")

	receiveLogger := slog.Default().With(
		slog.String("SessionID", sessionID),
		slog.String("FileID", fileID),
		slog.String("Token", token),
	)

	// Verify request parameters
	if sessionID == "" || fileID == "" || token == "" {
		receiveLogger.Warn("Upload missing parameters")
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	session, ok := sessions[sessionID]
	if !ok {
		receiveLogger.Warn("Invalid session ID")
		http.Error(w, "Couldn't find sessionID "+sessionID, http.StatusBadRequest)
	}

	sessionLogger := receiveLogger.With(slog.String("Alias", session.Alias))

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
			sessionLogger.Warn("Not on the allowed list, access denied")
			return
		}
	}

	// Use fileID to get the file name
	fileEntry, ok := session.Files[fileID]
	if !ok {
		sessionLogger.Warn("Invalid file ID")
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}
	fileName := fileEntry.FileName

	// Get output directory and file path
	outputDirectory, filePath, err := getOutputFilePath(config, session.Alias, fileName)
	if err != nil {
		slog.Error("Failed to get absolute output directory: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	transferLogger := sessionLogger.With(
		slog.String("OutputDirectory", outputDirectory),
		slog.String("FileName", fileName),
	)

	transferLogger.Info("Beginning to receive file")

	// Create the folder if it does not exist
	err = os.MkdirAll(outputDirectory, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		transferLogger.Warn("Error creating directory: " + err.Error())
		return
	}

	// Create a file
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		transferLogger.Warn("Error creating file: " + err.Error())
		return
	}
	defer file.Close()

	buffer := make([]byte, 2*1024*1024) // 2MB buffer
	for {
		n, err := r.Body.Read(buffer)
		if err != nil && err != io.EOF {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			transferLogger.Warn("Error reading file: " + err.Error())
			return
		}
		if n == 0 {
			break
		}

		_, err = file.Write(buffer[:n])
		if err != nil {
			http.Error(w, "Failed to write file", http.StatusInternalServerError)
			transferLogger.Warn("Error writing file: " + err.Error())
			return

		}
	}

	session.CompletedFiles += 1

	transferLogger.Info("Finished receiving file")
	w.WriteHeader(http.StatusOK)

	if session.CompletedFiles == session.TotalFiles {
		slog.Info(
			fmt.Sprintf("Transfer session complete, transferred %d file(s)", session.CompletedFiles),
			slog.String("SessionID", sessionID),
			slog.String("FilesTransferred", fmt.Sprintf("%d", session.CompletedFiles)),
		)

		// remove the session from memory
		delete(sessions, sessionID)

		// the session has finished, execute the callback
		sessionCompleteCallback(outputDirectory)
	}

}
