// Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations and
// renders the defined HTML document.
package quill

import "encoding/json"

func Render(ops []byte) ([]byte, error) {

	var oa = []Op
	err := json.Unmarshal(ops, &oa)
	if err != nil {
		return nil, err
	}

	for i := range oa {}

	return nil

}

type Op map[string]interface{}

type OpHandler func(prev Op, cur Op, )

var handlers = map[string]func(prev, cur Op){
	"string": func(prev, cur Op) {
	},
}
