package primitive

type ID interface {
	ID() string
}

type Type interface {
	Type() string
}

type TypedID interface {
	ID
	Type
}

type stringFunc func() string

func (s stringFunc) ID() string {
	return s()
}
func (s stringFunc) Type() string {
	return s()
}

func StringID(id string) ID {
	return stringFunc(func() string {
		return id
	})
}

func StringType(typ string) Type {
	return stringFunc(func() string {
		return typ
	})
}
