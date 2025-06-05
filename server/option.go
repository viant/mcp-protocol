package server

// Option is a default implementation of the server interface
type Option func(server *DefaultServer) error
