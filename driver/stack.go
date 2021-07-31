package driver

type stackI []interface{}

// IsEmpty: check if stack is empty
func (s *stackI) IsEmpty() bool {
	return s == nil || len(*s) == 0
}

// Push a new value onto the stack
func (s *stackI) Push(f interface{}) {
	*s = append(*s, f) // Simply append the new value to the end of the stack
}

// Remove and return top element of stack. Return false if stack is empty.
func (s *stackI) Pop() (interface{}, bool) {
	if s.IsEmpty() {
		return nil, false
	} else {
		index := len(*s) - 1   // Get the index of the top most element.
		element := (*s)[index] // Index into the slice and obtain the element.
		*s = (*s)[:index]      // Remove it from the stack by slicing it off.
		return element, true
	}
}

func (s *stackI) Range(fn func(element interface{}) bool) {
	for {
		r, ok := s.Pop()
		if !ok {
			return
		}
		if !fn(r) {
			return
		}
	}
}

func (s *stackI) Len() int {
	return len(*s)
}
