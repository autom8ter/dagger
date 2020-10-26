package primitive

type Export struct {
	Nodes []Node `json:"nodes"`
	Edges []*Edge `json:"edges"`
}
