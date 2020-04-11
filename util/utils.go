package util

import (
	"database/sql"
	"encoding/json"
	"golauth/model"
	"log"
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

func SendError(w http.ResponseWriter, err model.Error) {
	w.WriteHeader(err.StatusCode)
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

func ResultData(data interface{}, z interface{}, err error) (interface{}, error) {
	if err != nil {
		return z, err
	}

	return data, nil
}

func ResultSliceString(data []string, z []string, err error) ([]string, error) {
	if err != nil {
		return z, err
	}

	return data, nil
}

func LogFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func LogError(err error) {
	if err != nil {
		log.Println("ERROR: ", err)
	}
}

func sendError(w http.ResponseWriter, statusCode int, message string) {
	var e model.Error
	e.Message = message
	e.StatusCode = statusCode
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(e)
}
