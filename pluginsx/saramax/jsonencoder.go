package saramax

import "encoding/json"

type JSONEncoder struct {
	Data any
}

func (s JSONEncoder) Encode() ([]byte, error) {
	return json.Marshal(s.Data)
}

func (s JSONEncoder) Length() int {
	b, _ := json.Marshal(s.Data)
	return len(b)
}
