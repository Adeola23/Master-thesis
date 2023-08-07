package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
}


func NewAPIServer(listenAddr string) *APIServer {

	return &APIServer{
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

func (s *APIServer) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/ready/{addr}", makeHttpHandleFunc(s.handlePeerReady))

	http.ListenAndServe(s.listenAddr,r)

}

func (s *APIServer) handlePeerReady(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	JSON(w, http.StatusOK, vars)
	return nil
}