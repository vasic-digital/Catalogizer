// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

//go:build !cuda

// Tests compile against the stub backend (no GPU required). The CUDA
// backend is exercised only on thinker.local inside the container.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "digital.vasic.ocu-cuda-sidecar/proto"
)

// ---------------------------------------------------------------------------
// NVENCOpen / NVENCClose
// ---------------------------------------------------------------------------

func TestNVENCOpen_Valid(t *testing.T) {
	s := New()
	resp, err := s.NVENCOpen(context.Background(), &pb.OpenRequest{
		Width: 1920, Height: 1080, Framerate: 30, BitrateKbps: 4000,
	})
	if err != nil {
		t.Fatalf("NVENCOpen: %v", err)
	}
	if resp.SessionID == "" {
		t.Fatal("expected non-empty session_id")
	}
}

func TestNVENCOpen_InvalidDimensions(t *testing.T) {
	s := New()
	cases := []pb.OpenRequest{
		{Width: 0, Height: 1080, Framerate: 30, BitrateKbps: 4000},
		{Width: 1920, Height: 0, Framerate: 30, BitrateKbps: 4000},
		{Width: -1, Height: 1080, Framerate: 30, BitrateKbps: 4000},
	}
	for _, c := range cases {
		if _, err := s.NVENCOpen(context.Background(), &c); err == nil {
			t.Errorf("expected error for %+v", c)
		}
	}
}

func TestNVENCOpen_InvalidFramerate(t *testing.T) {
	s := New()
	_, err := s.NVENCOpen(context.Background(), &pb.OpenRequest{
		Width: 1920, Height: 1080, Framerate: 0, BitrateKbps: 4000,
	})
	if err == nil {
		t.Fatal("expected error for framerate=0")
	}
}

func TestNVENCOpen_InvalidBitrate(t *testing.T) {
	s := New()
	_, err := s.NVENCOpen(context.Background(), &pb.OpenRequest{
		Width: 1920, Height: 1080, Framerate: 30, BitrateKbps: 0,
	})
	if err == nil {
		t.Fatal("expected error for bitrate=0")
	}
}

func TestNVENCClose_Valid(t *testing.T) {
	s := New()
	open, err := s.NVENCOpen(context.Background(), &pb.OpenRequest{
		Width: 640, Height: 480, Framerate: 25, BitrateKbps: 2000,
	})
	if err != nil {
		t.Fatalf("NVENCOpen: %v", err)
	}
	if _, err := s.NVENCClose(context.Background(), &pb.CloseRequest{SessionID: open.SessionID}); err != nil {
		t.Fatalf("NVENCClose: %v", err)
	}
}

func TestNVENCClose_UnknownSession(t *testing.T) {
	s := New()
	_, err := s.NVENCClose(context.Background(), &pb.CloseRequest{SessionID: "no-such-session"})
	if err == nil {
		t.Fatal("expected error for unknown session")
	}
}

func TestNVENCClose_EmptySessionID(t *testing.T) {
	s := New()
	_, err := s.NVENCClose(context.Background(), &pb.CloseRequest{SessionID: ""})
	if err == nil {
		t.Fatal("expected error for empty session_id")
	}
}

func TestNVENCClose_DoubleFree(t *testing.T) {
	s := New()
	open, _ := s.NVENCOpen(context.Background(), &pb.OpenRequest{
		Width: 640, Height: 480, Framerate: 25, BitrateKbps: 2000,
	})
	s.NVENCClose(context.Background(), &pb.CloseRequest{SessionID: open.SessionID}) //nolint:errcheck
	_, err := s.NVENCClose(context.Background(), &pb.CloseRequest{SessionID: open.SessionID})
	if err == nil {
		t.Fatal("expected error on double-close")
	}
}

// ---------------------------------------------------------------------------
// Analyze
// ---------------------------------------------------------------------------

func TestAnalyze_InvalidDimensions(t *testing.T) {
	s := New()
	_, err := s.Analyze(context.Background(), &pb.AnalyzeRequest{
		Width: 0, Height: 480, BGRA: make([]byte, 0),
	})
	if err == nil {
		t.Fatal("expected error for width=0")
	}
}

func TestAnalyze_ValidFrame(t *testing.T) {
	s := New()
	w, h := int32(4), int32(4)
	bgra := make([]byte, w*h*4)
	resp, err := s.Analyze(context.Background(), &pb.AnalyzeRequest{
		BGRA: bgra, Width: w, Height: h,
	})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
}

// ---------------------------------------------------------------------------
// Diff
// ---------------------------------------------------------------------------

func TestDiff_IdenticalFrames(t *testing.T) {
	s := New()
	w, h := int32(4), int32(4)
	frame := make([]byte, w*h*4)
	resp, err := s.Diff(context.Background(), &pb.DiffRequest{
		Before: frame, After: frame, Width: w, Height: h,
	})
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if resp.TotalDelta != 0 {
		t.Errorf("identical frames must have delta=0, got %f", resp.TotalDelta)
	}
	if len(resp.Regions) != 0 {
		t.Errorf("expected no regions for identical frames, got %d", len(resp.Regions))
	}
}

func TestDiff_ChangedFrame(t *testing.T) {
	s := New()
	w, h := int32(4), int32(4)
	before := make([]byte, w*h*4)
	after := make([]byte, w*h*4)
	// Set all pixels in after to white.
	for i := range after {
		after[i] = 255
	}
	resp, err := s.Diff(context.Background(), &pb.DiffRequest{
		Before: before, After: after, Width: w, Height: h,
	})
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if resp.TotalDelta == 0 {
		t.Error("expected non-zero delta for changed frames")
	}
	if len(resp.Regions) == 0 {
		t.Error("expected at least one changed region")
	}
}

func TestDiff_InvalidDimensions(t *testing.T) {
	s := New()
	_, err := s.Diff(context.Background(), &pb.DiffRequest{
		Before: []byte{1}, After: []byte{1}, Width: -1, Height: 1,
	})
	if err == nil {
		t.Fatal("expected error for negative width")
	}
}

func TestDiff_MismatchedBuffers(t *testing.T) {
	s := New()
	_, err := s.Diff(context.Background(), &pb.DiffRequest{
		Before: make([]byte, 16), After: make([]byte, 8), Width: 2, Height: 2,
	})
	if err == nil {
		t.Fatal("expected error for mismatched buffer lengths")
	}
}

// ---------------------------------------------------------------------------
// Match
// ---------------------------------------------------------------------------

func TestMatch_InvalidFrameDims(t *testing.T) {
	s := New()
	_, err := s.Match(context.Background(), &pb.MatchRequest{
		FrameW: 0, FrameH: 480,
	})
	if err == nil {
		t.Fatal("expected error for frame_w=0")
	}
}

func TestMatch_InvalidTemplateDims(t *testing.T) {
	s := New()
	_, err := s.Match(context.Background(), &pb.MatchRequest{
		FrameW: 640, FrameH: 480, TmplW: 0, TmplH: 32,
	})
	if err == nil {
		t.Fatal("expected error for tmpl_w=0")
	}
}

func TestMatch_ValidReturnsEmptyHits(t *testing.T) {
	s := New()
	resp, err := s.Match(context.Background(), &pb.MatchRequest{
		Frame:    make([]byte, 640*480*4),
		Template: make([]byte, 16*16*4),
		FrameW:   640, FrameH: 480,
		TmplW:   16, TmplH: 16,
	})
	if err != nil {
		t.Fatalf("Match: %v", err)
	}
	// Stub always returns empty; CUDA backend may return hits.
	if resp.Hits == nil {
		t.Fatal("expected non-nil hits slice")
	}
}

// ---------------------------------------------------------------------------
// OCR
// ---------------------------------------------------------------------------

func TestOCR_InvalidDimensions(t *testing.T) {
	s := New()
	_, err := s.OCR(context.Background(), &pb.OCRRequest{
		Width: 0, Height: 100,
	})
	if err == nil {
		t.Fatal("expected error for width=0")
	}
}

func TestOCR_ValidReturnsEmpty(t *testing.T) {
	s := New()
	w, h := int32(64), int32(64)
	resp, err := s.OCR(context.Background(), &pb.OCRRequest{
		BGRA: make([]byte, w*h*4), Width: w, Height: h,
	})
	if err != nil {
		t.Fatalf("OCR: %v", err)
	}
	// Stub returns empty text; operator wires Tesseract in the CUDA build.
	if resp.Blocks == nil {
		t.Fatal("expected non-nil blocks slice")
	}
}

// ---------------------------------------------------------------------------
// HTTP shim — round-trip via RegisterRoutes
// ---------------------------------------------------------------------------

func TestHTTPShim_NVENCOpen(t *testing.T) {
	s := New()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	body, _ := json.Marshal(pb.OpenRequest{
		Width: 1280, Height: 720, Framerate: 30, BitrateKbps: 3000,
	})
	req := httptest.NewRequest(http.MethodPost, "/nvenc/open", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp pb.OpenResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.SessionID == "" {
		t.Error("expected non-empty session_id")
	}
}

func TestHTTPShim_WrongMethod(t *testing.T) {
	s := New()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/nvenc/open", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestHTTPShim_BadJSON(t *testing.T) {
	s := New()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/nvenc/open", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHTTPShim_DiffRoundTrip(t *testing.T) {
	s := New()
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	w, h := int32(2), int32(2)
	before := make([]byte, w*h*4)
	after := make([]byte, w*h*4)
	for i := range after {
		after[i] = 128
	}
	body, _ := json.Marshal(pb.DiffRequest{
		Before: before, After: after, Width: w, Height: h,
	})
	req := httptest.NewRequest(http.MethodPost, "/opencv/diff", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp pb.DiffResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalDelta == 0 {
		t.Error("expected non-zero delta")
	}
}
