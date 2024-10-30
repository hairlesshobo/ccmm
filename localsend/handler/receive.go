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
package handler

import (
	"encoding/json"
	"fmt"
	"gim/localsend/model"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	sessionIDCounter = 0
	sessionMutex     sync.Mutex
	fileNames        = make(map[string]string) // To save the file names
)

func PrepareReceive(w http.ResponseWriter, r *http.Request) {
	var req model.PrepareReceiveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Println("Received request:", req)

	sessionMutex.Lock()
	sessionIDCounter++
	sessionID := fmt.Sprintf("session-%d", sessionIDCounter)
	sessionMutex.Unlock()

	files := make(map[string]string)
	for fileID, fileInfo := range req.Files {
		token := fmt.Sprintf("token-%s", fileID)
		files[fileID] = token

		// Save file name
		fileNames[fileID] = fileInfo.FileName
	}

	resp := model.PrepareReceiveResponse{
		SessionID: sessionID,
		Files:     files,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func ReceiveHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	fileID := r.URL.Query().Get("fileId")
	token := r.URL.Query().Get("token")

	// Verify request parameters

	if sessionID == "" || fileID == "" || token == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	// Use fileID to get the file name
	fileName, ok := fileNames[fileID]
	if !ok {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// TODO: fix this assumption
	// Generate file paths, preserving file extensions
	filePath := filepath.Join("uploads", fileName)

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

	fmt.Println("Saved file:", filePath)
	w.WriteHeader(http.StatusOK)

}

// ReceiveHandler Handling file download requests
func DownloadRequestHandler(w http.ResponseWriter, r *http.Request) {
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
