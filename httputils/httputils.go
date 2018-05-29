package httputils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"errors"

	"github.com/snap10/go-snaplib/logging"
)

type (
	//appError
	appError struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Status  int    `json:"status"`
	}
	//error Resource
	errorResource struct {
		Data appError `json:"data"`
	}
)

func DisplayAppError(w http.ResponseWriter, handlerError error, message string, code int) {
	if handlerError == nil {
		handlerError = errors.New("")
	}

	errObj := appError{
		Error:   handlerError.Error(),
		Message: message,
		Status:  code,
	}

	logging.Error.Printf("[AppError]: %s\n, %s\n", handlerError, message)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if j, err := json.Marshal(errorResource{Data: errObj}); err != nil {
		logging.Error.Printf("[ServiceError]: Could not send error response")
	} else {
		w.Write(j)

	}
}

func SendAsDataJsonWithCache(w http.ResponseWriter, dataobject interface{}, message string, statuscode int, cacheSeconds int, errHandler func(http.ResponseWriter, error, string, int)) {
	w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(cacheSeconds))
	SendAsDataJson(w, dataobject, message, statuscode, errHandler)
}

func SendAsJsonWithCache(w http.ResponseWriter, dataobject interface{}, statuscode int, cacheSeconds int, errHandler func(http.ResponseWriter, error, string, int)) {
	w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(cacheSeconds))
	SendAsJson(w, dataobject, statuscode, errHandler)
}

func SendAsJson(w http.ResponseWriter, dataobject interface{}, statuscode int, errHandler func(http.ResponseWriter, error, string, int)) {
	j, err := json.Marshal(dataobject)
	if err != nil {
		errHandler(w, err, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)
	w.Write(j)
}

type DataResource struct {
	Message string `json:"message"`
	Data    []byte `json:"data"`
}

func SendAsDataJson(w http.ResponseWriter, dataobject interface{}, message string, statuscode int, errHandler func(http.ResponseWriter, error, string, int)) {
	j, err := json.Marshal(dataobject)
	if err != nil {
		errHandler(w, err, "Error marshalling json", http.StatusInternalServerError)
	}
	dataResource := DataResource{
		Message: message,
		Data:    j,
	}
	dataJson, err := json.Marshal(dataResource)
	if err != nil {
		errHandler(w, err, "Error marshalling json", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)
	w.Write(dataJson)
}
