package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (app *App) uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer file.Close()

	contentType := "application/binary"
	if len(handler.Header["Content-Type"]) > 0 {
		contentType = handler.Header["Content-Type"][0]
	}

	bucketName := DEFAULT_BUCKET

	payload := &BlobData{
		Filename:    handler.Filename,
		ContentType: contentType,
		Size:        handler.Size,
		Bucket:      bucketName,
	}

	resp, err := app.PutBlob(context.Background(), payload, file)
	if err != nil {
		log.Println(err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, resp)
}

func (app *App) getFile(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) > 0 {
	}
	respondError(w, http.StatusInternalServerError, err.Error())

}
