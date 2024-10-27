package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func CallServer(uri string, body any) ([]byte, int) {
	slog.Debug(fmt.Sprintf("util.CallServer: Calling URL '%s'", uri))

	jsonStr, _ := json.Marshal(body)
	slog.Debug(fmt.Sprintf("util.CallServer: Sending JSON body: '%s'", string(jsonStr)))

	req, _ := http.NewRequest("POST", uri, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(fmt.Sprintf("util.CallServer: Error occurred sending request: %s", err.Error()))
		panic(err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	// fmt.Println("response Headers:", resp.Header)
	slog.Debug(fmt.Sprintf("util.CallServer: Response status '%s'", resp.Status))
	slog.Debug(fmt.Sprintf("util.CallServer: Response body '%s'", string(responseBody)))

	return responseBody, resp.StatusCode
}
