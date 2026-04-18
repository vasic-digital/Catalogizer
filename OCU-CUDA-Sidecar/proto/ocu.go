// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

// Package proto contains the hand-authored Go equivalents of the
// types defined in ocu.proto. These are plain structs — no generated
// protobuf reflection is used. A full protoc-generated version (with
// proto.Message, binary marshalling, etc.) can replace this file once
// the operator runs:
//
//	protoc --go_out=. --go_opt=paths=source_relative \
//	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
//	       proto/ocu.proto
//
// Until then the gRPC server uses JSON encoding over gRPC (content-
// type application/grpc+json) which requires no generated code.
package proto

// ---------------------------------------------------------------------------
// Shared geometry
// ---------------------------------------------------------------------------

// Rect is an axis-aligned bounding box in pixel coordinates.
type Rect struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
	W int32 `json:"w"`
	H int32 `json:"h"`
}

// ---------------------------------------------------------------------------
// NVENCService messages
// ---------------------------------------------------------------------------

// OpenRequest allocates an NVENC session.
type OpenRequest struct {
	Width       int32 `json:"width"`
	Height      int32 `json:"height"`
	Framerate   int32 `json:"framerate"`
	BitrateKbps int32 `json:"bitrate_kbps"`
}

// OpenResponse carries the opaque session token.
type OpenResponse struct {
	SessionID string `json:"session_id"`
}

// EncodeChunk is one raw BGRA frame sent by the client.
type EncodeChunk struct {
	SessionID string `json:"session_id"`
	// BGRA holds width * height * 4 raw bytes.
	BGRA []byte `json:"bgra"`
}

// EncodedChunk carries one or more complete H.264 NAL units.
type EncodedChunk struct {
	H264 []byte `json:"h264"`
}

// CloseRequest releases an NVENC session.
type CloseRequest struct {
	SessionID string `json:"session_id"`
}

// CloseResponse is empty; presence signals success.
type CloseResponse struct{}

// ---------------------------------------------------------------------------
// OpenCVService — Analyze
// ---------------------------------------------------------------------------

// AnalyzeRequest carries a raw BGRA frame for UI element detection.
type AnalyzeRequest struct {
	BGRA   []byte `json:"bgra"`
	Width  int32  `json:"width"`
	Height int32  `json:"height"`
}

// UIElement is one detected UI element in a frame.
type UIElement struct {
	Kind       string  `json:"kind"`
	Rect       Rect    `json:"rect"`
	Confidence float64 `json:"confidence"`
}

// AnalyzeResponse contains all detected elements and a server-side
// latency measurement.
type AnalyzeResponse struct {
	Elements  []UIElement `json:"elements"`
	LatencyMs int32       `json:"latency_ms"`
}

// ---------------------------------------------------------------------------
// OpenCVService — Diff
// ---------------------------------------------------------------------------

// DiffRequest carries two raw BGRA frames (same dimensions).
type DiffRequest struct {
	Before []byte `json:"before"`
	After  []byte `json:"after"`
	Width  int32  `json:"width"`
	Height int32  `json:"height"`
}

// ChangeRegion is one bounding box where the frames differ significantly.
type ChangeRegion struct {
	Rect       Rect    `json:"rect"`
	Magnitude  float64 `json:"magnitude"`
	PixelCount int32   `json:"pixel_count"`
}

// DiffResponse lists changed regions and the scene-level mean delta.
type DiffResponse struct {
	Regions    []ChangeRegion `json:"regions"`
	TotalDelta float64        `json:"total_delta"`
}

// ---------------------------------------------------------------------------
// OpenCVService — Match
// ---------------------------------------------------------------------------

// MatchRequest carries a frame and a template to locate inside it.
type MatchRequest struct {
	Frame    []byte `json:"frame"`
	Template []byte `json:"template"`
	FrameW   int32  `json:"frame_w"`
	FrameH   int32  `json:"frame_h"`
	TmplW    int32  `json:"tmpl_w"`
	TmplH    int32  `json:"tmpl_h"`
}

// MatchHit is one location where the template was found above threshold.
type MatchHit struct {
	Rect       Rect    `json:"rect"`
	Confidence float64 `json:"confidence"`
}

// MatchResponse lists all template hits above the confidence threshold.
type MatchResponse struct {
	Hits []MatchHit `json:"hits"`
}

// ---------------------------------------------------------------------------
// OpenCVService — OCR
// ---------------------------------------------------------------------------

// OCRRequest asks for text extraction in an optional sub-region.
type OCRRequest struct {
	BGRA   []byte `json:"bgra"`
	Width  int32  `json:"width"`
	Height int32  `json:"height"`
	// Region restricts OCR to a sub-rect; zero value means full frame.
	Region Rect `json:"region"`
}

// OCRBlock is one text block detected by Tesseract.
type OCRBlock struct {
	Text       string  `json:"text"`
	Rect       Rect    `json:"rect"`
	Confidence float64 `json:"confidence"`
}

// OCRResponse contains all detected text blocks and the concatenated
// full-frame text.
type OCRResponse struct {
	Blocks   []OCRBlock `json:"blocks"`
	FullText string     `json:"full_text"`
}
