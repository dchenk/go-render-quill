// Package quill takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations
// and renders the defined HTML document.
//
// This library is designed to be easily extendable. Simply call RenderExtended with a function that may provide its
// own formats for certain kinds of ops and attributes.
package quill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Render takes a Delta array of insert operations and returns the rendered HTML using the built-in settings.
// If an error occurs while rendering, any HTML already rendered is returned.
func Render(ops []byte) ([]byte, error) {
	return RenderExtended(ops, nil)
}

// RenderExtended takes a Delta array of insert operations and, optionally, a function that may provide a Formatter to
// customize the way certain kinds of inserts are rendered, and returns the rendered HTML. If the given Formatter is nil,
// then the default one that is built in is used. If an error occurs while rendering, any HTML already rendered is returned.
func RenderExtended(ops []byte, customFormats func(string, *Op) Formatter) ([]byte, error) {

	raw := make([]rawOp, 0, 12)
	if err := json.Unmarshal(ops, &raw); err != nil {
		return nil, err
	}

	vars := renderVars{
		fs:  make(formatState, 0, 4),
		fms: make([]*Format, 0, 4),
		o:   Op{Attrs: make(map[string]string, 3)},
	}

	for i := range raw {

		if err := raw[i].makeOp(&vars.o); err != nil {
			return vars.finalBuf.Bytes(), err
		}

		vars.fms = vars.fms[:0] // Reset the slice for the current Op iteration.

		// To set up fms, first check the Op insert type.
		typeFmTer := vars.o.getFormatter(vars.o.Type, customFormats)
		if typeFmTer == nil {
			return vars.finalBuf.Bytes(), fmt.Errorf("quill: an op does not have a format defined for its type: %v", raw[i])
		}
		vars.o.addFmTer(&vars, typeFmTer)

		// Get a Formatter out of each of the attributes.
		for attr := range vars.o.Attrs {
			vars.o.addFmTer(&vars, vars.o.getFormatter(attr, customFormats))
		}

		// Open a block element, write its body, and close it to move on only when the ending "\n" of the block is reached.
		if strings.IndexByte(vars.o.Data, '\n') != -1 {

			// Extract text from between the block-terminating line feeds and write each part as its own Op.
			split := strings.Split(vars.o.Data, "\n")

			for j := range split {

				vars.o.Data = split[j]

				// If the current o.Data still has an "\n" following (its not the last in split), then it ends a block.
				if j < len(split)-1 {

					vars.o.writeBlock(&vars)

				} else if vars.o.Data != "" { // If the last element in split is just "" then the last character in the rawOp is "\n".

					vars.o.writeInline(&vars)

				}

			}

		} else {
			vars.o.writeInline(&vars)
		}

	}

	// Before writing out the final buffer, close the last remaining tags set by a FormatWrapper.
	// The FormatWrapper should see that all styling is now done.
	vars.fs.closePrevious(&vars.finalBuf, blankOp(), true)

	return vars.finalBuf.Bytes(), nil

}

// renderVars combines the variables created in RenderExtended into a single allocation.
type renderVars struct {
	finalBuf bytes.Buffer // the final output
	tempBuf  bytes.Buffer // temporary buffer reused for each block element
	fs       formatState  // the tags currently open in the order in which they were opened
	fms      []*Format    // reused slice for the the Formatter types defined for each Op
	o        Op           // an Op to reuse for all iterations
}

// addFmTer adds the format from fmTer to fms (the temporary, current Op's formats) if the format is not already set in the
// current format state. All FormatWrapper formats are added regardless of whether they are already set on fs. Data is written
// to the temporary buffer only.
func (o *Op) addFmTer(vars *renderVars, fmTer Formatter) {
	if fmTer == nil {
		return
	}
	fm := fmTer.Fmt()
	if fm == nil {
		// Check if the format is a FormatWriter. If it is, just write it out and continue.
		if wr, ok := fmTer.(FormatWriter); ok {
			wr.Write(&vars.tempBuf)
			o.Data = ""
		}
		return
	}
	fm.fm = fmTer
	if fw, ok := fmTer.(FormatWrapper); ok {
		fm.wrap = true
		fm.wrapPre, fm.wrapPost = fw.Wrap()
		vars.fms = append(vars.fms, fm)
		return
	}
	if !vars.fs.hasSet(fmTer.Fmt()) {
		vars.fms = append(vars.fms, fm)
	}
}

// An Op is a Delta insert operations (https://github.com/quilljs/delta#insert) that has been converted into this format for
// usability with the type safety in Go.
type Op struct {
	Data  string            // the text to insert or the value of the embed object (http://quilljs.com/docs/delta/#embeds)
	Type  string            // the type of the op (typically "text", but any other type can be registered)
	Attrs map[string]string // key is attribute name; value is either the attribute value or "y" (meaning true)
}

// writeBlock writes a block element (which may be nested inside another block element if it is a FormatWrapper).
// The opening HTML tag of a block element is written to the main buffer only after the "\n" character terminating the
// block is reached (the Op with the "\n" character holds the information about the block element).
func (o *Op) writeBlock(vars *renderVars) {

	// Close the inline formats opened within the block to the tempBuf and block formats of wrappers to finalBuf.
	closedTemp := make(formatState, 0, 1)

	for i := len(vars.fs) - 1; i >= 0; i-- { // Start with the last format opened.

		f := vars.fs[i]

		// If this format is not set on the current Op, close it.
		if (!f.wrap && !f.fm.HasFormat(o)) || (f.wrap && f.fm.(FormatWrapper).Close(vars.fs, o, true)) {

			// If we need to close a tag after which there are tags that should stay open, close the following tags for now.
			if i < len(vars.fs)-1 {
				for ij := len(vars.fs) - 1; ij > i; ij-- {
					closedTemp.add(vars.fs[ij])
					if f.wrap && f.Block {
						vars.fs.pop(&vars.finalBuf)
					} else {
						vars.fs.pop(&vars.tempBuf)
					}
				}
			}

			if f.wrap && f.Block {
				vars.fs.pop(&vars.finalBuf)
			} else {
				vars.fs.pop(&vars.tempBuf)
			}

		}

	}

	// Re-open the temporarily closed formats.
	closedTemp.writeFormats(&vars.tempBuf)
	vars.fs = append(vars.fs, closedTemp...) // Copy after the sorting.

	var block struct {
		tagName string
		classes []string
		style   string
	}

	// Merge all formats into a single tag.
	for i := range vars.fms {
		fm := vars.fms[i]
		// Apply only block-level formats.
		if fm.Block {
			v := fm.Val
			switch fm.Place {
			case Tag:
				// If an opening tag is not specified by the Op insert type, it may be specified by an attribute.
				block.tagName = v // Override whatever value is set.
			case Class:
				block.classes = append(block.classes, v)
			case Style:
				block.style += v
			}
		}
		// Write out all of FormatWrapper opening text (if there is any).
		if fm.wrap && fm.fm.(FormatWrapper).Open(vars.fs, o) {
			fm.Val = fm.wrapPre
			vars.fs.add(fm)
			vars.finalBuf.WriteString(fm.Val)
		}
	}

	// Avoid empty paragraphs and "\n" in the output for text blocks.
	if o.Data == "" && block.tagName == "p" && vars.tempBuf.Len() == 0 {
		o.Data = "<br>"
	}

	if block.tagName != "" {
		vars.finalBuf.WriteByte('<')
		vars.finalBuf.WriteString(block.tagName)
		vars.finalBuf.WriteString(classesList(block.classes))
		if block.style != "" {
			vars.finalBuf.WriteString(" style=")
			vars.finalBuf.WriteString(strconv.Quote(block.style))
		}
		vars.finalBuf.WriteByte('>')
	}

	vars.finalBuf.Write(vars.tempBuf.Bytes()) // Copy the temporary buffer to the final output.

	vars.finalBuf.WriteString(o.Data) // Copy the data of the current Op (usually just "<br>" or blank).

	if block.tagName != "" {
		closeTag(&vars.finalBuf, block.tagName)
	}

	vars.tempBuf.Reset()

}

// writeInline writes to the temporary buffer.
func (o *Op) writeInline(vars *renderVars) {

	vars.fs.closePrevious(&vars.tempBuf, o, false)

	// Save the formats being written now separately from fs.
	addNow := make(formatState, 0, len(vars.fms))

	for _, f := range vars.fms {
		// Apply only inline formats.
		if !f.Block {
			if f.wrap {
				// Add FormatWrapper formats only if they need to be written now.
				if f.fm.(FormatWrapper).Open(vars.fs, o) {
					f.Val = f.wrapPre
					addNow.add(f)
				}
			} else {
				addNow.add(f)
			}
		}
	}

	addNow.writeFormats(&vars.tempBuf)
	vars.fs = append(vars.fs, addNow...) // Copy after the sorting.

	vars.tempBuf.WriteString(o.Data)

}

// HasAttr says if the Op is not nil and has the attribute set to a non-blank value.
func (o *Op) HasAttr(attr string) bool {
	return o != nil && o.Attrs[attr] != ""
}

// getFormatter returns a formatter based on the keyword (either "text" or "" or an attribute name) and the Op settings.
// For every Op, first its Type is passed through here as the keyword, and then its attributes.
func (o *Op) getFormatter(keyword string, customFormats func(string, *Op) Formatter) Formatter {

	if customFormats != nil {
		if custom := customFormats(keyword, o); custom != nil {
			return custom
		}
	}

	switch keyword { // This is the list of currently recognized "keywords".
	case "text":
		return new(textFormat)
	case "header":
		return &headerFormat{
			level: o.Attrs["header"],
		}
	case "list":
		lf := &listFormat{
			indent: indentDepths[o.Attrs["indent"]],
		}
		if o.Attrs["list"] == "bullet" {
			lf.lType = "ul"
		} else {
			lf.lType = "ol"
		}
		return lf
	case "blockquote":
		return new(blockQuoteFormat)
	case "align":
		return &alignFormat{
			val: o.Attrs["align"],
		}
	case "image":
		return &imageFormat{
			src: o.Data,
		}
	case "link":
		return &linkFormat{
			href: o.Attrs["link"],
		}
	case "bold":
		return new(boldFormat)
	case "size":
		return sizeFormat(o.Attrs["size"])
	case "italic":
		return new(italicFormat)
	case "underline":
		return new(underlineFormat)
	case "color":
		return &colorFormat{
			c: o.Attrs["color"],
		}
	case "indent":
		return &indentFormat{
			in: o.Attrs["indent"],
		}
	case "strike":
		return new(strikeFormat)
	case "background":
		return &bkgFormat{
			c: o.Attrs["background"],
		}
	case "script":
		sf := new(scriptFormat)
		if o.Attrs["script"] == "super" {
			sf.t = "sup"
		} else {
			sf.t = "sub"
		}
		return sf
	case "code-block":
		return &codeBlockFormat{o}
	}

	return nil

}

// A FormatPlace is either an HTML tag name, a CSS class, or a style attribute value.
type FormatPlace uint8

const (
	Tag FormatPlace = iota
	Class
	Style
)

// A Formatter is able to give a Format and say whether a given Op should have that Format applied.
type Formatter interface {
	Fmt() *Format       // Format gives the string to write and where to place it.
	HasFormat(*Op) bool // Say if the Op has the Format that Fmt returns.
}

// A FormatWriter can write the body of an Op in a custom way (useful for embeds).
type FormatWriter interface {
	Formatter
	Write(io.Writer) // Write the entire body of the element.
}

// A FormatWrapper wraps text with additional text of any kind (such as "<ul>" for lists).
type FormatWrapper interface {
	Formatter
	Wrap() (pre, post string)        // Say what opening and closing wraps will be written.
	Open([]*Format, *Op) bool        // Given the open formats and current Op, say if to write the pre string.
	Close([]*Format, *Op, bool) bool // Given the open formats, current Op, and if the Op closes a block, say if to write the post string.
}

// A Format specifies how styling to text is applied. The Val string is what is printed in the place given by Place. Block indicates
// if this is a block-level format.
type Format struct {
	Val               string      // the value to print
	Place             FormatPlace // where this format is placed in the text
	Block             bool        // indicate whether this is a block-level format (not printed until a "\n" is reached)
	wrap              bool        // indicates whether this format was written as a FormatWrapper
	wrapPre, wrapPost string      // If this Format is a wrap, then Val holds the open and wrapPost holds the close.
	fm                Formatter   // where this instance of a Format came from
}

// A blankOp can be used to signal any FormatWrapper formats to write the final closing wrap.
func blankOp() *Op {
	return &Op{"", "text", make(map[string]string)}
}

// If cl has something, then classesList returns the class attribute to add to an HTML element with a space before the
// "class" attribute and spaces between each class name.
func classesList(cl []string) string {
	if len(cl) > 0 {
		return " class=" + strconv.Quote(strings.Join(cl, " "))
	}
	return ""
}

// closeTag writes a complete closing tag to buf.
func closeTag(buf *bytes.Buffer, tagName string) {
	buf.WriteString("</")
	buf.WriteString(tagName)
	buf.WriteByte('>')
}
