package models

type StreamArrayResult[T any] struct {
	Data []T
	Err  error
}

type StreamResult[T any] struct {
	Data T
	Err  error
}
