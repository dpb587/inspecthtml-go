// Compare the re-rendered results of html.Parse vs inspecthtml.Parse for a given input file to identify any discrepancies.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dpb587/inspecthtml-go/inspecthtml"
	"github.com/pmezard/go-difflib/difflib"
	"golang.org/x/net/html"
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

	err = func() error {
		htmlRoot, err := html.Parse(bytes.NewReader(buf))
		if err != nil {
			return fmt.Errorf("html: parse: %v", err)
		}

		inspecthtmlRoot, _, err := inspecthtml.Parse(bytes.NewReader(buf))
		if err != nil {
			return fmt.Errorf("inspecthtml: parse: %v", err)
		}

		//

		htmlRender := &bytes.Buffer{}

		if err := html.Render(htmlRender, htmlRoot); err != nil {
			return fmt.Errorf("html render: render: %v", err)
		}

		inspecthtmlRender := &bytes.Buffer{}

		if err := html.Render(inspecthtmlRender, inspecthtmlRoot); err != nil {
			return fmt.Errorf("inspecthtml render: %v", err)
		}

		if !bytes.Equal(htmlRender.Bytes(), inspecthtmlRender.Bytes()) {
			if true {
				patchText, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
					A:        difflib.SplitLines(htmlRender.String()),
					B:        difflib.SplitLines(inspecthtmlRender.String()),
					FromFile: "html",
					ToFile:   "inspecthtml",
					Context:  3,
				})
				if err != nil {
					return fmt.Errorf("render mismatch: diff: %v", err)
				}

				return fmt.Errorf("render mismatch\n\n%s", patchText)
			} else {
				return fmt.Errorf("render mismatch\n\n---html---\n%s\n\n---inspecthtml---\n%s", htmlRender, inspecthtmlRender)
			}
		}

		return nil
	}()
	if err != nil {
		return fmt.Errorf("file[%s]: %v", os.Args[1], err)
	}

	return nil
}
