package quill

import (
	"reflect"
	"testing"
)

func TestRawOp_makeOp(t *testing.T) {

	rawOps := []rawOp{
		{
			Insert: "stuff to insert.\n",
			Attrs: map[string]interface{}{
				"bold":      true,
				"link":      "https://widerwebs.com",
				"italic":    false,
				"underline": nil, // nil value is set if JSON value is null
			},
		},
		{
			Insert: "\n",
			Attrs: map[string]interface{}{
				"align": "center",
			},
		},
		{
			Insert: "\n",
			Attrs: map[string]interface{}{
				"align":      "center",
				"blockquote": true,
			},
		},
		{
			Insert: map[string]interface{}{
				"image": "url-or-base64",
			},
		},
	}

	want := []Op{
		{
			Data: "stuff to insert.\n",
			Type: "text",
			Attrs: map[string]string{
				"bold":      "y",
				"italic":    "",
				"link":      "https://widerwebs.com",
				"underline": "",
			},
		},
		{
			Data: "\n",
			Type: "text",
			Attrs: map[string]string{
				"align": "center",
			},
		},
		{
			Data: "\n",
			Type: "text",
			Attrs: map[string]string{
				"align":      "center",
				"blockquote": "y",
			},
		},
		{
			Data:  "url-or-base64",
			Type:  "image",
			Attrs: make(map[string]string), // like in code (already initialized)
		},
	}

	o := new(Op)                         // reuse in loop
	o.Attrs = make(map[string]string, 3) // initialize once here only (as in real code)

	for i := range rawOps {

		if err := rawOps[i].makeOp(o); err != nil {
			t.Errorf("error making Op: %s", err)
			t.FailNow()
		}

		if !reflect.DeepEqual(*o, want[i]) {
			t.Errorf("failed Op comparison; got %+v for index %d", o, i)
		}

	}

}

func TestExtractString(t *testing.T) {
	if extractString("random string") != "random string" {
		t.Errorf("failed stringc extract")
	}
	if extractString(true) != "y" {
		t.Errorf("failed bool true extract")
	}
	if extractString(false) != "" {
		t.Errorf("failed bool false extract")
	}
	if extractString(float64(3)) != "3" {
		t.Errorf("failed float64 extract")
	}
}
