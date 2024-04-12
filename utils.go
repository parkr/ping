package ping

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func writeJsonResponse(w http.ResponseWriter, input interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(input)
	if err != nil {
		fmt.Fprintf(w, `{"error":"json, `+err.Error()+`"}`)
	} else {
		w.Write(data)
	}
}

func sanitizeUserInput(input string) string {
	escapedInput := strings.ReplaceAll(input, "\n", "")
	escapedInput = strings.ReplaceAll(escapedInput, "\r", "")
	return escapedInput
}
