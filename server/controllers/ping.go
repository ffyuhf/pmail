package controllers

import (
	"github.com/ffyuhf/pmail/dto/response"
	"net/http"
)

func Ping(w http.ResponseWriter, req *http.Request) {
	response.NewSuccessResponse("pong").FPrint(w)
}
