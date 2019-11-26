package main

import (
	"fmt"
	"log"
	"net/http"
	"rt-media-upload/handlers"
)


func main() {
	fmt.Println("Listen on port....", handlers.PORT)

	http.HandleFunc("/up", handlers.UploadFileHandler())
	http.HandleFunc("/meete", handlers.UploadFileMeete())

	//	request rebateton file
	http.Handle("/files/", handlers.HandleQueryFile())

	//	request Meete file
	http.Handle("/meete/", handlers.HandleQueryMeeteFile())

	// log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading")
	log.Fatal(http.ListenAndServe(":"+handlers.PORT, nil))
}

