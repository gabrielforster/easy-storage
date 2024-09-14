package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// "path/filepath"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	// "github.com/gabrielforster/easy-storage/models"
	"github.com/gabrielforster/easy-storage/storage"
	// "github.com/gabrielforster/easy-storage/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize storage
	localStorage := storage.NewLocalStorage(os.Getenv("STORAGE_DIR"))

	// Initialize router
	r := mux.NewRouter()

	// Upload file endpoint
	r.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		key := r.FormValue("key")
		if key == "" {
			http.Error(w, "key is required", http.StatusBadRequest)
			return
		}

    metadata := make(map[string]string)
    for k, v := range r.Form {
      if strings.HasPrefix(k, "metadata-") {
				metadata[k[9:]] = v[0]
			}
    }

		err = localStorage.Upload(context.Background(), key, file, handler.Filename, metadata)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}).Methods("POST")

	// Download file endpoint
	r.HandleFunc("/download/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]

		file, err := localStorage.Download(context.Background(), key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))
		w.Header().Set("Content-Type", file.ContentType)

    for k, v := range file.Metadata {
      w.Header().Set(k, v)
    }

    bufReader := bytes.NewReader(file.Content)
		http.ServeContent(w, r, file.Name, file.LastModified, bufReader)
	}).Methods("GET")

	// Signed URL endpoint
	r.HandleFunc("/signed-url/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]

		signedURL, err := localStorage.GetSignedURL(context.Background(), key, 15*time.Minute)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(signedURL))
	}).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}

