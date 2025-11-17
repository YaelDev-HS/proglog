package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type httpServer struct {
	Log *Log
}

func newHTTPServer() *httpServer {
	return &httpServer{
		Log: NewLog(),
	}
}

type ProduceRequest struct {
	Record Record `json:"record"`
}
type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}
type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}
type ConsumeResponse struct {
	Record Record `json:"record"`
}

func NewHttpServer(addr string) *http.Server {
	srv := newHTTPServer()

	r := mux.NewRouter()
	r.HandleFunc("/", srv.handleProducer).Methods("POST")
	r.HandleFunc("/", srv.handleConsume).Methods("GET")

	return &http.Server{
		Addr:    ":" + addr,
		Handler: r,
	}
}

func (h *httpServer) handleProducer(w http.ResponseWriter, r *http.Request) {
	var req ProduceRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	off, err := h.Log.Append(req.Record)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ProduceResponse{Offset: off}
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := h.Log.Read(req.Offset)

	if err != nil {
		if errors.Is(err, ErrOffsetNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ConsumeResponse{Record: record}
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
