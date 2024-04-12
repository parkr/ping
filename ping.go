package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/parkr/ping/analytics"
	"github.com/parkr/ping/database"
	"github.com/parkr/ping/dnt"
	"github.com/parkr/ping/jsv1"
	"github.com/parkr/ping/jsv2"
)

var (
	allowedHosts  map[string]bool
	hostAllowlist = flag.String("hosts", "", "The hosts allowed to use this service. Comma-separated.")
)

var db *sqlx.DB

func allowedHost(host string) bool {
	if hostAllowlist == nil || *hostAllowlist == "" {
		return true
	}

	if allowedHosts == nil || len(allowedHosts) == 0 {
		allowedHostsList := strings.Split(*hostAllowlist, ",")
		allowedHosts = make(map[string]bool, len(allowedHostsList))
		for _, allowedHost := range allowedHostsList {
			allowedHosts[allowedHost] = true
		}
	}

	_, ok := allowedHosts[host]
	return ok
}

// ping routes to pingv1 or pingv2 depending on the version code in the form.
func ping(w http.ResponseWriter, r *http.Request) {
	version := r.FormValue("v")
	switch version {
	case "2":
		pingv2(w, r)
	default:
		pingv1(w, r)
	}
}

// pingv1 implements the referer-based logging.
// When a request comes in, the referer and remote IP (or X-Forwarded-For)
// are used to write the ping entry.
func pingv1(w http.ResponseWriter, r *http.Request) {
	if dnt.RequestsDoNotTrack(r) {
		log.Println("dnt requested")
		jsv1.DoNotTrack(w)
		return
	}

	referrer := r.Referer()
	if referrer == "" {
		log.Println("empty referrer")
		jsv1.Error(w, http.StatusBadRequest, "empty referrer")
		return
	}

	url, err := url.Parse(referrer)

	if err != nil {
		log.Println("invalid referrer:", sanitizeUserInput(referrer))
		jsv1.Error(w, http.StatusInternalServerError, "Couldn't parse referrer: "+err.Error())
		return
	}

	if !allowedHost(url.Host) {
		log.Println("unauthorized host:", sanitizeUserInput(url.Host))
		jsv1.Error(w, http.StatusUnauthorized, "unauthorized host")
		return
	}

	var ip string
	if res := r.Header.Get("X-Forwarded-For"); res != "" {
		log.Println("Fetching IP from proxy:", sanitizeUserInput(res))
		ip = res
	} else {
		ip = r.RemoteAddr
	}

	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		log.Println("empty user-agent")
		jsv1.Error(w, http.StatusBadRequest, "empty user-agent")
		return
	}

	visit := &database.Visit{
		IP:        ip,
		Host:      url.Host,
		Path:      url.Path,
		UserAgent: userAgent,
		CreatedAt: time.Now().UTC().Format(database.SQLDateTimeFormat),
	}
	log.Println("Logging visit:", sanitizeUserInput(visit.String()))

	err = visit.Save(db)

	if err != nil {
		log.Println("Error saving to db:", err)
		jsv1.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	jsv1.Write(w, http.StatusCreated)
}

// pingv2 implements the js-based logging.
// When a request comes in, it returns JS that will call /submit.js to capture
// the full path of the page visited.
func pingv2(w http.ResponseWriter, r *http.Request) {
	if dnt.RequestsDoNotTrack(r) {
		log.Println("dnt requested")
		jsv1.DoNotTrack(w)
		return
	}

	referrer := r.Referer()
	if referrer == "" {
		log.Println("empty referrer")
		jsv1.Error(w, http.StatusBadRequest, "empty referrer")
		return
	}

	url, err := url.Parse(referrer)

	if err != nil {
		log.Println("invalid referrer:", sanitizeUserInput(referrer))
		jsv1.Error(w, http.StatusInternalServerError, "Couldn't parse referrer: "+err.Error())
		return
	}

	if !allowedHost(url.Host) {
		log.Println("unauthorized host:", sanitizeUserInput(url.Host))
		jsv1.Error(w, http.StatusUnauthorized, "unauthorized host")
		return
	}

	jsv2.Write(w, http.StatusOK)
}

// submitv2 takes an XHR request with the host & path in the form and rewrites
// as a pingv1 request using the referer.
func submitv2(w http.ResponseWriter, r *http.Request) {
	host := r.FormValue("host")
	path := r.FormValue("path")
	if host == "" || path == "" {
		jsv1.Error(w, http.StatusBadRequest, "missing param")
		return
	}
	referer := url.URL{Host: host, Path: path}

	req, err := http.NewRequest(http.MethodGet, "/ping.js", nil)
	if err != nil {
		jsv1.Error(w, http.StatusInternalServerError, "unable to rewrite")
		return
	}
	req.Header.Set("Referer", referer.String())
	req.Header.Set("User-Agent", r.Header.Get("User-Agent"))
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	log.Printf("forwarding v2 to v1")

	pingv1(w, req)
}

func counts(w http.ResponseWriter, r *http.Request) {
	host := r.FormValue("host")
	path := r.FormValue("path")

	var err error
	var views, visitors int
	if host == "" || path == "" {
		http.Error(w, "Missing param", 400)
	} else {
		views, err = analytics.ViewsForHostPath(db, host, path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		visitors, err = analytics.VisitorsForHostPath(db, host, path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		addCorsHeaders(w, r)
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

		addCorsHeaders(w, r)
		writeJsonResponse(w, map[string][]string{"entries": entries})
	} else {
		http.Error(w, "Missing param", 400)
		return
	}
}

func health(w http.ResponseWriter, r *http.Request) {
	if db == nil {
		http.Error(w, "error initializing db", http.StatusInternalServerError)
		return
	}

	if err := db.Ping(); err != nil {
		http.Error(w, "error pinging db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "healthy")
}

func buildHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/_health", health)
	mux.HandleFunc("/ping", ping)
	mux.HandleFunc("/ping.js", ping)
	mux.HandleFunc("/submit", submitv2)
	mux.HandleFunc("/submit.js", submitv2)
	mux.Handle("/counts", &corsHandler{counts})
	mux.Handle("/all", &corsHandler{all})
	return mux
}

func main() {
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8000"
	}

	var binding string
	flag.StringVar(&binding, "http", ":"+defaultPort, "The IP/port to bind to.")
	flag.Parse()

	var err error
	db, err = database.Initialize()
	if err != nil {
		log.Fatalf("unable to initialize db: %+v", err)
	}

	http.Handle("/", buildHandler())

	log.Println("Listening on", binding, "...")
	log.Fatal(http.ListenAndServe(binding, nil))
}
