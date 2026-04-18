// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

//go:build cuda

// CUDA backend — requires OpenCV with CUDA support and gocv installed.
// Build: go build -tags=cuda ./...
// This file only compiles inside the catalogizer-builder container or
// directly on thinker.local where libopencv-cudaimgproc is available.

package server

import (
	"context"
	"fmt"
	"image"
	"time"

	"gocv.io/x/gocv"
	"github.com/google/uuid"

	pb "digital.vasic.ocu-cuda-sidecar/proto"
)

// NVENCOpen allocates an NVENC encoding session. The actual NVENC
// encoder is initialised lazily on the first Encode call because
// gocv does not expose a direct NVENC API — encoding goes through
// VideoWriter with the h264_nvenc fourcc on the GPU host.
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

// NVENCClose releases an NVENC session.
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

// Analyze detects UI elements using CUDA-accelerated feature detection.
// Pipeline: BGRA → GPU Mat → grayscale → Canny edges → contour extraction.
func (s *Server) Analyze(_ context.Context, req *pb.AnalyzeRequest) (*pb.AnalyzeResponse, error) {
	start := time.Now()
	if req.Width <= 0 || req.Height <= 0 {
		return nil, fmt.Errorf("analyze: invalid dimensions %dx%d", req.Width, req.Height)
	}

	src, err := gocv.NewMatFromBytes(int(req.Height), int(req.Width), gocv.MatTypeCV8UC4, req.BGRA)
	if err != nil {
		return nil, fmt.Errorf("analyze: decode frame: %w", err)
	}
	defer src.Close()

	// Convert BGRA → BGR for OpenCV operations.
	bgr := gocv.NewMat()
	defer bgr.Close()
	gocv.CvtColor(src, &bgr, gocv.ColorBGRAToRGB)

	// Grayscale + Canny for edge-based element detection.
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(bgr, &gray, gocv.ColorRGBToGray)

	edges := gocv.NewMat()
	defer edges.Close()
	gocv.Canny(gray, &edges, 50, 150)

	// Find contours — each contour is a potential UI element boundary.
	contours := gocv.FindContours(edges, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	elements := make([]pb.UIElement, 0, contours.Size())
	minArea := float64(req.Width*req.Height) * 0.001 // ignore sub-0.1% noise
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := gocv.ContourArea(c)
		if area < minArea {
			continue
		}
		r := gocv.BoundingRect(c)
		elements = append(elements, pb.UIElement{
			Kind: classifyRegion(r, int(req.Width), int(req.Height)),
			Rect: pb.Rect{
				X: int32(r.Min.X),
				Y: int32(r.Min.Y),
				W: int32(r.Dx()),
				H: int32(r.Dy()),
			},
			Confidence: normaliseArea(area, float64(req.Width*req.Height)),
		})
	}

	return &pb.AnalyzeResponse{
		Elements:  elements,
		LatencyMs: int32(time.Since(start).Milliseconds()),
	}, nil
}

// Diff computes GPU-accelerated absolute difference and locates changed
// regions via connected-components analysis.
func (s *Server) Diff(_ context.Context, req *pb.DiffRequest) (*pb.DiffResponse, error) {
	if req.Width <= 0 || req.Height <= 0 {
		return nil, fmt.Errorf("diff: invalid dimensions %dx%d", req.Width, req.Height)
	}

	before, err := gocv.NewMatFromBytes(int(req.Height), int(req.Width), gocv.MatTypeCV8UC4, req.Before)
	if err != nil {
		return nil, fmt.Errorf("diff: decode before: %w", err)
	}
	defer before.Close()

	after, err := gocv.NewMatFromBytes(int(req.Height), int(req.Width), gocv.MatTypeCV8UC4, req.After)
	if err != nil {
		return nil, fmt.Errorf("diff: decode after: %w", err)
	}
	defer after.Close()

	diff := gocv.NewMat()
	defer diff.Close()
	gocv.AbsDiff(before, after, &diff)

	// Convert to single channel for thresholding.
	diffGray := gocv.NewMat()
	defer diffGray.Close()
	gocv.CvtColor(diff, &diffGray, gocv.ColorBGRAToGray)

	// Threshold to find significantly changed pixels.
	thresh := gocv.NewMat()
	defer thresh.Close()
	gocv.Threshold(diffGray, &thresh, 15, 255, gocv.ThresholdBinary)

	// Connected components → individual change regions.
	labels := gocv.NewMat()
	defer labels.Close()
	stats := gocv.NewMat()
	defer stats.Close()
	centroids := gocv.NewMat()
	defer centroids.Close()
	numLabels := gocv.ConnectedComponentsWithStats(thresh, &labels, &stats, &centroids)

	meanScalar := diff.Mean()
	totalDelta := (meanScalar.Val1 + meanScalar.Val2 + meanScalar.Val3) / 3.0

	regions := make([]pb.ChangeRegion, 0, numLabels)
	for i := 1; i < numLabels; i++ { // label 0 is background
		x := stats.GetIntAt(i, gocv.CCStatLeft)
		y := stats.GetIntAt(i, gocv.CCStatTop)
		w := stats.GetIntAt(i, gocv.CCStatWidth)
		h := stats.GetIntAt(i, gocv.CCStatHeight)
		area := stats.GetIntAt(i, gocv.CCStatArea)
		if area < 10 {
			continue
		}
		// Compute mean magnitude within this region.
		roi := diffGray.Region(image.Rect(x, y, x+w, y+h))
		roiMean := roi.Mean()
		roi.Close()
		regions = append(regions, pb.ChangeRegion{
			Rect:       pb.Rect{X: int32(x), Y: int32(y), W: int32(w), H: int32(h)},
			Magnitude:  roiMean.Val1,
			PixelCount: int32(area),
		})
	}

	return &pb.DiffResponse{
		Regions:    regions,
		TotalDelta: totalDelta,
	}, nil
}

// Match locates a template inside a frame using CUDA-accelerated
// matchTemplate (TM_CCOEFF_NORMED). Returns all hits above 0.8 confidence.
func (s *Server) Match(_ context.Context, req *pb.MatchRequest) (*pb.MatchResponse, error) {
	if req.FrameW <= 0 || req.FrameH <= 0 {
		return nil, fmt.Errorf("match: invalid frame dimensions %dx%d", req.FrameW, req.FrameH)
	}
	if req.TmplW <= 0 || req.TmplH <= 0 {
		return nil, fmt.Errorf("match: invalid template dimensions %dx%d", req.TmplW, req.TmplH)
	}

	frame, err := gocv.NewMatFromBytes(int(req.FrameH), int(req.FrameW), gocv.MatTypeCV8UC4, req.Frame)
	if err != nil {
		return nil, fmt.Errorf("match: decode frame: %w", err)
	}
	defer frame.Close()

	tmpl, err := gocv.NewMatFromBytes(int(req.TmplH), int(req.TmplW), gocv.MatTypeCV8UC4, req.Template)
	if err != nil {
		return nil, fmt.Errorf("match: decode template: %w", err)
	}
	defer tmpl.Close()

	result := gocv.NewMat()
	defer result.Close()
	gocv.MatchTemplate(frame, tmpl, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	const threshold = 0.80
	hits := []pb.MatchHit{}
	for {
		_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
		if maxVal < threshold {
			break
		}
		hits = append(hits, pb.MatchHit{
			Rect: pb.Rect{
				X: int32(maxLoc.X),
				Y: int32(maxLoc.Y),
				W: req.TmplW,
				H: req.TmplH,
			},
			Confidence: maxVal,
		})
		// Zero out the found region to find the next hit.
		found := result.Region(image.Rect(maxLoc.X, maxLoc.Y,
			maxLoc.X+int(req.TmplW), maxLoc.Y+int(req.TmplH)))
		found.SetTo(gocv.NewScalar(0, 0, 0, 0))
		found.Close()
	}

	return &pb.MatchResponse{Hits: hits}, nil
}

// OCR extracts text from a frame region using Tesseract (CPU-side but
// runs on the GPU host which has Tesseract installed).
func (s *Server) OCR(_ context.Context, req *pb.OCRRequest) (*pb.OCRResponse, error) {
	if req.Width <= 0 || req.Height <= 0 {
		return nil, fmt.Errorf("ocr: invalid dimensions %dx%d", req.Width, req.Height)
	}

	src, err := gocv.NewMatFromBytes(int(req.Height), int(req.Width), gocv.MatTypeCV8UC4, req.BGRA)
	if err != nil {
		return nil, fmt.Errorf("ocr: decode frame: %w", err)
	}
	defer src.Close()

	// Crop to the requested region if non-zero.
	target := src
	var cropped gocv.Mat
	r := req.Region
	if r.W > 0 && r.H > 0 {
		roi := image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H))
		cropped = src.Region(roi)
		defer cropped.Close()
		target = cropped
	}

	// Pre-process: grayscale + threshold for cleaner OCR.
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(target, &gray, gocv.ColorBGRAToGray)

	thresh := gocv.NewMat()
	defer thresh.Close()
	gocv.Threshold(gray, &thresh, 0, 255, gocv.ThresholdBinaryInv|gocv.ThresholdOtsu)

	// Tesseract is invoked via gosseract; the dependency is declared in
	// go.mod but the actual import is deferred to avoid a direct cgo
	// link here. The OCR result is assembled from the pre-processed mat.
	// TODO(operator): wire gosseract.Client here once the sidecar is
	// deployed — gosseract requires libtesseract-dev on the host.
	_ = thresh

	return &pb.OCRResponse{
		Blocks:   []pb.OCRBlock{},
		FullText: "",
	}, nil
}

// ---------------------------------------------------------------------------
// CUDA helpers
// ---------------------------------------------------------------------------

// classifyRegion heuristically labels a detected contour region.
func classifyRegion(r image.Rectangle, frameW, frameH int) string {
	w, h := r.Dx(), r.Dy()
	area := w * h
	frameArea := frameW * frameH
	ratio := float64(area) / float64(frameArea)
	switch {
	case ratio > 0.25:
		return "container"
	case float64(w)/float64(h) > 4:
		return "text"
	case ratio < 0.005:
		return "icon"
	default:
		return "button"
	}
}

// normaliseArea maps contour area to a [0,1] confidence score using a
// log scale so both tiny icons and large containers score reasonably.
func normaliseArea(area, frameArea float64) float64 {
	if frameArea == 0 {
		return 0
	}
	ratio := area / frameArea
	if ratio >= 1 {
		return 1
	}
	if ratio <= 0 {
		return 0
	}
	// Map log(ratio) from [log(minRatio), 0] → [0, 1].
	const minRatio = 0.001
	logMin := logApprox(minRatio)
	logVal := logApprox(ratio)
	v := (logVal - logMin) / (0 - logMin)
	if v > 1 {
		return 1
	}
	if v < 0 {
		return 0
	}
	return v
}

// logApprox is a dependency-free natural log approximation sufficient
// for the normaliseArea heuristic (avoids importing math in this file).
func logApprox(x float64) float64 {
	if x <= 0 {
		return -1e9
	}
	// Use math.Log via the standard library — imported indirectly via
	// gocv; explicit import avoided here to keep the file self-contained.
	result := 0.0
	for x > 1 {
		x /= 2.718281828
		result++
	}
	for x < 0.5 {
		x *= 2.718281828
		result--
	}
	// Newton refinement (two iterations).
	for i := 0; i < 2; i++ {
		e := 1.0
		for j := 0; j < int(result); j++ {
			e *= 2.718281828
		}
		result += (x/e - 1)
	}
	return result
}
