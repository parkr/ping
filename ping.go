package ping

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/parkr/ping/analytics"
	"github.com/parkr/ping/cors"
	"github.com/parkr/ping/database"
	"github.com/parkr/ping/dnt"
	"github.com/parkr/ping/jsv1"
	"github.com/parkr/ping/jsv2"
)

const xForwardedForHeaderName = "X-Forwarded-For"

var db *sqlx.DB

func Initialize(connection string) error {
	var err error
	db, err = database.Initialize(connection)
	return err
}

func parseReferer(referer string) (*url.URL, error) {
	if referer == "" {
		return nil, errors.New("referer is empty")
	}

	return url.Parse(referer)
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
	parsedReferer, err := parseReferer(r.Referer())
	if err != nil {
		log.Printf("referer invalid (%q): %v", sanitizeUserInput(r.Referer()), err)
		jsv1.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	var ip string
	if res := r.Header.Get(xForwardedForHeaderName); res != "" {
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
		IP:        sanitizeUserInput(ip),
		Host:      sanitizeUserInput(parsedReferer.Host),
		Path:      sanitizeUserInput(parsedReferer.Path),
		UserAgent: sanitizeUserInput(userAgent),
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
func pingv2(w http.ResponseWriter, _ *http.Request) {
	jsv2.Write(w, http.StatusOK)
}

type submitv2Handler struct {
	nextHandler http.Handler
}

// submitv2 takes an XHR request with the host & path in the form and rewrites
// as a pingv1 request using the referer.
func (s submitv2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.FormValue("host")
	path := r.FormValue("path")
	if host == "" || path == "" {
		log.Printf("host=%s path=%s", host, path)
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

	remoteAddr := r.Header.Get(xForwardedForHeaderName)
	if remoteAddr == "" {
		remoteAddr = r.RemoteAddr
	}
	req.Header.Set(xForwardedForHeaderName, remoteAddr)

	log.Printf("forwarding v2 to v1")

	s.nextHandler.ServeHTTP(w, req)
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

type statsHandler struct {
	pingBaseURL string
}

func (s statsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jsv1.WriteStats(w, s.pingBaseURL)
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

func NewHandler(allowedHosts []string, pingBaseURL string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/_health", health)
	pingHandler := dnt.NewMiddleware(
		NewHostAuthMiddleware(allowedHosts,
			http.HandlerFunc(ping)))
	mux.Handle("/ping", pingHandler)
	mux.Handle("/ping.js", pingHandler)
	submitHandler := cors.NewMiddleware(allowedHosts,
		dnt.NewMiddleware(
			NewHostAuthMiddleware(allowedHosts,
				submitv2Handler{pingHandler})))
	mux.Handle("/submit", submitHandler)
	mux.Handle("/submit.js", submitHandler)
	mux.Handle("/counts", cors.NewMiddleware(allowedHosts, http.HandlerFunc(counts)))
	mux.Handle("/all", cors.NewMiddleware(allowedHosts, http.HandlerFunc(all)))
	mux.Handle("/stats.js", cors.NewMiddleware(allowedHosts, statsHandler{pingBaseURL}))
	return mux
}
