// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

//go:build !cuda

// Stub backend — compiles without OpenCV or CUDA headers. Used for unit
// tests and for development machines that do not have a GPU.

package server

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	pb "digital.vasic.ocu-cuda-sidecar/proto"
)

// NVENCOpen records a new encoding session and returns its ID.
func (s *Server) NVENCOpen(_ context.Context, req *pb.OpenRequest) (*pb.OpenResponse, error) {
	if req.Width <= 0 || req.Height <= 0 {
		return nil, fmt.Errorf("nvenc: invalid dimensions %dx%d", req.Width, req.Height)
	}
	if req.Framerate <= 0 {
		return nil, fmt.Errorf("nvenc: framerate must be > 0, got %d", req.Framerate)
	}
	if req.BitrateKbps <= 0 {
		return nil, fmt.Errorf("nvenc: bitrate must be > 0, got %d", req.BitrateKbps)
	}
	id := uuid.New().String()
	s.mu.Lock()
	s.sessions[id] = &nvencSession{
		width:       req.Width,
		height:      req.Height,
		framerate:   req.Framerate,
		bitrateKbps: req.BitrateKbps,
	}
	s.mu.Unlock()
	return &pb.OpenResponse{SessionID: id}, nil
}

// NVENCClose releases an encoding session.
func (s *Server) NVENCClose(_ context.Context, req *pb.CloseRequest) (*pb.CloseResponse, error) {
	if req.SessionID == "" {
		return nil, fmt.Errorf("nvenc: session_id is required")
	}
	s.mu.Lock()
	_, ok := s.sessions[req.SessionID]
	if ok {
		delete(s.sessions, req.SessionID)
	}
	s.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("nvenc: unknown session %q", req.SessionID)
	}
	return &pb.CloseResponse{}, nil
}

// Analyze returns an empty element list — stub has no vision model.
func (s *Server) Analyze(_ context.Context, req *pb.AnalyzeRequest) (*pb.AnalyzeResponse, error) {
	start := time.Now()
	if req.Width <= 0 || req.Height <= 0 {
		return nil, fmt.Errorf("analyze: invalid dimensions %dx%d", req.Width, req.Height)
	}
	return &pb.AnalyzeResponse{
		Elements:  []pb.UIElement{},
		LatencyMs: int32(time.Since(start).Milliseconds()),
	}, nil
}

// Diff computes a trivial byte-level difference without GPU acceleration.
// Returns a single region covering the whole frame with the mean delta.
func (s *Server) Diff(_ context.Context, req *pb.DiffRequest) (*pb.DiffResponse, error) {
	if req.Width <= 0 || req.Height <= 0 {
		return nil, fmt.Errorf("diff: invalid dimensions %dx%d", req.Width, req.Height)
	}
	n := len(req.Before)
	if n == 0 || len(req.After) != n {
		return nil, fmt.Errorf("diff: before/after length mismatch (%d vs %d)", len(req.Before), len(req.After))
	}
	var sum uint64
	for i := 0; i < n; i++ {
		d := int(req.Before[i]) - int(req.After[i])
		if d < 0 {
			d = -d
		}
		sum += uint64(d)
	}
	mean := float64(sum) / float64(n)
	regions := []pb.ChangeRegion{}
	if mean > 0 {
		regions = append(regions, pb.ChangeRegion{
			Rect:       pb.Rect{X: 0, Y: 0, W: req.Width, H: req.Height},
			Magnitude:  mean,
			PixelCount: req.Width * req.Height,
		})
	}
	return &pb.DiffResponse{
		Regions:    regions,
		TotalDelta: mean,
	}, nil
}

// Match returns no hits — stub has no template matching implementation.
func (s *Server) Match(_ context.Context, req *pb.MatchRequest) (*pb.MatchResponse, error) {
	if req.FrameW <= 0 || req.FrameH <= 0 {
		return nil, fmt.Errorf("match: invalid frame dimensions %dx%d", req.FrameW, req.FrameH)
	}
	if req.TmplW <= 0 || req.TmplH <= 0 {
		return nil, fmt.Errorf("match: invalid template dimensions %dx%d", req.TmplW, req.TmplH)
	}
	return &pb.MatchResponse{Hits: []pb.MatchHit{}}, nil
}

// OCR returns empty text — stub has no OCR engine.
func (s *Server) OCR(_ context.Context, req *pb.OCRRequest) (*pb.OCRResponse, error) {
	if req.Width <= 0 || req.Height <= 0 {
		return nil, fmt.Errorf("ocr: invalid dimensions %dx%d", req.Width, req.Height)
	}
	return &pb.OCRResponse{
		Blocks:   []pb.OCRBlock{},
		FullText: "",
	}, nil
}
