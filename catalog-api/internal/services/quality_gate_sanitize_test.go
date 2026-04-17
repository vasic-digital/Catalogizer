package services

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"math/rand"
	"strings"
	"testing"

	"catalogizer/repository"

	"digital.vasic.assets/pkg/resolver"
)

// TestQualityGate_W3_SanitizesProviderStringsInPersist locks in the W3
// fix. A malicious SourceHint must not land in the image_quality_
// assessments table with script tags intact — the gate runs every
// provider-originated string through bluemonday.StrictPolicy before
// persistence.
func TestQualityGate_W3_SanitizesProviderStringsInPersist(t *testing.T) {
	db := setupTestAssetDBForQG(t)
	repo := repository.NewImageQualityRepository(db)

	innerBytes := noisyJPEGForSanitize(t, 700, 1050, 92, 99)
	inner := &fakeResolver{
		name: "evil_provider", priority: 10, can: true, mime: "image/jpeg",
		data: innerBytes,
	}
	gate := NewQualityGate(inner, WithQualityRepository(repo), WithSourceLabel(`tmdb<script>alert(1)</script>`))
	req := &resolver.ResolveRequest{
		EntityType: `movie<img src=x onerror=alert(1)>`,
		EntityID:   "77",
		SourceHint: `https://evil.test/x.jpg"><script>alert(1)</script>`,
		Metadata:   map[string]string{"variant": `primary<b>bold</b>`},
	}

	res, err := gate.Resolve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.Copy(io.Discard, res.Content)
	_ = res.Content.Close()

	stored, err := repo.Find(context.Background(), `movie`, 77, "primary")
	if err != nil {
		// If the literal raw entity_type was persisted, the lookup
		// misses — which is itself acceptable because it means the
		// sanitizer successfully rewrote it. Try the sanitized
		// lookup.
		stored, err = repo.Find(context.Background(), SanitizeMetadataString(`movie<img src=x onerror=alert(1)>`), 77, SanitizeMetadataString(`primary<b>bold</b>`))
		if err != nil {
			t.Fatalf("assessment not found after sanitize: %v", err)
		}
	}
	if strings.Contains(stored.ProviderRef, "<script") || strings.Contains(stored.ProviderRef, "alert(1)") {
		t.Errorf("ProviderRef leaked HTML: %q", stored.ProviderRef)
	}
	if strings.Contains(stored.Source, "<script") || strings.Contains(stored.Source, "alert(1)") {
		t.Errorf("Source leaked HTML: %q", stored.Source)
	}
	if strings.Contains(stored.EntityType, "<") || strings.Contains(stored.EntityType, "onerror") {
		t.Errorf("EntityType leaked HTML: %q", stored.EntityType)
	}
	if strings.Contains(stored.Variant, "<") {
		t.Errorf("Variant leaked HTML: %q", stored.Variant)
	}
}

// noisyJPEGForSanitize is a tiny helper local to the W3 test — the
// other sanitize tests live in sanitize_test.go and do not need
// image fixtures.
func noisyJPEGForSanitize(t *testing.T, w, h, q int, seed int64) []byte {
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
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return buf.Bytes()
}
