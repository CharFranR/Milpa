package handler

import (
	"net/http"

	"milpa/infrastructure/adapters/primary/api/httpx"
)

func respond(w http.ResponseWriter, status int, data any) {
	httpx.Respond(w, status, data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	httpx.RespondError(w, status, msg)
}

func statusCode(err error) int {
	return httpx.StatusCode(err)
}

func handleError(w http.ResponseWriter, err error) {
	httpx.HandleError(w, err)
}
