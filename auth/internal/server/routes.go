package server

import (
	"database/sql"
	"net/http"
)

func New(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	return Recovery(Logger(mux))
}
