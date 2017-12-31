package quill

import (
	"bytes"
	"sort"
	"testing"
)

func TestFormatState_add(t *testing.T) {

	cases := []struct {
		current []*Format
		keyword string
		o       *Op
		want    []*Format
	}{
		{
			current: []*Format{}, // no formats
			keyword: "italic",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"italic": "y"},
			},
			want: []*Format{
				{
					Val:   "em",
					Place: Tag,
				},
			},
		},
		{
			current: []*Format{
				{ // One format already set.
					Val:   "em",
					Place: Tag,
				},
			},
			keyword: "italic",
			o: &Op{
				Data:  "stuff",
				Type:  "text",
				Attrs: map[string]string{"italic": "y"},
			},
			want: []*Format{ // The way add works, it does not check if the format is already added.
				{
					Val:   "em",
					Place: Tag,
				},
				{
					Val:   "em",
					Place: Tag,
				},
			},
		},
	}

	fs := new(formatState) // reuse

	for i, ca := range cases {

		fs.open = ca.current

		fmTer := ca.o.getFormatter(ca.keyword, nil)
		fm := fmTer.Fmt()
		fm.fm = fmTer

		fs.add(fm)

		if len(ca.want) != len(fs.open) {
			t.Errorf("(index %d) unequal count of formats", i)
			t.FailNow()
		}

		for j := range fs.open {
			if ca.want[j].Val != fs.open[j].Val {
				t.Errorf("did not add format Val correctly (index %d); got %q", i, fs.open[j].Val)
			}
			if ca.want[j].Place != fs.open[j].Place {
				t.Errorf("did not add format Place correctly (index %d); got %v", i, fs.open[j].Place)
			}
			if ca.want[j].Block != fs.open[j].Block {
				t.Errorf("did not add format Block correctly (index %d); got %v", i, fs.open[j].Block)
			}
		}

	}

}

func TestFormatState_closePrevious(t *testing.T) {

	o := &Op{
		Data: "stuff",
		Type: "text",
		// no attributes set
	}

	cases := []formatState{
		{[]*Format{
			{"em", Tag, false, false, "", "", o.getFormatter("italic", nil)},
			{"strong", Tag, false, false, "", "", o.getFormatter("bold", nil)},
		}},
	}

	want := []string{"</strong></em>", "</strong></li></ul>"}

	buf := new(bytes.Buffer)

	for i := range cases {

		cases[i].closePrevious(buf, o, false)
		got := buf.String()
		if got != want[i] {
			t.Errorf("closed formats wrong (index %d); wanted %q; got %q\n", i, want[i], got)
		}

		buf.Reset()

	}

}

func TestFormatState_Sort(t *testing.T) {

	o := &Op{
		Data: "stuff",
		Type: "text",
		// no attributes set
	}

	cases := [][]struct {
		Val     string
		Place   FormatPlace
		keyword string
	}{
		{
			{"strong", Tag, "bold"},
			{"em", Tag, "italic"},
		},
		{
			{"u", Tag, "underline"},
			{"align-center", Class, "align"},
			{"strong", Tag, "bold"},
		},
		{
			{"color:#e0e0e0;", Style, "color"},
			{"em", Tag, "italic"},
		},
		{
			{"em", Tag, "italic"},
			{`<a href="https://widerwebs.com" target="_blank">`, Tag, "link"}, // link wrapper
		},
	}

	want := [][]struct {
		Val     string
		Place   FormatPlace
		keyword string
	}{
		{
			{"em", Tag, "italic"},
			{"strong", Tag, "bold"},
		},
		{
			{"strong", Tag, "bold"},
			{"u", Tag, "underline"},
			{"align-center", Class, "align"},
		},
		{
			{"em", Tag, "italic"},
			{"color:#e0e0e0;", Style, "color"},
		},
		{
			{`<a href="https://widerwebs.com" target="_blank">`, Tag, "link"}, // link wrapper
			{"em", Tag, "italic"},
		},
	}

	for i := range cases {

		fsCase := new(formatState)
		for _, s := range cases[i] {
			fsCase.open = append(fsCase.open, &Format{
				Val:   s.Val,
				Place: s.Place,
				fm:    o.getFormatter(s.keyword, nil),
			})
		}

		sort.Sort(fsCase)

		fsWant := new(formatState)
		for _, s := range want[i] {
			fsWant.open = append(fsWant.open, &Format{
				Val:   s.Val,
				Place: s.Place,
				fm:    o.getFormatter(s.keyword, nil),
			})
		}

		ok := true
		for j := range fsCase.open {
			if fsCase.open[j].Val != fsWant.open[j].Val {
				ok = false
			} else if fsCase.open[j].Place != fsWant.open[j].Place {
				ok = false
			}
		}
		if !ok {
			t.Errorf("bad sorting (index %d); got:\n", i)
			for k := range fsCase.open {
				t.Errorf("  (%d) %+v\n", k, *fsCase.open[k])
			}
		}

	}

}
