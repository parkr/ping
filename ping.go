package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/parkr/ping/analytics"
	"github.com/zenazn/goji"
)

const returnedJavaScript = "(function(){})();"
const lengthOfJavaScript = "17"

func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", lengthOfJavaScript)
	fmt.Fprintf(w, returnedJavaScript)

	referrer := r.Referer()
	if referrer == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url, err := url.Parse(referrer)

	if err != nil {
		http.Error(w, "Couldn't parse "+referrer+": "+err.Error(), 500)
		return
	}

	var ip string
	if res := r.Header.Get("X-Forwarded-For"); res != "" {
		log.Println("Fetching IP from proxy:", res)
		ip = res
	} else {
		ip = r.RemoteAddr
	}

	visit := &Visit{
		IP:        ip,
		Host:      url.Host,
		Path:      url.Path,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	log.Println("Logging visit:", visit.String())

	err = visit.Save()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func counts(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")

	var err error
	var views, visitors int
	if path == "" {
		http.Error(w, "Missing param", 400)
	} else {
		views, err = analytics.ViewsForPath(db, path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		visitors, err = analytics.VisitorsForPath(db, path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		writeJsonResponse(w, map[string]int{
			"views":    views,
			"visitors": visitors,
		})
	}
}

func all(w http.ResponseWriter, r *http.Request) {
	thing := r.FormValue("type")

	if thing == "path" || thing == "host" {
		entries, err := analytics.ListDistinctColumn(db, thing)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		writeJsonResponse(w, map[string][]string{"entries": entries})
	} else {
		http.Error(w, "Missing param", 400)
		return
	}
}

func main() {
	goji.Get("/ping", ping)
	goji.Get("/ping.js", ping)
	goji.Get("/counts", counts)
	goji.Get("/all", all)
	goji.Serve()
}
