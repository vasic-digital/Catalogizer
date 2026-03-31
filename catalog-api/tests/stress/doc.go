// Package stress contains stress and load tests for the Catalogizer API.
//
// These tests push the system beyond normal operating conditions to identify breaking points,
// resource leaks, and degradation behavior under extreme concurrency. They complement the k6
// load tests by providing Go-native stress scenarios that can be run as part of the standard
// test suite with resource limits enforced.
package stress
