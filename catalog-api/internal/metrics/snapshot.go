package metrics

import (
	"runtime"
	"sort"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// bucketAcc is a single histogram bucket: cumulative sample count at or
// below the upper edge. Used to build the sorted bucket slice passed to
// percentileFromBuckets.
type bucketAcc struct {
	upper float64
	count uint64
}

// HTTPSnapshot is a point-in-time view of the HTTP metrics that the
// reporting service uses to populate ResponseTimes / PerformanceMetrics /
// ErrorRates without having to track its own sliding windows.
type HTTPSnapshot struct {
	// TotalRequests is the sum of HTTPRequestsTotal across all labels.
	TotalRequests uint64
	// TotalDurationSeconds is the sum of all observed durations.
	TotalDurationSeconds float64

	// Percentile durations in seconds, approximated from the histogram
	// buckets via linear interpolation. 0 if no data is recorded yet.
	P50Seconds float64
	P95Seconds float64
	P99Seconds float64

	// MinSeconds is the lower edge of the first non-empty bucket.
	// MaxSeconds is the upper edge of the last non-empty bucket (or +Inf).
	MinSeconds float64
	MaxSeconds float64

	// Errors4xx / Errors5xx are counts of requests with 4xx / 5xx status.
	Errors4xx uint64
	Errors5xx uint64

	// InFlight is the current value of HTTPActiveConnections.
	InFlight float64
}

// SnapshotHTTP gathers the current HTTP metrics and returns a HTTPSnapshot.
// Safe to call concurrently with metric updates.
func SnapshotHTTP() HTTPSnapshot {
	var snap HTTPSnapshot

	// HTTPRequestDuration is a histogram — gather all children and merge
	// their buckets into a single virtual histogram.
	durationChan := make(chan prometheus.Metric, 1024)
	go func() {
		HTTPRequestDuration.Collect(durationChan)
		close(durationChan)
	}()

	bucketMap := map[float64]uint64{}
	var totalCount uint64
	var totalSum float64

	for m := range durationChan {
		var dm dto.Metric
		if err := m.Write(&dm); err != nil {
			continue
		}
		h := dm.GetHistogram()
		if h == nil {
			continue
		}
		totalCount += h.GetSampleCount()
		totalSum += h.GetSampleSum()
		for _, b := range h.GetBucket() {
			bucketMap[b.GetUpperBound()] += b.GetCumulativeCount()
		}
	}

	snap.TotalRequests = totalCount
	snap.TotalDurationSeconds = totalSum

	if totalCount > 0 {
		buckets := make([]bucketAcc, 0, len(bucketMap))
		for ub, c := range bucketMap {
			buckets = append(buckets, bucketAcc{upper: ub, count: c})
		}
		sort.Slice(buckets, func(i, j int) bool {
			return buckets[i].upper < buckets[j].upper
		})

		// Cumulative counts across children may exceed totalCount because
		// each child's buckets are already cumulative and we've summed them;
		// normalize by taking the max bucket count as the effective total.
		effectiveTotal := totalCount
		if len(buckets) > 0 && buckets[len(buckets)-1].count > effectiveTotal {
			effectiveTotal = buckets[len(buckets)-1].count
		}

		snap.P50Seconds = percentileFromBuckets(buckets, 0.50, effectiveTotal)
		snap.P95Seconds = percentileFromBuckets(buckets, 0.95, effectiveTotal)
		snap.P99Seconds = percentileFromBuckets(buckets, 0.99, effectiveTotal)

		// Min = first bucket with any count.
		for i, b := range buckets {
			prev := uint64(0)
			if i > 0 {
				prev = buckets[i-1].count
			}
			if b.count > prev {
				snap.MinSeconds = b.upper
				break
			}
		}
		// Max = last bucket's upper bound (may be +Inf).
		if len(buckets) > 0 {
			snap.MaxSeconds = buckets[len(buckets)-1].upper
		}
	}

	// HTTPRequestsTotal is a counter with a "status" label — gather all
	// children and bucket them into 4xx / 5xx.
	countChan := make(chan prometheus.Metric, 1024)
	go func() {
		HTTPRequestsTotal.Collect(countChan)
		close(countChan)
	}()
	for m := range countChan {
		var dm dto.Metric
		if err := m.Write(&dm); err != nil {
			continue
		}
		var statusLabel string
		for _, l := range dm.GetLabel() {
			if l.GetName() == "status" {
				statusLabel = l.GetValue()
				break
			}
		}
		if statusLabel == "" {
			continue
		}
		status, err := strconv.Atoi(statusLabel)
		if err != nil {
			continue
		}
		c := dm.GetCounter()
		if c == nil {
			continue
		}
		n := uint64(c.GetValue())
		switch {
		case status >= 400 && status < 500:
			snap.Errors4xx += n
		case status >= 500 && status < 600:
			snap.Errors5xx += n
		}
	}

	// HTTPActiveConnections is a gauge — read current value.
	gaugeChan := make(chan prometheus.Metric, 1)
	go func() {
		HTTPActiveConnections.Collect(gaugeChan)
		close(gaugeChan)
	}()
	for m := range gaugeChan {
		var dm dto.Metric
		if err := m.Write(&dm); err != nil {
			continue
		}
		if g := dm.GetGauge(); g != nil {
			snap.InFlight = g.GetValue()
		}
	}

	return snap
}

// percentileFromBuckets approximates a percentile from cumulative histogram
// buckets via linear interpolation. Returns the upper bound of the bucket
// containing the target count if interpolation is not possible (e.g.,
// +Inf bucket).
func percentileFromBuckets(buckets []bucketAcc, q float64, total uint64) float64 {
	if total == 0 || len(buckets) == 0 {
		return 0
	}
	target := float64(total) * q
	var prevUpper float64
	var prevCount uint64
	for _, b := range buckets {
		if float64(b.count) >= target {
			if b.count == prevCount {
				return b.upper
			}
			// Linear interpolation within the bucket.
			frac := (target - float64(prevCount)) / float64(b.count-prevCount)
			return prevUpper + frac*(b.upper-prevUpper)
		}
		prevUpper = b.upper
		prevCount = b.count
	}
	return buckets[len(buckets)-1].upper
}

// SystemSnapshot is a point-in-time view of Go runtime stats that the
// reporting service uses to populate SystemLoad without pulling in
// gopsutil or /proc parsing.
type SystemSnapshot struct {
	Goroutines   int
	AllocBytes   uint64
	SysBytes     uint64
	HeapInuse    uint64
	NumGC        uint32
	PauseTotalNs uint64
}

// SnapshotSystem reads current Go runtime stats.
func SnapshotSystem() SystemSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return SystemSnapshot{
		Goroutines:   runtime.NumGoroutine(),
		AllocBytes:   m.Alloc,
		SysBytes:     m.Sys,
		HeapInuse:    m.HeapInuse,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
	}
}
