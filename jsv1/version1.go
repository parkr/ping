package jsv1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/parkr/ping/dnt"
)

const returnedJavaScript = "(function(){})();"
const lengthOfJavaScript = "17"

func Write(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", lengthOfJavaScript)
	fmt.Fprintf(w, returnedJavaScript)
}

func DoNotTrack(w http.ResponseWriter) {
	Write(w, http.StatusOK)
	dnt.SetDoNotTrack(w)
}

func Error(w http.ResponseWriter, code int, err string) {
	w.WriteHeader(code)
	content := fmt.Sprintf(`(function(){console.error("%s")})();`, err)
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	fmt.Fprintf(w, content)
}
