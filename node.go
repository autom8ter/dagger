package dagger

const (
	ID_KEY   = "_id"
	TYPE_KEY = "_type"
)

// Node is a single element in the Graph. It may be connected to other Nodes via Edges
type Node struct {
	// Attributes returns arbitrary data found in the node(if it exists)
	Values
}

func (n *Node) Type() string {
	val, ok := n.Get(TYPE_KEY)
	if ok {
		return val.(string)
	}
	return ""
}

func (n *Node) ID() string {
	val, ok := n.Get(ID_KEY)
	if ok {
		return val.(string)
	}
	return ""
}

func (n *Node) HasID() bool {
	return n.Exists(ID_KEY)
}

func (n *Node) HasType() bool {
	return n.Exists(TYPE_KEY)
}

func (n *Node) SetID(id string) {
	n.Set(ID_KEY, id)
}

func (n *Node) SetType(nodeType string) {
	n.Set(TYPE_KEY, nodeType)
}
