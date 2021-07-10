package controller

import (
	"database/sql"
	"encoding/json"
	"errors"
	"golauth/model"
	"net/http"
)

func sendServerError(w http.ResponseWriter, err error) {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusInternalServerError
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(e)
}

func sendBadRequest(w http.ResponseWriter, err error) {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusBadRequest
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(e)
}

func sendNotFound(w http.ResponseWriter, err error) {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusNotFound
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(e)
}

func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func sendResult(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendNotFound(w, err)
		} else {
			sendServerError(w, err)
		}
		return
	}

	sendSuccess(w, data)
}
