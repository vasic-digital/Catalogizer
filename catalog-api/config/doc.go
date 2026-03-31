// Package config provides configuration loading and management for the Catalogizer API.
//
// It supports a layered precedence model: environment variables override .env file values,
// which override config.json settings, which override built-in defaults. This package
// handles parsing, validation, and runtime access to all application configuration including
// server ports, database settings, JWT secrets, and external API keys.
package config
