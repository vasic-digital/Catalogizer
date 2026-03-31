// Package concurrency provides semaphore-based concurrency control for limiting parallel
// operations in the Catalogizer API.
//
// It implements a counting semaphore that restricts the number of concurrent goroutines
// executing resource-intensive operations such as file scanning, media analysis, and network
// requests. This prevents resource exhaustion on the host machine, which is critical given
// the project's 30-40% maximum host resource utilization constraint.
package concurrency
