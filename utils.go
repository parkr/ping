package main

import (
	"fmt"
	"net/http"

	"github.com/parkr/gossip/src/gossip/serializer"
)

func writeJsonResponse(w http.ResponseWriter, json interface{}) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, serializer.MarshalJson(json))
}
