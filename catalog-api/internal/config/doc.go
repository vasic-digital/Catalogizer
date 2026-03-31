// Package config provides the internal configuration subsystem for the Catalogizer API.
//
// It handles loading and parsing of internal configuration parameters that are not exposed
// through the top-level config package. This includes infrastructure-specific settings for
// internal services such as cache tuning, connection pool parameters, and internal timeouts
// that complement the public configuration API.
package config
