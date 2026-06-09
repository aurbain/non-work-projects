TOOL_NAME: edit_existing_file
BEGIN_ARG: filepath
"tetris/src/input/input.go"
END_ARG
BEGIN_ARG: changes
"package input

import (
	\"bufio\"
	\"os\"
)

type InputEvent int

const (
	MoveLeft InputEvent = iota
	MoveRight
	Rotate
	SoftDrop
	HardDrop
