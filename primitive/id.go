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

func RandomID() ID {
	return stringFunc(func() string {
		return uuid()
	})
}

func StringType(typ string) Type {
	return stringFunc(func() string {
		return typ
	})
}

type foreignKey struct {
	typ string
	id  string
}

func (f foreignKey) ID() string {
	return f.id
}

func (f foreignKey) Type() string {
	return f.typ
}

func ForeignKey(typ string, id string) TypedID {
	return &foreignKey{
		typ: typ,
		id:  id,
	}
}
