package dagger

import (
	"fmt"
	"github.com/autom8ter/dagger/primitive"
)

// ForeignKey satisfies primitive.TypedID interface
type ForeignKey struct {
	XID   string
	XType string
}

func (f *ForeignKey) ID() string {
	return f.XID
}

func (f *ForeignKey) Type() string {
	return f.XType
}

func (f *ForeignKey) Path() string {
	return fmt.Sprintf("%s.%s", f.Type(), f.ID())
}

type stringFunc func() string

func (s stringFunc) ID() string {
	return s()
}
func (s stringFunc) Type() string {
	return s()
}

func StringID(id string) primitive.ID {
	return stringFunc(func() string {
		return id
	})
}

func RandomID() primitive.ID {
	return stringFunc(func() string {
		return primitive.UUID()
	})
}

func AnyType() primitive.Type {
	return stringFunc(func() string {
		return primitive.AnyType
	})
}

func DefaultType() primitive.Type {
	return stringFunc(func() string {
		return primitive.DefaultType
	})
}

func StringType(typ string) primitive.Type {
	return stringFunc(func() string {
		return typ
	})
}
