package primitive

const DefaultType = "default"

const AnyType = "*"

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
