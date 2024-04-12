package jsv1

import (
	"fmt"
	"net/http"
	"strconv"
)

const statsJS = `
/**
 * Fetch ping stats and write in a site-customizable way.
 *
 * Caller must implement a function writePingStatsToHTML(document, data)
 * where document is the JS DOM document, and data is an object containing
 * the response from /counts, generally {views: Numeric, visitors: Numeric}.
 * More fields may be added to 'data' later.
 */

function fetchPingStats(document, callback) {
	var httpRequest = new XMLHttpRequest();
	httpRequest.onreadystatechange = () => {
		if (httpRequest.readyState === XMLHttpRequest.DONE) {
			if (httpRequest.status > 100 && httpRequest.status < 300) {
				const data = JSON.parse(httpRequest.responseText)
				callback(document, data)
			} else {
				console.error('There was a problem with the request.')
				console.error(httpRequest.status, httpRequest.responseText, httpRequest)
			}
		}
	};
	const countsURLSearchParams = new URLSearchParams()
	countsURLSearchParams.append('host', document.location.hostname)
	countsURLSearchParams.append('path', document.location.pathname)
	const countsURL = new URL('%s/counts')
	countsURL.search = "?" + countsURLSearchParams.toString()
	httpRequest.open('GET', countsURL.toString(), true);
	httpRequest.send();
}

(function(){
	document.addEventListener('readystatechange', (event) => {
		if (document.readyState === 'complete') {
			fetchPingStats(document, writePingStatsToHTML)
		}
	});
})()
`

func WriteStats(w http.ResponseWriter, pingBaseURL string) {
	w.Write([]byte(fmt.Sprintf(statsJS, pingBaseURL)))
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(statsJS)))
}
