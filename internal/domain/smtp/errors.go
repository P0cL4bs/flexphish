package smtp

import "errors"

var (
	ErrConnectionAlreadyExists = errors.New("smtp connection already exists")
)
