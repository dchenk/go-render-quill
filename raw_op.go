package quill

import (
	"fmt"
	"html"
	"strconv"
)

type rawOp struct {
	// Insert is the string containing the data.
	Insert interface{} `json:"insert"`

	// Attrs contains the "attributes" property of the op.
	Attrs map[string]interface{} `json:"attributes"`
}

// makeOp takes a raw Delta op as extracted from the JSON and turns it into an Op to make it usable for rendering.
func (ro *rawOp) makeOp(o *Op) error {

	if ro.Insert == nil {
		return fmt.Errorf("op %+v lacks an insert", *ro)
	}

	switch ins := ro.Insert.(type) {
	case string:
		// This op is a simple string insert.
		o.Type = "text"
		o.Data = html.EscapeString(ins)
	case map[string]interface{}:
		if len(ins) == 0 {
			return fmt.Errorf("op %+v lacks a non-text insert", *ro)
		}
		// There should be one item in the map (the element's key being the insert type).
		for mk := range ins {
			o.Type = mk
			o.Data = extractString(ins[mk])
			break
		}
	default:
		return fmt.Errorf("op %+v lacks an insert", *ro)
	}

	// Clear the map for reuse.
	for k := range o.Attrs {
		delete(o.Attrs, k)
	}

	if ro.Attrs != nil {
		// The map was already made
		for attr := range ro.Attrs {
			o.Attrs[attr] = extractString(ro.Attrs[attr])
		}
	}

	return nil

}

func extractString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val == true {
			return "y"
		}
	case float64:
		return strconv.FormatFloat(val, 'f', 0, 64)
	}
	return ""
}
