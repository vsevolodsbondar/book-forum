package handler

import "net/http"

// Logout revokes the presented session without revealing its state.
func (h *SessionHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.sessionService.Logout(r.Context(), token); err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
