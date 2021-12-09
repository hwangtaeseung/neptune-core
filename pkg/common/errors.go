package common

import "fmt"

type NeptuneError struct {
	code    int
	message string
	action  string
	err     error
}

func (h *NeptuneError) Error() string {
	return fmt.Sprintf("message:%v,  code:%v,  action:%v, err:%v", h.message, h.code, h.action, h.err)
}

func (h *NeptuneError) Code() int {
	return h.code
}

func (h *NeptuneError) Message() string {
	return h.message
}

func (h *NeptuneError) Action() string {
	return h.action
}

func (h *NeptuneError) SysErr() error {
	return h.err
}

func (h *NeptuneError) SetAction(action string) *NeptuneError {
	h.action = action
	return h
}

func (h *NeptuneError) SetMessage(message string) *NeptuneError {
	h.message = message
	return h
}

func (h *NeptuneError) Copy(err error) *NeptuneError {
	return &NeptuneError{
		code:    h.code,
		message: h.message,
		action:  h.action,
		err:     err,
	}
}

func New(code int, message string, action string) *NeptuneError {
	return &NeptuneError{
		code:    code,
		message: message,
		action:  action,
	}
}

const (
	ErrSigTermCode = 9
)

var (
	ErrSigTerm = New(ErrSigTermCode, "signal term", "retry")
)
