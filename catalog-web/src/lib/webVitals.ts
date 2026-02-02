import type { MetricType } from 'web-vitals'

export type WebVitalsCallback = (metric: MetricType) => void

/**
 * Reports all Web Vitals metrics using the provided callback.
 * Measures CLS, FCP, INP, LCP, and TTFB.
 */
export function reportWebVitals(onReport: WebVitalsCallback): void {
  import('web-vitals').then(({ onCLS, onFCP, onINP, onLCP, onTTFB }) => {
    onCLS(onReport)
    onFCP(onReport)
    onINP(onReport)
    onLCP(onReport)
    onTTFB(onReport)
  })
}

/**
 * Logs Web Vitals metrics to the console in development mode.
 */
export function logWebVitals(): void {
  if (import.meta.env.DEV) {
    reportWebVitals((metric) => {
      console.log(`[Web Vitals] ${metric.name}:`, {
        value: metric.value,
        rating: metric.rating,
        delta: metric.delta,
        id: metric.id,
        navigationType: metric.navigationType,
      })
    })
  }
}

/**
 * Sends Web Vitals metrics to an analytics endpoint.
 * The endpoint URL is configured via the VITE_ANALYTICS_URL environment variable.
 */
export function sendToAnalytics(): void {
  const analyticsUrl = import.meta.env.VITE_ANALYTICS_URL

  if (!analyticsUrl) {
    return
  }

  reportWebVitals((metric) => {
    const body = JSON.stringify({
      name: metric.name,
      value: metric.value,
      rating: metric.rating,
      delta: metric.delta,
      id: metric.id,
      navigationType: metric.navigationType,
    })

    if (typeof navigator.sendBeacon === 'function') {
      navigator.sendBeacon(analyticsUrl, body)
    } else {
      fetch(analyticsUrl, {
        method: 'POST',
        body,
        headers: { 'Content-Type': 'application/json' },
        keepalive: true,
      })
    }
  })
}
