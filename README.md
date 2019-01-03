# go-render-quill [![Build Status](https://travis-ci.org/dchenk/go-render-quill.svg?branch=master)](https://travis-ci.org/dchenk/go-render-quill)

Package `quill` takes a Quill-based Delta (https://github.com/quilljs/delta) as a JSON array of `insert` operations
and renders the defined HTML document.

Complete documentation at GoDoc: [https://godoc.org/github.com/dchenk/go-render-quill](https://godoc.org/github.com/dchenk/go-render-quill)

## Usage

```
import "github.com/dchenk/go-render-quill"

var delta = `[{"insert":"This "},{"attributes":{"italic":true},"insert":"is"},
    {"insert":" "},{"attributes":{"bold":true},"insert":"great!"},{"insert":"\n"}]`

html, err := quill.Render(delta)
if err != nil {
	panic(err)
}
fmt.Println(string(html))
// Output: <p>This <em>is</em> <strong>great!</strong></p>
```

## Supported Formats

### Inline
 - Background color
 - Bold
 - Text color
 - Italic
 - Link
 - Size
 - Strikethrough
 - Superscript/Subscript
 - Underline

### Block
 - Blockquote
 - Header
 - Indent
 - List (ul and ol, including nested lists)
 - Text alignment
 - Code block

### Embeds
 - Image (an inline format)

## Extending

The simple `Formatter` interface is all you need to implement for most block and inline formats. Instead of `Render` use `RenderExtended`
and provide a function that returns a `Formatter` for inserts that have the format you need.

For more control, you can also implement `FormatWriter` or `FormatWrapper`.
