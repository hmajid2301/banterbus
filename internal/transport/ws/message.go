package ws

import "encoding/json"

type message struct {
	ExtraFields map[string]interface{} `json:"-"`
	EventName   string                 `json:"event_name"`
}

func (m *message) UnmarshalJSON(data []byte) error {
	type Alias message
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// INFO: Unmarshal extra fields into ExtraFields map, let event handler functions deal with them.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.ExtraFields = make(map[string]interface{})
	for k, v := range raw {
		if k != "event_name" {
			m.ExtraFields[k] = v
		}
	}

	return nil
}
