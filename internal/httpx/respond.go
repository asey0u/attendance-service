package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Error struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func (e *Error) Error() string { return e.Message }

func NewError(status int, msg string) *Error {
	return &Error{Status: status, Message: msg}
}

var (
	ErrBadRequest   = NewError(http.StatusBadRequest, "bad request")
	ErrUnauthorized = NewError(http.StatusUnauthorized, "unauthorized")
	ErrForbidden    = NewError(http.StatusForbidden, "forbidden")
	ErrNotFound     = NewError(http.StatusNotFound, "not found")
	ErrConflict     = NewError(http.StatusConflict, "conflict")
	ErrInternal     = NewError(http.StatusInternalServerError, "internal server error")
)

func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func WriteError(w http.ResponseWriter, err error) {
	var herr *Error
	if errors.As(err, &herr) {
		JSON(w, herr.Status, herr)
		return
	}
	JSON(w, http.StatusInternalServerError, &Error{Message: err.Error()})
}

func ReadJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return NewError(http.StatusBadRequest, "invalid JSON body: "+err.Error())
	}
	return nil
}

func ParseIntParam(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, NewError(http.StatusBadRequest, "invalid id")
	}
	return n, nil
}
