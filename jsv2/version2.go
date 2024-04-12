package jsv2

import (
	"fmt"
	"net/http"
	"strconv"
)

const returnedJavaScript = `
function logVisit(document) {
	var httpRequest = new XMLHttpRequest();
	httpRequest.onreadystatechange = () => {
		if (httpRequest.readyState === XMLHttpRequest.DONE) {
			if (httpRequest.status > 100 && httpRequest.status < 300) {
				console.log("visit log result:", httpRequest.responseText)
			} else {
				console.error('There was a problem with the request.')
				console.error(httpRequest.status, httpRequest.responseText, httpRequest)
			}
		}
	};
	const visitSearchParams = new URLSearchParams()
	visitSearchParams.append('host', document.location.hostname)
	visitSearchParams.append('path', document.location.pathname)
	const visitURL = new URL('https://ping.parkermoo.re/submit.js')
	visitURL.search = "?" + visitSearchParams.toString()
	httpRequest.open('POST', visitURL.toString(), true);
	httpRequest.send();
}
(function(){
	document.addEventListener('readystatechange', (event) => {
		if (document.readyState === 'complete') {
			logVisit(document)
		}
	});
})()
`

func Write(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(returnedJavaScript)))
	// Note: All w.Header() modifications must be made BEFORE this call.
	w.WriteHeader(code)
	fmt.Fprintf(w, returnedJavaScript)
}
