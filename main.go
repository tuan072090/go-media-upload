package main

import (
	"fmt"
	"log"
	"net/http"
	"go-media-upload/services"
	config "go-media-upload/share"
)


func main() {
	fmt.Println("Listen on port....", config.PORT)

	http.HandleFunc("/upload", services.UploadFileHandler())
	http.Handle("/files/", services.HandleQueryFile())

	// log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading")
	log.Fatal(http.ListenAndServe(":"+config.PORT, nil))
}

