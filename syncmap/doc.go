// Package syncmap provides a very small, generic wrapper around Go’s built-in
// map type guarded by a sync.RWMutex.
//
// It trades a little indirection for convenience when a concurrent, goroutine
// safe map is needed but the full feature set (and memory cost) of
// `sync.Map` is not justified.
package syncmap
