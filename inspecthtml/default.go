package inspecthtml

import (
	"io"

	"golang.org/x/net/html"
)

func Parse(r io.Reader) (*html.Node, *ParseMetadata, error) {
	return NewParser(r).Parse()
}

func ParseWithOptions(r io.Reader, opts ...html.ParseOption) (*html.Node, *ParseMetadata, error) {
	return NewParser(r).ParseWithOptions(opts...)
}
