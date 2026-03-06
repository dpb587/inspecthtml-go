// Output the intermediate stream containing smuggled metadata which is sent to html.Parse (for debugging purposes).
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dpb587/inspecthtml-go/inspecthtml"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)

		os.Exit(1)
	}
}

func mainErr() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <html-file>", filepath.Base(os.Args[0]))
	}

	buf, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}

	_, _, err = inspecthtml.NewParser(bytes.NewReader(buf), inspecthtml.ParserConfig{}.
		SetReaderInterceptor(func(r io.Reader) io.Reader {
			return io.TeeReader(r, os.Stdout)
		}),
	).Parse()
	if err != nil {
		return fmt.Errorf("inspecthtml: parse: %v", err)
	}

	return nil
}
