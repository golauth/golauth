package util

import (
	"database/sql"
	"encoding/json"
	"golauth/model"
	"net/http"
)

func SendServerError(w http.ResponseWriter, err error) {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusInternalServerError
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(e)
}

func SendBadRequest(w http.ResponseWriter, err error) {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusBadRequest
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(e)
}

func SendNotFound(w http.ResponseWriter, err error) {
	var e model.Error
	e.Message = err.Error()
	e.StatusCode = http.StatusNotFound
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(e)
}

func SendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func SendError(w http.ResponseWriter, err *model.Error) {
	if err.StatusCode == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(err.StatusCode)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(err)
}

func SendResult(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		if err == sql.ErrNoRows {
			SendNotFound(w, err)
		} else {
			SendServerError(w, err)
		}
		return
	}

	SendSuccess(w, data)
}
