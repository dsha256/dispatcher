package types

type Response[T any] struct {
	Data T      `json:"data,omitempty"`
	Err  string `json:"err,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func NewSuccessResponse[T any](msg string, data T) Response[T] {
	return Response[T]{
		Msg:  msg,
		Data: data,
	}
}

func NewErrorResponse[T any](err string) Response[T] {
	return Response[T]{
		Err: err,
	}
}
