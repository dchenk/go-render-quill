package quill_test

import (
	"fmt"

	"github.com/dchenk/go-render-quill"
)

var ops = []byte(`
[
	{
		"insert": "Heading1"
	},
	{
		"attributes": {
			"header": 1
		},
		"insert": "\n"
	},
	{
		"insert": "Hello, this is text.\nAnd "
	},
	{
		"attributes": {
			"italic": true
		},
		"insert": "here is italic "
	},
	{
		"insert": "(and not).\nAnd "
	},
	{
		"attributes": {
			"bold": true
		},
		"insert": "here is bold"
	},
	{
		"insert": "\n"
	}
]`)

func Example() {
	html, err := quill.Render(ops)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(html))
	// Output: <h1>Heading1</h1><p>Hello, this is text.</p><p>And <em>here is italic </em>(and not).</p><p>And <strong>here is bold</strong></p>
}
