package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func writeJsonResponse(w http.ResponseWriter, input interface{}) {
	w.Header().Set("Content-Type", "application/json")
	addCorsHeaders(w)
	data, err := json.Marshal(input)
	if err != nil {
		fmt.Fprintf(w, `{"error":"json, `+err.Error()+`"}`)
	} else {
		w.Write(data)
	}
}
