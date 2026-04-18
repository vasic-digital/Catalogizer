// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package server implements the OCU CUDA sidecar gRPC services.
//
// Build tags select the backend:
//
//	go build -tags=cuda   ./...   — CUDA/OpenCV backend (requires gocv + CUDA)
//	go build              ./...   — stub backend (unit-testable, no GPU deps)
//
// The Server type is the same in both builds; only the method bodies differ.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	pb "digital.vasic.ocu-cuda-sidecar/proto"
)

// ---------------------------------------------------------------------------
// Session registry (shared across both backends)
// ---------------------------------------------------------------------------

type nvencSession struct {
	width       int32
	height      int32
	framerate   int32
	bitrateKbps int32
}

// Server is the gRPC service implementation. Embed it in a net/http
// handler (using the JSON-over-HTTP shim below) or wire it to a real
// gRPC server once protoc-generated stubs are available.
type Server struct {
	mu       sync.Mutex
	sessions map[string]*nvencSession
}

// New creates a Server ready to handle requests.
func New() *Server {
	return &Server{
		sessions: make(map[string]*nvencSession),
	}
}

// ---------------------------------------------------------------------------
// HTTP/JSON shim — routes /nvenc/* and /opencv/* to service methods.
// The shim lets the sidecar be callable without a full gRPC client; a
// real gRPC server can be layered on top once .pb.go files are generated.
// ---------------------------------------------------------------------------

// RegisterRoutes mounts all OCU endpoints onto mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/nvenc/open", s.handleNVENCOpen)
	mux.HandleFunc("/nvenc/close", s.handleNVENCClose)
	mux.HandleFunc("/opencv/analyze", s.handleAnalyze)
	mux.HandleFunc("/opencv/diff", s.handleDiff)
	mux.HandleFunc("/opencv/match", s.handleMatch)
	mux.HandleFunc("/opencv/ocr", s.handleOCR)
}

// ListenAndServe starts the HTTP server on addr and blocks until ctx
// is cancelled or an unrecoverable error occurs.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)
	srv := &http.Server{
		Addr:        addr,
		Handler:     mux,
		ReadTimeout: 60 * time.Second,
		// WriteTimeout is intentionally large — Encode is a long stream.
		WriteTimeout: 10 * time.Minute,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("server: listen %s: %w", addr, err)
	}
	log.Printf("ocu-cuda-sidecar: listening on %s", ln.Addr())
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()
	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// ---------------------------------------------------------------------------
// NVENC endpoints
// ---------------------------------------------------------------------------

func (s *Server) handleNVENCOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb.OpenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.NVENCOpen(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func (s *Server) handleNVENCClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb.CloseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.NVENCClose(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

// ---------------------------------------------------------------------------
// OpenCV endpoints
// ---------------------------------------------------------------------------

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateFrameDims(req.Width, req.Height, req.BGRA); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.Analyze(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb.DiffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateFrameDims(req.Width, req.Height, req.Before); err != nil {
		http.Error(w, "before: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateFrameDims(req.Width, req.Height, req.After); err != nil {
		http.Error(w, "after: "+err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.Diff(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func (s *Server) handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.Match(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

func (s *Server) handleOCR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb.OCRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := s.OCR(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("ocu-cuda-sidecar: write response: %v", err)
	}
}

func validateFrameDims(w, h int32, data []byte) error {
	if w <= 0 || h <= 0 {
		return fmt.Errorf("invalid dimensions %dx%d", w, h)
	}
	expected := int(w) * int(h) * 4
	if len(data) != expected {
		return fmt.Errorf("data length %d does not match %dx%d*4=%d", len(data), w, h, expected)
	}
	return nil
}

// encodeStream is a read-closer that the NVENC encode endpoint drains.
// Kept here so both backends can share the shim logic.
func drainEncodeStream(_ io.Reader) ([][]byte, error) {
	// Streaming encode is handled by the backend-specific Encode method.
	// This placeholder exists so the shared HTTP shim can reference it.
	return nil, nil
}
