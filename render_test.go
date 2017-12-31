package quill

import "fmt"

func Example() {
	ops := []byte(`[
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
		}
	]`)
	fmt.Println(Render(ops))
	// Output: <h1>Heading1</h1><p>Hello, this is text.</p><p>And <em>here is italic </em>(and not).</p><p>And <strong>here is bold</strong>
}
