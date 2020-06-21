package types

type (
	// StringAttribute defines en Event object wrapper where all the attributes
	// contain key/value pairs that are strings instead of raw bytes.
	StringEvent struct {
		Type       string      `json:"type,omitempty"`
		Attributes []Attribute `json:"attributes,omitempty"`
	}

	// StringAttributes defines a slice of StringEvents objects.
	StringEvents []StringEvent
)

type (
	// Event is a type alias for an ABCI Event
	Event abci.Event

	// Attribute defines an attribute wrapper where the key and value are
	// strings instead of raw bytes.
	Attribute struct {
		Key   string `json:"key"`
		Value string `json:"value,omitempty"`
	}

	// Events defines a slice of Event objects
	Events []Event
)
