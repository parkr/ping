package jsv1

import (
	"fmt"
	"net/http"
	"strconv"
)

const returnedJavaScript = "(function(){})();"
const lengthOfJavaScript = "17"

func Write(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", lengthOfJavaScript)
	// Note: All w.Header() modifications must be made BEFORE this call.
	w.WriteHeader(code)
	fmt.Fprintf(w, returnedJavaScript)
}

func Error(w http.ResponseWriter, code int, err string) {
	content := fmt.Sprintf(`(function(){console.error("%s")})();`, err)
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	// Note: All w.Header() modifications must be made BEFORE this call.
	w.WriteHeader(code)
	fmt.Fprintf(w, content)
}
