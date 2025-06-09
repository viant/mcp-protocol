package server

// Option can be supplied to WithDefaultHandler to mutate the handler before use.
type Option func(server *DefaultHandler) error
