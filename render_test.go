package quill

import (
	"bytes"
	"fmt"
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

	pairNames := []string{"ops1", "nested", "ordering", "list1", "list2", "list3", "list4", "indent"}

	for _, n := range pairNames {
		if err := testPair(n+".json", n+".html"); err != nil {
			t.Errorf("(name: %s) %s", n, err)
		}
	}

}

func testPair(opsFile, htmlFile string) error {
	ops, err := ioutil.ReadFile("./tests/" + opsFile)
	if err != nil {
		return fmt.Errorf("could not read %s; %s", opsFile, err)
	}
	html, err := ioutil.ReadFile("./tests/" + htmlFile)
	if err != nil {
		return fmt.Errorf("could not read %s; %s", htmlFile, err)
	}
	got, err := Render(ops)
	if err != nil {
		return fmt.Errorf("error rendering; %s", err)
	}
	if !bytes.Equal(html, got) {
		return fmt.Errorf("bad rendering; \nwanted: \n%s\ngot: \n%s", html, got)
	}
	return nil
}
