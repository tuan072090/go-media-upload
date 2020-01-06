package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

func HandleQueryMeeteFile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUri := r.RequestURI
		filePath := strings.TrimPrefix(requestUri, "/")

		fmt.Println("meete filePath....", filePath)

		//w.Write(fs)
		w.Header().Set("Cache-Control", "public,max-age=604800")
		http.ServeFile(w, r, filePath)
	})
}
