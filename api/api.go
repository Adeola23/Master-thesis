package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	listenAddr string
}


func NewServer(listenAddr string) *Server {

	return &Server{
		listenAddr: listenAddr,
	}

}

type apiFunc func (w http.ResponseWriter, r *http.Request) error

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		if err := f(w,r); err != nil {
			JSON(w, http.StatusBadRequest, map[string] any{"error": err.Error()})
		}
	}
}

func JSON (w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func (s *Server) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/ready", makeHttpHandleFunc(s.handlePeerReady))

	http.ListenAndServe(s.listenAddr,r)

}

func (s *Server) handlePeerReady(w http.ResponseWriter, r *http.Request) error {
	return nil
}