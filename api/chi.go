package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/d0ssan/CRUD-MariaDB-MongoDB/model"
	"github.com/d0ssan/CRUD-MariaDB-MongoDB/service"
)

const timeout = 15

// Router is the connection to service.Conn.
type Router struct {
	Service service.Conn
}

// Handlers creates a router and collect all handle function.
func Handlers(a Router) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(timeout * time.Second))
	r.Route("/users", func(r chi.Router) {
		r.Post("/", a.Insert)
		r.Get("/", a.All)
		r.Get("/{id}", a.ByID)
		r.Put("/{id}", a.Update)
		r.Delete("/{id}", a.Delete)
	})

	return r
}

// ByID gets a user info addressing to the database
// and responses with the requested data AND row id.
func (a Router) ByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := a.Service.DB.User(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeResp(w, http.StatusOK, respBytes)
}

// Insert parses a request body then sends it to the database.
// and responses with the stored data with row's id.
func (a Router) Insert(w http.ResponseWriter, r *http.Request) {
	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.Service.DB.Insert(context.Background(), &u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	uBytes, err := json.Marshal(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeResp(w, http.StatusCreated, uBytes)
}

// Update replaces the existing data to the requested one and responses with the new data.
func (a Router) Update(w http.ResponseWriter, r *http.Request) {
	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u.ID = int64(id)

	err = a.Service.DB.Update(context.Background(), u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeResp(w, http.StatusOK, nil)
}

// Delete removes a row from a database. This method handle id from the url parameter "users/{id}".
func (a Router) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = a.Service.DB.Delete(context.Background(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeResp(w, http.StatusOK, nil)
}

// All parse all rows from the database and shows them to the user with StatusOk.
// Otherwise, there will error messages from the db and StatusInternalServerError.
func (a Router) All(w http.ResponseWriter, _ *http.Request) {
	resp, err := a.Service.DB.All(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeResp(w, http.StatusOK, respBytes)
}

func writeResp(w http.ResponseWriter, status int, data []byte) {
	if data == nil {
		w.WriteHeader(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
