package quill

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestSimple(t *testing.T) {

	cases := []string{
		`[{"insert": "\n"}]`,                                                                                             // blank
		`[{"insert": "line1\nline2\n"}]`,                                                                                 // two paragraphs (single op)
		`[{"insert": "line1\n\nline3\n"}]`,                                                                               // blank line
		`[{"insert": "bkqt"}, {"attributes": {"blockquote": true}, "insert": "\n"}]`,                                     // blockquote
		`[{"attributes": {"color": "#a10000"}, "insert": "colored"}, {"insert": "\n"}]`,                                  // color
		`[{"attributes":{"strike":true},"insert":"striked"},{"insert":"\n"}]`,                                            // strikethrough
		`[{"insert":"abc "},{"attributes":{"bold":true},"insert":"bld"},{"attributes":{"list":"bullet"},"insert":"\n"}]`, // list
		`[{"insert":{"image":"source-url"}},{"insert":"\n"}]`,                                                            // image
		`[{"insert":"text "},{"insert":{"image":"source-url"}},{"insert":" more text\n"}]`,                               // image
		`[{"insert":"abc "},{"attributes":{"background":"#66a3e0"},"insert":"colored"},{"insert":" plain\n"}]`,           // background
		`[{"attributes":{"underline":true},"insert":"underlined"},{"insert":"\n"}]`,                                      // underlined
		`[{"insert":"plain"},{"attributes":{"script":"super"},"insert":"super"},{"insert":"\n"}]`,                        // superscript
		`[{"insert":"plain"},{"attributes":{"script":"sub"},"insert":"sub"},{"insert":"\n"}]`,                            // subscript
	}

	want := []string{
		"<p><br></p>",
		"<p>line1</p><p>line2</p>",
		"<p>line1</p><p><br></p><p>line3</p>",
		"<blockquote>bkqt</blockquote>",
		`<p><span style="color:#a10000;">colored</span></p>`,
		"<p><s>striked</s></p>",
		"<ul><li>abc <strong>bld</strong></li></ul>",
		`<p><img src="source-url"></p>`,
		`<p>text <img src="source-url"> more text</p>`,
		`<p>abc <span style="background-color:#66a3e0;">colored</span> plain</p>`,
		"<p><u>underlined</u></p>",
		"<p>plain<sup>super</sup></p>",
		"<p>plain<sub>sub</sub></p>",
	}

	for i := range cases {

		bts, err := Render([]byte(cases[i]))
		if err != nil {
			t.Errorf("%s", err)
			t.FailNow()
		}
		if string(bts) != want[i] {
			t.Errorf("bad rendering (index %d); got: %s", i, bts)
		}

	}

}

func TestRender(t *testing.T) {

	pairNames := []string{"ops1", "nested", "ordering", "list1", "list2", "list3", "list4", "indent", "code1", "code2"}

	for _, n := range pairNames {

		ops, err := ioutil.ReadFile("./tests/" + n + ".json")
		if err != nil {
			t.Errorf("could not read %s.json; %s", n, err)
			t.FailNow()
		}

		html, err := ioutil.ReadFile("./tests/" + n + ".html")
		if err != nil {
			t.Errorf("could not read %s.html; %s", n, err)
			t.FailNow()
		}

		got, err := Render(ops)
		if err != nil {
			t.Errorf("error rendering; %s", err)
		}

		if !bytes.Equal(html, got) {
			t.Errorf("bad rendering (name: %s):\nwanted: \n%s\ngot: \n%s", n, html, got)
		}

	}

}
