package services

import (
	"net/http"
	"strings"
)

func HandleQueryFile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUri := r.RequestURI
		filePath := strings.TrimPrefix(requestUri, "/files")

		//w.Write(fs)
		w.Header().Set("Cache-Control", "max-age=604800")
		http.ServeFile(w, r, "upload"+filePath)
	})
}
