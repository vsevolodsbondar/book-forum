package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func decodeRequest(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			return err
		}
		return ErrInvalidRequest
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return ErrInvalidRequest
	}

	return nil
}
