package quill

import (
	"io"
	"strconv"
)

// bold
type boldFormat struct{}

func (*boldFormat) Fmt() *Format {
	return &Format{
		Val:   "strong",
		Place: Tag,
	}
}

func (*boldFormat) HasFormat(o *Op) bool {
	return o.HasAttr("bold")
}

// italic
type italicFormat struct{}

func (*italicFormat) Fmt() *Format {
	return &Format{
		Val:   "em",
		Place: Tag,
	}
}

func (*italicFormat) HasFormat(o *Op) bool {
	return o.HasAttr("italic")
}

// underline
type underlineFormat struct{}

func (*underlineFormat) Fmt() *Format {
	return &Format{
		Val:   "u",
		Place: Tag,
	}
}

func (*underlineFormat) HasFormat(o *Op) bool {
	return o.HasAttr("underline")
}

// text color
type colorFormat struct {
	c string
}

func (cf *colorFormat) Fmt() *Format {
	return &Format{
		Val:   "color:" + cf.c + ";",
		Place: Style,
	}
}

func (cf *colorFormat) HasFormat(o *Op) bool {
	return o.Attrs["color"] == cf.c
}

// link
type linkFormat struct {
	href string
}

func (*linkFormat) Fmt() *Format { return new(Format) } // Only a wrapper.

func (lf *linkFormat) HasFormat(*Op) bool {
	return false
}

func (lf *linkFormat) Wrap() (string, string) {
	return `<a href=` + strconv.Quote(lf.href) + ` target="_blank">`, "</a>"
}

func (lf *linkFormat) Open(_ []*Format, _ *Op) bool {
	return true // This format will only appear when there is a "link" attribute set.
}

func (lf *linkFormat) Close(_ []*Format, o *Op, _ bool) bool {
	return o.Attrs["link"] != lf.href
}

// image
type imageFormat struct {
	src, alt string
}

func (*imageFormat) Fmt() *Format { return nil } // The body contains the entire element.

func (imf *imageFormat) HasFormat(o *Op) bool {
	return o.Type == "image" && o.Data == imf.src
}

// imageFormat implements the FormatWriter interface.
func (imf *imageFormat) Write(buf io.Writer) {

	buf.Write([]byte("<img src="))
	buf.Write([]byte(strconv.Quote(imf.src)))
	if imf.alt != "" {
		buf.Write([]byte(" alt="))
		buf.Write([]byte(strconv.Quote(imf.alt)))
	}
	buf.Write([]byte{'>'})

}
