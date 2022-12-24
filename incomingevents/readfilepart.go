package incomingevents

type ReadFilePart struct {
	Event

	Data []byte `json:"data"`
}
