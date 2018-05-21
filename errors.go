package v4l2

import "errors"

var (
	ErrorWrongDevice  = errors.New("Wrong V4L2 device")
	ErrorNotSpecified = errors.New("Not specify device")
)
