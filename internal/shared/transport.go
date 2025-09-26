package shared

import "encoding/json"

func (c Command) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

func CommandFromJSON(data []byte) (Command, error) {
	var cmd Command
	err := json.Unmarshal(data, &cmd)
	return cmd, err
}

func (r Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func ResponseFromJSON(data []byte) (Response, error) {
	var resp Response
	err := json.Unmarshal(data, &resp)
	return resp, err
}
