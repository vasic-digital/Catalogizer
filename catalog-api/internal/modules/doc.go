// Package modules provides a registry for wiring external Go submodules into the Catalogizer
// API at startup.
//
// It manages the initialization and configuration of the 22+ independent Go submodules
// (such as Auth, Cache, Config, Database, Filesystem, and others) that are linked via
// replace directives in go.mod. The module registry ensures each submodule is properly
// initialized with its dependencies and integrated into the application lifecycle.
package modules
