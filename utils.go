package ping

import (
	"encoding/json"
	"net/http"
	"strings"
)

func writeJsonResponse(w http.ResponseWriter, input interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(input)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
	} else {
		w.Write(data)
	}
}

func jsonError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	// Note: All w.Header() modifications must be made BEFORE this call.
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func sanitizeUserInput(input string) string {
	escapedInput := strings.ReplaceAll(input, "\n", "")
	escapedInput = strings.ReplaceAll(escapedInput, "\r", "")
	return escapedInput
}
