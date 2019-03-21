package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	url := fmt.Sprintf("http://127.0.0.1:%s/_health", port)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("error from %q: %#v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}
}
