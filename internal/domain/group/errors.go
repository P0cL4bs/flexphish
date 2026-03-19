package group

import "errors"

var (
	ErrNameAlreadyExists = errors.New("group name already exists")
	ErrInvalidTargets    = errors.New("one or more targets are invalid")
	ErrTargetEmailExists = errors.New("target email already exists in group")
)
