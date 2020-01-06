package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var PORT = "4002"

const maxUploadSize = 10 * 1024 * 1024 // 10 mb

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

var meetePath = "meete/" + yearStr + "/" + monthStr + "/" + dayStr
var uploadPath = "upload/" + yearStr + "/" + monthStr + "/" + dayStr

func UploadFileHandler() http.HandlerFunc {
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


		// Open a new image.
		src, err := imaging.Open(newPath)
		if err != nil {
			renderError(w, "failed to open image",  http.StatusInternalServerError)
			return
		}

		defer newFile.Close() // idempotent, okay to call twice
		resultPath := strings.TrimPrefix(newPath, "upload")

		//	resize if bigger than 2048
		if src.Bounds().Max.X > 2148 {
			src = imaging.Resize(src, 2048, 0, imaging.Lanczos)

			// Save the resulting image.
			err = imaging.Save(src, newPath)
			if err != nil {
				renderError(w, "failed to save image",  http.StatusInternalServerError)
				return
			}
		}


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
