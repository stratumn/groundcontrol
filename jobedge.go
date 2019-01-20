package groundcontrol

type JobEdge struct {
	Cursor string `json:"cursor"`
	Node   *Job   `json:"node"`
}
