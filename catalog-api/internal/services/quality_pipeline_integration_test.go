package services

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math/rand"
	"testing"

	"catalogizer/repository"

	"digital.vasic.assets/pkg/resolver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQualityPipeline_RejectsBadUntilGoodArrives asserts end-to-end behavior:
// a blurry "share" candidate is rejected by the gate, the chain falls through
// to a high-res "provider" candidate that passes, and the persisted
// assessment reflects the winning source.
func TestQualityPipeline_RejectsBadUntilGoodArrives(t *testing.T) {
	db := setupTestAssetDBForQG(t)
	repo := repository.NewImageQualityRepository(db)

	shareBytes := uniformJPEG(t, 100, 150) // low-res + blurry
	tmdbBytes := noisyJPEGForQG(t, 700, 1050, 92, 5)

	shareInner := &fakeResolver{
		name: "share", priority: 5, can: true, mime: "image/jpeg", data: shareBytes,
	}
	tmdbInner := &fakeResolver{
		name: "tmdb", priority: 10, can: true, mime: "image/jpeg", data: tmdbBytes,
	}
	chain := resolver.NewChain(
		NewQualityGate(shareInner, WithQualityRepository(repo), WithSourceLabel("share")),
		NewQualityGate(tmdbInner, WithQualityRepository(repo), WithSourceLabel("tmdb")),
	)

	res, err := chain.Resolve(context.Background(), &resolver.ResolveRequest{
		EntityType: "movie", EntityID: "101", Metadata: map[string]string{},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	defer res.Content.Close()

	body, err := io.ReadAll(res.Content)
	require.NoError(t, err)
	assert.Equal(t, tmdbBytes, body, "chain should serve the tmdb result")

	rec, err := repo.Find(context.Background(), "movie", 101, "primary")
	require.NoError(t, err)
	assert.Equal(t, "pass", rec.Verdict)
	assert.Equal(t, "tmdb", rec.Source)
}

// TestQualityPipeline_AllFailFallsThroughToError verifies that when every
// candidate fails the gate, the chain surfaces a terminal error rather than
// returning a low-quality image.
func TestQualityPipeline_AllFailFallsThroughToError(t *testing.T) {
	repo := repository.NewImageQualityRepository(setupTestAssetDBForQG(t))
	a := &fakeResolver{name: "a", priority: 5, can: true, mime: "image/jpeg", data: uniformJPEG(t, 50, 50)}
	b := &fakeResolver{name: "b", priority: 6, can: true, mime: "image/jpeg", data: uniformJPEG(t, 80, 80)}
	chain := resolver.NewChain(
		NewQualityGate(a, WithQualityRepository(repo), WithSourceLabel("a")),
		NewQualityGate(b, WithQualityRepository(repo), WithSourceLabel("b")),
	)

	_, err := chain.Resolve(context.Background(), &resolver.ResolveRequest{
		EntityType: "movie", EntityID: "202",
	})
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrQualityGateFailed),
		"chain returns a terminal 'no resolver' error, not a gate error, once exhausted")
}

// uniformJPEG produces a uniform-color JPEG that trips the blur and often the
// bpp gates at small sizes. Used to stand in for share-sourced low-quality art.
func uniformJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}))
	return buf.Bytes()
}

func noisyJPEGForQG(t *testing.T, w, h, q int, seed int64) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	r := rand.New(rand.NewSource(seed))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(r.Intn(256))
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 0xff})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}))
	return buf.Bytes()
}

// BenchmarkQualityGate measures the per-call overhead of the gate on a
// 1MP (700x1500) JPEG — target < 5ms per call.
func BenchmarkQualityGate(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 700, 1050))
	r := rand.New(rand.NewSource(11))
	for y := 0; y < 1050; y++ {
		for x := 0; x < 700; x++ {
			v := uint8(r.Intn(256))
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 0xff})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 92})
	data := buf.Bytes()

	inner := &fakeResolver{name: "bench", can: true, mime: "image/jpeg", data: data}
	gate := NewQualityGate(inner)
	req := &resolver.ResolveRequest{EntityType: "movie", EntityID: "1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		res, err := gate.Resolve(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, res.Content)
		_ = res.Content.Close()
	}
}
