package services

import "os"

// getenv is a thin wrapper that tests cannot shadow but the resolver
// package keeps local so overriding os.Getenv via build flags remains
// feasible if we ever need to plug in a secrets manager.
func getenv(k string) string { return os.Getenv(k) }
