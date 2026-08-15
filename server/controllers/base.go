package controllers

import (
	"github.com/ffyuhf/pmail/utils/context"
	"net/http"
)

type HandlerFunc func(*context.Context, http.ResponseWriter, *http.Request)
