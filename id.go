package dagger

import "github.com/autom8ter/dagger/primitive"

// ForeignKey is a helper that returns a primitive.TypedID from the given type and id
func ForeignKey(typ, id string) primitive.TypedID {
	return primitive.ForeignKey(typ, id)
}

// StringID returns a primitive.ID implemenation that is just the input string
func StringID(id string) primitive.ID {
	return primitive.StringID(id)
}

// StringType returns a primitive.Type implemenation that is just the input string
func StringType(typ string) primitive.Type {
	return primitive.StringType(typ)
}
