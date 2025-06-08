// Package logger defines a minimal, leveled logging interface used throughout
// the MCP reference implementation.
//
// The interface mirrors the traditional syslog severity levels (Debug, Info,
// Notice, Warning, Error, Critical, Alert, Emergency) and allows callers to
// obtain sub-loggers by name.  Implementers can route the events to their
// logging backend of choice while remaining decoupled from the core protocol
// code.
package logger
