package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var PORT = "4002"

const maxUploadSize = 10 * 1024 * 1024 // 5 mb

//	base URL của 3 môi trường
var baseLocalUrl = "http://localhost:" + PORT
var baseDevUrl = "https://upload.dev.rebateton.com"
var baseProdUrl = "https://upload.rebateton.com"

var env = os.Getenv("ENV")

type Response = struct {
	Url string `json:"url"`
}

type ErrorResponse = struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Code    int    `json:"code"`
}

var year, month, day = time.Now().Date()

var monthMapping = map[string]string{
	"January":   "1",
	"February":  "2",
	"March":     "3",
	"April":     "4",
	"May":       "5",
	"June":      "6",
	"July":      "7",
	"August":    "8",
	"September": "9",
	"October":   "10",
	"November":  "11",
	"December":  "12",
}

var yearStr = strconv.Itoa(year)
var monthStr = monthMapping[month.String()]
var dayStr = strconv.Itoa(day)

var uploadPath = "upload/" + yearStr + "/" + monthStr + "/" + dayStr
var meetePath = "meete/" + yearStr + "/" + monthStr + "/" + dayStr

func main() {
	fmt.Println("Listen on port....", PORT)

	http.HandleFunc("/up", uploadFileHandler())
	http.HandleFunc("/meete", uploadFileMeete())

	//	request rebateton file
	http.Handle("/files/", handleQueryFile())

	//	request Meete file
	http.Handle("/meete/", handleQueryMeeteFile())

	// log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading")
	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

func handleQueryFile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUri := r.RequestURI
		filePath := strings.TrimPrefix(requestUri, "/files")

		//w.Write(fs)
		http.ServeFile(w, r, "upload"+filePath)
	})
}

func handleQueryMeeteFile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUri := r.RequestURI
		filePath := strings.TrimPrefix(requestUri, "/")

		//w.Write(fs)
		http.ServeFile(w, r, filePath)
	})
}

func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//	config CORS
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}

		// check folder exist and create folder
		if _, err := os.Stat(uploadPath); err != nil {
			if os.IsNotExist(err) {
				// file does not exist
				createDirectory(uploadPath)
			} else {
				// other error
			}
		}
		// validate file size
		// fmt.Println("maxUploadSize", maxUploadSize)

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
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
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
		fileName := randToken(5)
		fileEndings, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}

		fullFileName := fileName + fileEndings[0]
		newPath := filepath.Join(uploadPath, fullFileName)
		//fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			//fmt.Println(err)
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}

		resultPath := strings.TrimPrefix(newPath, "upload")

		url := baseLocalUrl + "/files" + resultPath

		//	set url tùy môi trường
		if env == "development" {
			url = baseDevUrl + "/files" + resultPath
		} else if env == "production" {
			url = baseProdUrl + "/files" + resultPath
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

//upload file meete
func uploadFileMeete() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//auth := r.Header.Get("Authorization")
		//fmt.Println("auth.....", auth)

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
		// fmt.Println("maxUploadSize", maxUploadSize)

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
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		fileName := randToken(5)

		// check file type, detectcontenttype only needs the first 512 bytes
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

		fullFileName := fileName + fileEndings[0]
		newPath := filepath.Join(meetePath, fullFileName)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			//fmt.Println(err)
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
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

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	res := ErrorResponse{
		message,
		http.StatusBadRequest,
		http.StatusBadRequest,
	}
	js, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func createDirectory(dirName string) bool {
	src, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			panic(err)
		}
		return true
	}

	if src.Mode().IsRegular() {
		fmt.Println(dirName, "already exist as a file!")
		return false
	}

	return false
}
