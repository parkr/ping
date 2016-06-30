package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/parkr/ping/analytics"
	"github.com/parkr/ping/database"
)

var (
	allowedHosts []string
	whitelist    = flag.String("hosts", "", "The hosts allowed to use this service. Comma-separated.")
)

const returnedJavaScript = "(function(){})();"
const lengthOfJavaScript = "17"

var db = database.InitializeDatabase()

func javascriptRespond(w http.ResponseWriter, code int, err string) {
	w.WriteHeader(code)

	var content string
	if err == "" {
		content = returnedJavaScript
	} else {
		content = fmt.Sprintf(`(function(){console.error("%s")})();`, err)
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	fmt.Fprintf(w, content)
}

func allowedHost(host string) bool {
	if whitelist == nil || *whitelist == "" {
		return true
	}

	if len(allowedHosts) == 0 {
		allowedHosts = strings.SplitN(*whitelist, ",", -1)
	}

	for _, allowed := range allowedHosts {
		if allowed == host {
			return true
		}
	}
	return false
}

func ping(w http.ResponseWriter, r *http.Request) {
	referrer := r.Referer()
	if referrer == "" {
		log.Println("empty referrer")
		javascriptRespond(w, http.StatusBadRequest, "empty referrer")
		return
	}

	url, err := url.Parse(referrer)

	if err != nil {
		log.Println("invalid referrer:", referrer)
		javascriptRespond(w, 500, "Couldn't parse referrer: "+err.Error())
		return
	}

	if !allowedHost(url.Host) {
		log.Println("unauthorized host:", url.Host)
		javascriptRespond(w, 403, "love the host, except noooope.")
		return
	}

	var ip string
	if res := r.Header.Get("X-Forwarded-For"); res != "" {
		log.Println("Fetching IP from proxy:", res)
		ip = res
	} else {
		ip = r.RemoteAddr
	}

	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		log.Println("empty user-agent")
		javascriptRespond(w, http.StatusBadRequest, "empty user-agent")
		return
	}

	visit := &database.Visit{
		IP:        ip,
		Host:      url.Host,
		Path:      url.Path,
		UserAgent: userAgent,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	log.Println("Logging visit:", visit.String())

	err = visit.Save(db)

	if err != nil {
		javascriptRespond(w, 500, err.Error())
		return
	}

	javascriptRespond(w, 201, "")
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
	flag.Parse()

	http.HandleFunc("/ping", ping)
	http.HandleFunc("/ping.js", ping)
	http.HandleFunc("/counts", counts)
	http.HandleFunc("/all", all)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
