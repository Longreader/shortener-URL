package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Longreader/go-shortener-url.git/internal/app/auth"
	"github.com/Longreader/go-shortener-url.git/internal/repository"
	"github.com/sirupsen/logrus"
)

func (h *Handler) APIDeleteUserURLsHandler(w http.ResponseWriter, r *http.Request) {

	user, err := auth.GetUser(r.Context())
	if err != nil {
		log.Printf("unable to parse user uid: %v", err)
		h.httpJSONError(w, "Server error", http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(r.Body)

	if err != nil || len(b) == 0 {
		log.Printf("error read request body: %v", err)
		h.httpJSONError(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ids := make([]repository.ID, 0)

	err = json.Unmarshal(b, &ids)
	if err != nil {
		h.httpJSONError(w, "Bad request", http.StatusBadRequest)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
		defer cancel()

		err := h.st.Delete(ctx, ids, user)
		if err != nil {
			logrus.Debug("Unable to delete", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusAccepted)
}
