package dagger

// Edge is a relationship between two nodes
type Edge struct {
	// An edge implements Node because it has an Identifier and attributes
	*Node
	// From returns the root node of the edge
	From *Node
	// To returns the target node of the edge
	To *Node
}
