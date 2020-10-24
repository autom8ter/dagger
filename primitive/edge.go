package primitive

import "encoding/json"

// Edge is a relationship between two nodes
type Edge struct {
	// An edge implements Node because it has an Identifier and attributes
	Node `json:"node"`
	// From returns the root node of the edge
	From Node `json:"from"`
	// To returns the target node of the edge
	To Node `json:"to"`
}

func (e *Edge) JSON() ([]byte, error) {
	return json.Marshal(e)
}
