package quill

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestSimple(t *testing.T) {

	cases := []string{
		`[{"insert": "\n"}]`,
		`[{"insert": "line1\nline2\n"}]`,
		`[{"insert": "line1\n\nline3\n"}]`,
		`[{"insert": "bkqt"}, {"attributes": {"blockquote": true}, "insert": "\n"}]`,
		`[{"attributes": {"color": "#a10000"}, "insert": "colored"}, {"insert": "\n"}]`,
		`[{"insert":"abc "},{"attributes":{"bold":true},"insert":"bld"},{"attributes":{"list":"bullet"},"insert":"\n"}]`,
		`[{"insert":{"image":"source-url"}},{"insert":"\n"}]`,
		`[{"insert":"text "},{"insert":{"image":"source-url"}},{"insert":" more text\n"}]`,
	}

	want := []string{
		"<p><br></p>",
		"<p>line1</p><p>line2</p>",
		"<p>line1</p><p><br></p><p>line3</p>",
		"<blockquote>bkqt</blockquote>",
		`<p><span style="color:#a10000;">colored</span></p>`,
		"<ul><li>abc <strong>bld</strong></li></ul>",
		`<p><img src="source-url"></p>`,
		`<p>text <img src="source-url"> more text</p>`,
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

	pairNames := []string{"ops1", "nested", "ordering", "list1", "list2", "list3", "list4"}

	for _, n := range pairNames {
		if err := testPair(n+".json", n+".html"); err != nil {
			t.Errorf("(name: %s) %s", n, err)
		}
	}

}

func testPair(opsFile, htmlFile string) error {
	ops, err := ioutil.ReadFile("./tests/" + opsFile)
	if err != nil {
		return fmt.Errorf("could not read %s; %s\n", opsFile, err)
	}
	html, err := ioutil.ReadFile("./tests/" + htmlFile)
	if err != nil {
		return fmt.Errorf("could not read %s; %s\n", htmlFile, err)
	}
	got, err := Render(ops)
	if err != nil {
		return fmt.Errorf("error rendering; %s\n", err)
	}
	if !bytes.Equal(html, got) {
		return fmt.Errorf("bad rendering; \nwanted: \n%s\ngot: \n%s\n", html, got)
	}
	return nil
}
