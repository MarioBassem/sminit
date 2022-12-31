package stack

import (
	"errors"
)

type primitive interface {
	floatingPoing | integer | string
}

type floatingPoing interface {
	float32 | float64
}

type integer interface {
	int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64
}

type Stack[P primitive] struct {
	container []P
}

func (s *Stack[P]) Push(item P) {
	s.container = append(s.container, item)
}

func (s *Stack[P]) Top() (P, error) {
	var ret P
	if len(s.container) == 0 {
		return ret, errors.New("stack is empty")
	}
	ret = s.container[len(s.container)-1]
	s.container = s.container[:len(s.container)-1]
	return ret, nil
}

func (s *Stack[P]) Empty() bool {
	return len(s.container) == 0
}

func (s *Stack[P]) Size() int {
	return len(s.container)
}
