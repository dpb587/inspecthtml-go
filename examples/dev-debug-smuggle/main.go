// Output the intermediate stream containing smuggled metadata which is sent to html.Parse (for debugging purposes).
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/dpb587/inspecthtml-go/inspecthtml"
)

func main() {
	buf, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	_, _, err = inspecthtml.NewParser(bytes.NewReader(buf), inspecthtml.ParserConfig{}.
		SetReaderInterceptor(func(r io.Reader) io.Reader {
			return io.TeeReader(r, os.Stdout)
		}),
	).Parse()
	if err != nil {
		panic(fmt.Errorf("inspecthtml: parse: %v", err))
	}
}
