package game

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrGameNotRunning = errors.New("game is not running")
	ErrForbidden      = errors.New("forbidden")
	ErrNoTestcases    = errors.New("no testcases")
)
