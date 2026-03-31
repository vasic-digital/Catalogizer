// Package lifecycle provides the LazyServiceRegistry for deferred service initialization
// with dependency ordering in the Catalogizer API.
//
// It manages the startup and shutdown lifecycle of application services, ensuring that
// dependencies are initialized in the correct order and torn down in reverse order during
// graceful shutdown. The lazy initialization pattern defers expensive resource allocation
// until services are first accessed, improving startup time.
package lifecycle
