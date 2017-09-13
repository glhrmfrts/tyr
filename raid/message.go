package raid

type Message struct {
	Header map[string]interface{} `json:"header"`
	Body   interface{}            `json:"body"`
}

func (m *Message) Action() string {
	return m.Header["action"].(string)
}

func (m *Message) Etag() Etag {
	if v, ok := m.Header["etag"]; ok {
		switch etag := v.(type) {
		case Etag:
			return etag
		case string:
			return Etag(etag)
		}
	}

	return Etag("")
}
