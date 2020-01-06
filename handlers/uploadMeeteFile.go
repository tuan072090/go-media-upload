package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

//upload file meete
func UploadFileMeete() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//auth := r.Header.Get("Authorization")
		fmt.Println("upload Meete.....")

		//	config CORS
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}

		// check folder exist and create folder
		if _, err := os.Stat(meetePath); err != nil {
			if os.IsNotExist(err) {
				// file does not exist
				createDirectory(meetePath)
			} else {
				// other error

			}
		}
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}

		// parse and validate file and post parameters
		file, _, err := r.FormFile("file")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}


		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		defer file.Close()

		// check file type, detect content type only needs the first 512 bytes
		detectedFileType := http.DetectContentType(fileBytes)
		switch detectedFileType {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
			// case "application/pdf":
			break
		default:
			renderError(w, "INVALID_FILE_TYPE", http.StatusBadRequest)
			return
		}

		fileEndings, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}

		fileName := randToken(5)

		fullFileName := fileName + fileEndings[0]
		newPath := filepath.Join(meetePath, fullFileName)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}

		//	save file
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}

		// Open a new image.
		src, err := imaging.Open(newPath)
		if err != nil {
			renderError(w, "failed to open image",  http.StatusInternalServerError)
			return
		}

		defer newFile.Close() // idempotent, okay to call twice

		//	resize if bigger than 1024
		if src.Bounds().Max.X > 1024 {
			src = imaging.Resize(src, 1024, 0, imaging.Lanczos)

			// Save the resulting image as JPEG.
			err = imaging.Save(src, newPath)
			if err != nil {
				renderError(w, "failed to save image",  http.StatusInternalServerError)
				return;
			}
		}

		url := baseLocalUrl + "/" + newPath

		//	set url tùy môi trường
		if env == "development" {
			url = baseDevUrl + "/" + newPath
		} else if env == "production" {
			url = baseProdUrl + "/" + newPath
		}

		//	set response payload
		res := Response{url}

		js, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
}
