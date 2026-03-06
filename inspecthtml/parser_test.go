package inspecthtml

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func visitNode(n *html.Node, f func(n *html.Node)) {
	f(n)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visitNode(c, f)
	}
}

func TestReaderHeadSpaceBug(t *testing.T) {
	document, _, err := Parse(strings.NewReader("<html><head> <template>content</template> <meta/> </head></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// previous implementation replaced whitespace with non-whitespace
	// which has significance to the tokenizer (and started the body early)

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head> <template>content</template> <meta/> </head><body></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTag(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagNameOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       13,
				LineColumn: cursorio.TextLineColumn{0, 13},
			},
			Until: cursorio.TextOffset{
				Byte:       14,
				LineColumn: cursorio.TextLineColumn{0, 14},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("tag name: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       24,
				LineColumn: cursorio.TextLineColumn{0, 24},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetInnerOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("inner: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p>hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttr(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p class=\"text-sm\">hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       21,
				LineColumn: cursorio.TextLineColumn{0, 21},
			},
			Until: cursorio.TextOffset{
				Byte:       30,
				LineColumn: cursorio.TextLineColumn{0, 30},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p class=\"text-sm\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrQuoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title=\"a &quot; mark\">hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       21,
				LineColumn: cursorio.TextLineColumn{0, 21},
			},
			Until: cursorio.TextOffset{
				Byte:       36,
				LineColumn: cursorio.TextLineColumn{0, 36},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"a &#34; mark\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrSingleQuoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title='a &quot; mark'>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       21,
				LineColumn: cursorio.TextLineColumn{0, 21},
			},
			Until: cursorio.TextOffset{
				Byte:       36,
				LineColumn: cursorio.TextLineColumn{0, 36},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"a &#34; mark\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrEmptyUnquoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title=>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets != nil {
			t.Errorf("tag attr value: expected nil, got non-nil")
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrNull(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets != nil {
			t.Errorf("tag attr value: expected nil, got non-nil")
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrNullOther(t *testing.T) {
	document, _, err := Parse(strings.NewReader(`<address itemscope itemtype="http://microformats.org/profile/hcard"/>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), `<html><head></head><body><address itemscope="" itemtype="http://microformats.org/profile/hcard"></address></body></html>`; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrUnquoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title=none>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       21,
				LineColumn: cursorio.TextLineColumn{0, 21},
			},
			Until: cursorio.TextOffset{
				Byte:       25,
				LineColumn: cursorio.TextLineColumn{0, 25},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"none\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrSingleCharUnquotedBeforeQuoted(t *testing.T) {
	// Regression test: single-character unquoted attribute value before quoted attribute
	// Before fix, the regex would skip the first character when searching for unquoted values
	document, documentOffsets, err := Parse(strings.NewReader(`<html><body data-rsssl=1 class="test">hello</body></html>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.Body {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 2; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       22,
				LineColumn: cursorio.TextLineColumn{0, 22},
			},
		}); _a != _e {
			t.Errorf("tag attr[0] key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr[0] value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       23,
				LineColumn: cursorio.TextLineColumn{0, 23},
			},
			Until: cursorio.TextOffset{
				Byte:       24,
				LineColumn: cursorio.TextLineColumn{0, 24},
			},
		}); *_a != _e {
			t.Errorf("tag attr[0] value: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[1].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       25,
				LineColumn: cursorio.TextLineColumn{0, 25},
			},
			Until: cursorio.TextOffset{
				Byte:       30,
				LineColumn: cursorio.TextLineColumn{0, 30},
			},
		}); _a != _e {
			t.Errorf("tag attr[1] key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[1].ValueOffsets == nil {
			t.Errorf("tag attr[1] value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[1].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       31,
				LineColumn: cursorio.TextLineColumn{0, 31},
			},
			Until: cursorio.TextOffset{
				Byte:       37,
				LineColumn: cursorio.TextLineColumn{0, 37},
			},
		}); *_a != _e {
			t.Errorf("tag attr[1] value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), `<html><head></head><body data-rsssl="1" class="test">hello</body></html>`; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrInvalidQuoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title=\"quoted\"suffix\">hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 2; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       21,
				LineColumn: cursorio.TextLineColumn{0, 21},
			},
			Until: cursorio.TextOffset{
				Byte:       29,
				LineColumn: cursorio.TextLineColumn{0, 29},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"quoted\" suffix\"=\"\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrSpaceQuoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title =\"quoted\">hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       22,
				LineColumn: cursorio.TextLineColumn{0, 22},
			},
			Until: cursorio.TextOffset{
				Byte:       30,
				LineColumn: cursorio.TextLineColumn{0, 30},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"quoted\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrSpaceUnquoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title =unquoted>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       22,
				LineColumn: cursorio.TextLineColumn{0, 22},
			},
			Until: cursorio.TextOffset{
				Byte:       30,
				LineColumn: cursorio.TextLineColumn{0, 30},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"unquoted\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrSpaceSpaceQuoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title = \"quoted\">hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       23,
				LineColumn: cursorio.TextLineColumn{0, 23},
			},
			Until: cursorio.TextOffset{
				Byte:       31,
				LineColumn: cursorio.TextLineColumn{0, 31},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"quoted\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrSpaceSpaceUnquoted(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p title = unquoted>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Errorf("tag attr count: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("tag attr key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Errorf("tag attr value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       23,
				LineColumn: cursorio.TextLineColumn{0, 23},
			},
			Until: cursorio.TextOffset{
				Byte:       31,
				LineColumn: cursorio.TextLineColumn{0, 31},
			},
		}); *_a != _e {
			t.Errorf("tag attr value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p title=\"unquoted\">hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagClosedByParent(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p>hello</body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagNameOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       13,
				LineColumn: cursorio.TextLineColumn{0, 13},
			},
			Until: cursorio.TextOffset{
				Byte:       14,
				LineColumn: cursorio.TextLineColumn{0, 14},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("tag name: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetInnerOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("inner: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p>hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagClosedBySibling(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><dl><dt>hello<dd>world</dl></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.Dt {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       16,
				LineColumn: cursorio.TextLineColumn{0, 16},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagNameOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
			Until: cursorio.TextOffset{
				Byte:       19,
				LineColumn: cursorio.TextLineColumn{0, 19},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("tag name: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       16,
				LineColumn: cursorio.TextLineColumn{0, 16},
			},
			Until: cursorio.TextOffset{
				Byte:       25,
				LineColumn: cursorio.TextLineColumn{0, 25},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetInnerOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
			Until: cursorio.TextOffset{
				Byte:       25,
				LineColumn: cursorio.TextLineColumn{0, 25},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("inner: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><dl><dt>hello</dt><dd>world</dd></dl></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagEmptyClosedByParent(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.P {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagNameOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       13,
				LineColumn: cursorio.TextLineColumn{0, 13},
			},
			Until: cursorio.TextOffset{
				Byte:       14,
				LineColumn: cursorio.TextLineColumn{0, 14},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("tag name: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetInnerOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("inner: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p></p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagEmptyClosedBySibling(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><dl><dt><dd>world</dl></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.Dt {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       16,
				LineColumn: cursorio.TextLineColumn{0, 16},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.TagNameOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
			Until: cursorio.TextOffset{
				Byte:       19,
				LineColumn: cursorio.TextLineColumn{0, 19},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("tag name: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       16,
				LineColumn: cursorio.TextLineColumn{0, 16},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetInnerOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a == nil || *_a != _e {
			t.Errorf("inner: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><dl><dt></dt><dd>world</dd></dl></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextWithinTag(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p>hello</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p>hello</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextEntityWithinTag(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p>hello &amp; world</p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       32,
				LineColumn: cursorio.TextLineColumn{0, 32},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       32,
				LineColumn: cursorio.TextLineColumn{0, 32},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p>hello &amp; world</p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextBetweenTag(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p></p>hello<p></p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       19,
				LineColumn: cursorio.TextLineColumn{0, 19},
			},
			Until: cursorio.TextOffset{
				Byte:       24,
				LineColumn: cursorio.TextLineColumn{0, 24},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       19,
				LineColumn: cursorio.TextLineColumn{0, 19},
			},
			Until: cursorio.TextOffset{
				Byte:       24,
				LineColumn: cursorio.TextLineColumn{0, 24},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p></p>hello<p></p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextAfterSelfClosing(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><br/>hello<p></p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
			Until: cursorio.TextOffset{
				Byte:       22,
				LineColumn: cursorio.TextLineColumn{0, 22},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
			Until: cursorio.TextOffset{
				Byte:       22,
				LineColumn: cursorio.TextLineColumn{0, 22},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><br/>hello<p></p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextBeforeSelfClosing(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body>hello<br/><p></p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body>hello<br/><p></p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextAfterComment(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><!-- -->hello<p></p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
			Until: cursorio.TextOffset{
				Byte:       25,
				LineColumn: cursorio.TextLineColumn{0, 25},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
			Until: cursorio.TextOffset{
				Byte:       25,
				LineColumn: cursorio.TextLineColumn{0, 25},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><!-- -->hello<p></p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextBeforeComment(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body>hello<!-- --><p></p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       12,
				LineColumn: cursorio.TextLineColumn{0, 12},
			},
			Until: cursorio.TextOffset{
				Byte:       17,
				LineColumn: cursorio.TextLineColumn{0, 17},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body>hello<!-- --><p></p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTextReparentedMerged(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><table>one<tr>two<td>three</td></tr></table></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var foundOne, foundTwo bool

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		} else if n.Data == "one" {
			np, ok := documentOffsets.GetNodeMetadata(n)
			if !ok {
				t.Fatal("expected metadata")
			} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
				From: cursorio.TextOffset{
					Byte:       19,
					LineColumn: cursorio.TextLineColumn{0, 19},
				},
				Until: cursorio.TextOffset{
					Byte:       22,
					LineColumn: cursorio.TextLineColumn{0, 22},
				},
			}); _a != _e {
				t.Errorf("token: expected %v, got %v", _e, _a)
			} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
				From: cursorio.TextOffset{
					Byte:       19,
					LineColumn: cursorio.TextLineColumn{0, 19},
				},
				Until: cursorio.TextOffset{
					Byte:       22,
					LineColumn: cursorio.TextLineColumn{0, 22},
				},
			}); _a != _e {
				t.Errorf("outer: expected %v, got %v", _e, _a)
			}

			foundOne = true
		} else if n.Data == "two" {
			np, ok := documentOffsets.GetNodeMetadata(n)
			if !ok {
				t.Fatal("expected metadata")
			} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
				From: cursorio.TextOffset{
					Byte:       26,
					LineColumn: cursorio.TextLineColumn{0, 26},
				},
				Until: cursorio.TextOffset{
					Byte:       29,
					LineColumn: cursorio.TextLineColumn{0, 29},
				},
			}); _a != _e {
				t.Errorf("token: expected %v, got %v", _e, _a)
			} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
				From: cursorio.TextOffset{
					Byte:       26,
					LineColumn: cursorio.TextLineColumn{0, 26},
				},
				Until: cursorio.TextOffset{
					Byte:       29,
					LineColumn: cursorio.TextLineColumn{0, 29},
				},
			}); _a != _e {
				t.Errorf("outer: expected %v, got %v", _e, _a)
			}

			foundTwo = true
		}
	})

	if !foundOne {
		t.Error("expected to find 'one'")
	} else if !foundTwo {
		t.Error("expected to find 'two'")
	}

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body>onetwo<table><tbody><tr><td>three</td></tr></tbody></table></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderUnopenedRootBug(t *testing.T) {
	document, _, err := Parse(strings.NewReader("</p>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderUnopenedGenericTagBug(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body></nav>hello</body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		} else if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       18,
				LineColumn: cursorio.TextLineColumn{0, 18},
			},
			Until: cursorio.TextOffset{
				Byte:       23,
				LineColumn: cursorio.TextLineColumn{0, 23},
			},
		}); _a != _e {
			t.Errorf("token: expected %v, got %v", _e, _a)
		} else if _a, _e := np.GetOuterOffsets(), (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       18,
				LineColumn: cursorio.TextLineColumn{0, 18},
			},
			Until: cursorio.TextOffset{
				Byte:       23,
				LineColumn: cursorio.TextLineColumn{0, 23},
			},
		}); _a != _e {
			t.Errorf("outer: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body>hello</body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func dumpTraversal(n *html.Node) string {
	if n == nil {
		return "/"
	}

	return dumpTraversal(n.Parent) + "/" + n.Data
}

func TestTokenizerElementInterrupt(t *testing.T) {
	document, documentOffsets, err := Parse(strings.NewReader("<html><body><p><custom-element><ul><li>hello</li></ul></custom-element></p></body></html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedNodeMetadataListIdx := -1
	expectedNodeMetadataList := []*NodeMetadata{
		nil, // Document
		{ // html
			TokenOffsets: cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				Until: cursorio.TextOffset{Byte: 6, LineColumn: cursorio.TextLineColumn{0, 6}},
			},
			TagNameOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
				Until: cursorio.TextOffset{Byte: 5, LineColumn: cursorio.TextLineColumn{0, 5}},
			},
			EndTagTokenOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 82, LineColumn: cursorio.TextLineColumn{0, 82}},
				Until: cursorio.TextOffset{Byte: 89, LineColumn: cursorio.TextLineColumn{0, 89}},
			},
		},
		nil, // injected <head />
		{ // body
			TokenOffsets: cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 6, LineColumn: cursorio.TextLineColumn{0, 6}},
				Until: cursorio.TextOffset{Byte: 12, LineColumn: cursorio.TextLineColumn{0, 12}},
			},
			TagNameOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 7, LineColumn: cursorio.TextLineColumn{0, 7}},
				Until: cursorio.TextOffset{Byte: 11, LineColumn: cursorio.TextLineColumn{0, 11}},
			},
			EndTagTokenOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 75, LineColumn: cursorio.TextLineColumn{0, 75}},
				Until: cursorio.TextOffset{Byte: 82, LineColumn: cursorio.TextLineColumn{0, 82}},
			},
		},
		{ // p
			TokenOffsets: cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 12, LineColumn: cursorio.TextLineColumn{0, 12}},
				Until: cursorio.TextOffset{Byte: 15, LineColumn: cursorio.TextLineColumn{0, 15}},
			},
			TagNameOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 13, LineColumn: cursorio.TextLineColumn{0, 13}},
				Until: cursorio.TextOffset{Byte: 14, LineColumn: cursorio.TextLineColumn{0, 14}},
			},
			// interrupted; end tag inferred based on effective last child
			EndTagTokenOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 31, LineColumn: cursorio.TextLineColumn{0, 31}},
				Until: cursorio.TextOffset{Byte: 31, LineColumn: cursorio.TextLineColumn{0, 31}},
			},
		},
		{ // custom-element
			TokenOffsets: cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 15, LineColumn: cursorio.TextLineColumn{0, 15}},
				Until: cursorio.TextOffset{Byte: 31, LineColumn: cursorio.TextLineColumn{0, 31}},
			},
			TagNameOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 16, LineColumn: cursorio.TextLineColumn{0, 16}},
				Until: cursorio.TextOffset{Byte: 30, LineColumn: cursorio.TextLineColumn{0, 30}},
			},
			// tokenizer closed custom-element early
			EndTagTokenOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 31, LineColumn: cursorio.TextLineColumn{0, 31}},
				Until: cursorio.TextOffset{Byte: 31, LineColumn: cursorio.TextLineColumn{0, 31}},
			},
		},
		{ // ul
			TokenOffsets: cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 31, LineColumn: cursorio.TextLineColumn{0, 31}},
				Until: cursorio.TextOffset{Byte: 35, LineColumn: cursorio.TextLineColumn{0, 35}},
			},
			TagNameOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 32, LineColumn: cursorio.TextLineColumn{0, 32}},
				Until: cursorio.TextOffset{Byte: 34, LineColumn: cursorio.TextLineColumn{0, 34}},
			},
			EndTagTokenOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 49, LineColumn: cursorio.TextLineColumn{0, 49}},
				Until: cursorio.TextOffset{Byte: 54, LineColumn: cursorio.TextLineColumn{0, 54}},
			},
		},
		{ // li
			TokenOffsets: cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 35, LineColumn: cursorio.TextLineColumn{0, 35}},
				Until: cursorio.TextOffset{Byte: 39, LineColumn: cursorio.TextLineColumn{0, 39}},
			},
			TagNameOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 36, LineColumn: cursorio.TextLineColumn{0, 36}},
				Until: cursorio.TextOffset{Byte: 38, LineColumn: cursorio.TextLineColumn{0, 38}},
			},
			EndTagTokenOffsets: &cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 44, LineColumn: cursorio.TextLineColumn{0, 44}},
				Until: cursorio.TextOffset{Byte: 49, LineColumn: cursorio.TextLineColumn{0, 49}},
			},
		},
		{ // text
			TokenOffsets: cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 39, LineColumn: cursorio.TextLineColumn{0, 39}},
				Until: cursorio.TextOffset{Byte: 44, LineColumn: cursorio.TextLineColumn{0, 44}},
			},
		},
		nil, // p (interrupted)
	}

	visitNode(document, func(n *html.Node) {
		expectedNodeMetadataListIdx++
		expectedNodeMetadata := expectedNodeMetadataList[expectedNodeMetadataListIdx]

		np, _ := documentOffsets.GetNodeMetadata(n)

		if np == expectedNodeMetadata {
			return
		} else if np == nil {
			t.Fatalf("%s: unexpected nil", dumpTraversal(n))
		} else if expectedNodeMetadata == nil {
			t.Fatalf("%s: unexpected nil", dumpTraversal(n))
		}

		if _a, _e := np.TokenOffsets.String(), expectedNodeMetadata.TokenOffsets.String(); _a != _e {
			t.Fatalf("%s: TokenOffsets: expected %v, got %v", dumpTraversal(n), _e, _a)
		}

		if np.TagNameOffsets != nil || expectedNodeMetadata.TagNameOffsets != nil {
			if np.TagNameOffsets == nil {
				t.Fatalf("%s: TagNameOffsets: expected %v, got nil", dumpTraversal(n), expectedNodeMetadata.TagNameOffsets.String())
			} else if expectedNodeMetadata.TagNameOffsets == nil {
				t.Fatalf("%s: TagNameOffsets: expected nil, got %v", dumpTraversal(n), np.TagNameOffsets.String())
			} else if _a, _e := np.TagNameOffsets.String(), expectedNodeMetadata.TagNameOffsets.String(); _a != _e {
				t.Fatalf("%s: TagNameOffsets: expected %v, got %v", dumpTraversal(n), _e, _a)
			}
		}

		if np.EndTagTokenOffsets != nil || expectedNodeMetadata.EndTagTokenOffsets != nil {
			if np.EndTagTokenOffsets == nil {
				t.Fatalf("%s: EndTagTokenOffsets: expected nil, got %v", dumpTraversal(n), np.EndTagTokenOffsets.String())
			} else if expectedNodeMetadata.EndTagTokenOffsets == nil {
				t.Fatalf("%s: EndTagTokenOffsets: expected nil, got %v", dumpTraversal(n), np.EndTagTokenOffsets.String())
			} else if _a, _e := np.EndTagTokenOffsets.String(), expectedNodeMetadata.EndTagTokenOffsets.String(); _a != _e {
				t.Fatalf("%s: EndTagTokenOffsets: expected %v, got %v", dumpTraversal(n), _e, _a)
			}
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><p><custom-element></custom-element></p><ul><li>hello</li></ul><p></p></body></html>"; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderLessTrivialBug(t *testing.T) {
	document, _, err := Parse(strings.NewReader(`<address itemscope itemtype="http://microformats.org/profile/hcard">
 <strong itemprop="fn"><span itemprop="n" itemscope><span itemprop="given-name">Alfred</span>
 <span itemprop="family-name">Person</span></span></strong> <br>
 <span itemprop="adr" itemscope>
  <span itemprop="street-address">1600 Amphitheatre Parkway</span> <br>
  <span itemprop="street-address">Building 43, Second Floor</span> <br>
  <span itemprop="locality">Mountain View</span>,
   <span itemprop="region">CA</span> <span itemprop="postal-code">94043</span>
 </span>
</address>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// previous implementation formatted lookup keys with alpha[numeric] characters which broke boundaries

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), `<html><head></head><body><address itemscope="" itemtype="http://microformats.org/profile/hcard">
 <strong itemprop="fn"><span itemprop="n" itemscope=""><span itemprop="given-name">Alfred</span>
 <span itemprop="family-name">Person</span></span></strong> <br/>
 <span itemprop="adr" itemscope="">
  <span itemprop="street-address">1600 Amphitheatre Parkway</span> <br/>
  <span itemprop="street-address">Building 43, Second Floor</span> <br/>
  <span itemprop="locality">Mountain View</span>,
   <span itemprop="region">CA</span> <span itemprop="postal-code">94043</span>
 </span>
</address></body></html>`; _a != _e {
		t.Errorf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestShortMalformedComments(t *testing.T) {
	// Regression test for slice bounds out of range error with short/malformed comments
	testCases := []struct {
		name           string
		input          string
		expected       string
		expectedOffset cursorio.TextOffsetRange
	}{
		{
			"empty comment",
			"<!---->",
			"<!----><html><head></head><body></body></html>",
			cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				Until: cursorio.TextOffset{Byte: 7, LineColumn: cursorio.TextLineColumn{0, 7}},
			},
		},
		{
			"short malformed comment 1",
			"<!-->",
			"<!--&gt;--><html><head></head><body></body></html>",
			cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				Until: cursorio.TextOffset{Byte: 5, LineColumn: cursorio.TextLineColumn{0, 5}},
			},
		},
		{
			"short malformed comment 2",
			"<!--->",
			"<!---&gt;--><html><head></head><body></body></html>",
			cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				Until: cursorio.TextOffset{Byte: 6, LineColumn: cursorio.TextLineColumn{0, 6}},
			},
		},
		{
			"comment with content",
			"<!-- comment -->",
			"<!-- comment --><html><head></head><body></body></html>",
			cursorio.TextOffsetRange{
				From:  cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
				Until: cursorio.TextOffset{Byte: 16, LineColumn: cursorio.TextLineColumn{0, 16}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			document, documentOffsets, err := Parse(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.input, err)
			}

			var foundComment bool

			visitNode(document, func(n *html.Node) {
				if n.Type != html.CommentNode {
					return
				}

				foundComment = true
				np, ok := documentOffsets.GetNodeMetadata(n)
				if !ok {
					t.Fatal("expected metadata for comment node")
				}

				if _a, _e := np.TokenOffsets, tc.expectedOffset; _a != _e {
					t.Errorf("token offsets: expected %v, got %v", _e, _a)
				}
			})

			if !foundComment {
				t.Fatal("expected to find a comment node")
			}

			var rendered = &bytes.Buffer{}

			if err = html.Render(rendered, document); err != nil {
				t.Fatalf("unexpected render error for %q: %v", tc.input, err)
			} else if _a, _e := rendered.String(), tc.expected; _a != _e {
				t.Fatalf("rendered: expected %q, got %q", _e, _a)
			}
		})
	}
}

func TestParserInjectedElementWithAttributes(t *testing.T) {
	// Regression test: HTML parser injects/restarts elements when encountering malformed markup
	// Previously this caused a panic because the injected element had attributes but was
	// not tracked in nodeTagByKey, resulting in a nil value being stored in metadataByNode.
	// With improved regex patterns, we now correctly track the malformed tag's offsets.
	document, documentOffsets, err := Parse(strings.NewReader(`<html><body><div>content</div><body</div></body></html>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we can traverse the tree without panicking
	var foundDivCount, foundBodyCount, foundBodyMalformedCount int
	visitNode(document, func(n *html.Node) {
		// Getting metadata should not panic even for injected elements
		_, _ = documentOffsets.GetNodeMetadata(n)

		if n.Type == html.ElementNode {
			switch n.DataAtom {
			case atom.Html:
				np, ok := documentOffsets.GetNodeMetadata(n)
				if !ok {
					t.Fatal("expected metadata for html element")
				}
				if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
					From: cursorio.TextOffset{
						Byte:       0,
						LineColumn: cursorio.TextLineColumn{0, 0},
					},
					Until: cursorio.TextOffset{
						Byte:       6,
						LineColumn: cursorio.TextLineColumn{0, 6},
					},
				}); _a != _e {
					t.Fatalf("html TokenOffsets: expected %v, got %v", _e, _a)
				}

			case atom.Div:
				foundDivCount++
				np, ok := documentOffsets.GetNodeMetadata(n)
				if !ok {
					t.Fatal("expected metadata for div element")
				}
				if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
					From: cursorio.TextOffset{
						Byte:       12,
						LineColumn: cursorio.TextLineColumn{0, 12},
					},
					Until: cursorio.TextOffset{
						Byte:       17,
						LineColumn: cursorio.TextLineColumn{0, 17},
					},
				}); _a != _e {
					t.Fatalf("div TokenOffsets: expected %v, got %v", _e, _a)
				}

			case atom.Body:
				foundBodyCount++
				// Well-formed body element
				np, ok := documentOffsets.GetNodeMetadata(n)
				if !ok {
					t.Fatal("expected metadata for body element")
				}
				if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
					From: cursorio.TextOffset{
						Byte:       6,
						LineColumn: cursorio.TextLineColumn{0, 6},
					},
					Until: cursorio.TextOffset{
						Byte:       12,
						LineColumn: cursorio.TextLineColumn{0, 12},
					},
				}); _a != _e {
					t.Fatalf("body TokenOffsets: expected %v, got %v", _e, _a)
				}

			default:
				// Check for malformed body< element (parsed as element with name "body<")
				if n.Data == "body<" {
					foundBodyMalformedCount++
					// Malformed <body< tag - now properly tracked with improved regex
					np, ok := documentOffsets.GetNodeMetadata(n)
					if !ok {
						t.Fatal("expected metadata for malformed body< element")
					}
					if _a, _e := np.TokenOffsets, (cursorio.TextOffsetRange{
						From: cursorio.TextOffset{
							Byte:       30,
							LineColumn: cursorio.TextLineColumn{0, 30},
						},
						Until: cursorio.TextOffset{
							Byte:       41,
							LineColumn: cursorio.TextLineColumn{0, 41},
						},
					}); _a != _e {
						t.Fatalf("body< TokenOffsets: expected %v, got %v", _e, _a)
					}
				}
			}
		}
	})

	// The malformed <body tag causes parser to handle things strangely, but should not panic
	if foundDivCount != 1 {
		t.Fatalf("expected exactly 1 div element, got %d", foundDivCount)
	} else if foundBodyCount != 1 {
		t.Fatalf("expected exactly 1 well-formed body element, got %d", foundBodyCount)
	} else if foundBodyMalformedCount != 1 {
		t.Fatalf("expected exactly 1 malformed body< element, got %d", foundBodyMalformedCount)
	}

	// Most importantly, verify rendering doesn't panic
	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check the full rendered output matches expected
	// The malformed <body tag results in spurious attributes from HTML parser error recovery
	expectedOutput := `<html><head></head><body><div>content</div><body< div=""></body<></body></html>`
	if _a, _e := rendered.String(), expectedOutput; _a != _e {
		t.Fatalf("rendered: expected %q, got %q", _e, _a)
	}
}

func TestReaderTagAttrNoSpaceBetween(t *testing.T) {
	// Regression test: attributes without spaces between them (malformed HTML)
	// E.g., class="x"href="y" instead of class="x" href="y"
	// This was causing panic: index out of range [N] with length N
	// because TagAttr slice length didn't match node.Attr length
	document, documentOffsets, err := Parse(strings.NewReader(`<html><body><a class="x"href="y"data="z">link</a></body></html>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.A {
			return
		}

		// Must not panic when accessing attributes
		if len(n.Attr) != 3 {
			t.Fatalf("expected 3 attributes, got %d", len(n.Attr))
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		}

		// TagAttr length must always match Attr length, even if some entries are nil
		if _a, _e := len(np.TagAttr), 3; _a != _e {
			t.Fatalf("TagAttr length (%d) must match Attr length (%d)", _a, _e)
		}

		// Attribute 0: class="x"
		if np.TagAttr[0] == nil {
			t.Fatal("attribute 0 (class) has nil metadata")
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
		}); _a != _e {
			t.Fatalf("attr 0 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Fatalf("attr 0 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       21,
				LineColumn: cursorio.TextLineColumn{0, 21},
			},
			Until: cursorio.TextOffset{
				Byte:       24,
				LineColumn: cursorio.TextLineColumn{0, 24},
			},
		}); *_a != _e {
			t.Fatalf("attr 0 value: expected %v, got %v", _e, _a)
		}

		// Attribute 1: href="y"
		if np.TagAttr[1] == nil {
			t.Fatal("attribute 1 (href) has nil metadata")
		} else if _a, _e := np.TagAttr[1].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       24,
				LineColumn: cursorio.TextLineColumn{0, 24},
			},
			Until: cursorio.TextOffset{
				Byte:       28,
				LineColumn: cursorio.TextLineColumn{0, 28},
			},
		}); _a != _e {
			t.Fatalf("attr 1 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[1].ValueOffsets == nil {
			t.Fatalf("attr 1 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[1].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       29,
				LineColumn: cursorio.TextLineColumn{0, 29},
			},
			Until: cursorio.TextOffset{
				Byte:       32,
				LineColumn: cursorio.TextLineColumn{0, 32},
			},
		}); *_a != _e {
			t.Fatalf("attr 1 value: expected %v, got %v", _e, _a)
		}

		// Attribute 2: data="z"
		if np.TagAttr[2] == nil {
			t.Fatal("attribute 2 (data) has nil metadata")
		} else if _a, _e := np.TagAttr[2].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       32,
				LineColumn: cursorio.TextLineColumn{0, 32},
			},
			Until: cursorio.TextOffset{
				Byte:       36,
				LineColumn: cursorio.TextLineColumn{0, 36},
			},
		}); _a != _e {
			t.Fatalf("attr 2 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[2].ValueOffsets == nil {
			t.Fatalf("attr 2 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[2].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       37,
				LineColumn: cursorio.TextLineColumn{0, 37},
			},
			Until: cursorio.TextOffset{
				Byte:       40,
				LineColumn: cursorio.TextLineColumn{0, 40},
			},
		}); *_a != _e {
			t.Fatalf("attr 2 value: expected %v, got %v", _e, _a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><a class=\"x\" href=\"y\" data=\"z\">link</a></body></html>"; _a != _e {
		t.Fatalf("rendered: expected %v, got %v", _e, _a)
	}
}

func TestReaderTagAttrUnquotedFollowedByAttribute(t *testing.T) {
	// Regression test: unquoted attribute value followed by another attribute with minimal spacing
	// E.g., href=https://example.com/path as=script
	// This was causing "regex attr failed" errors when the unquoted value wasn't properly consumed
	document, documentOffsets, err := Parse(strings.NewReader(`<html><head><link rel=preload href=https://example.com/script.js as=script></head></html>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.Link {
			return
		}

		// Verify all attributes are present
		if len(n.Attr) != 3 {
			t.Fatalf("expected 3 attributes, got %d: %+v", len(n.Attr), n.Attr)
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		}

		// TagAttr length must match Attr length
		if _a, _e := len(np.TagAttr), 3; _a != _e {
			t.Fatalf("TagAttr length (%d) must match Attr length (%d)", _a, _e)
		}

		// Attribute 0: rel=preload
		if np.TagAttr[0] == nil {
			t.Fatal("attribute 0 (rel) has nil metadata")
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       18,
				LineColumn: cursorio.TextLineColumn{0, 18},
			},
			Until: cursorio.TextOffset{
				Byte:       21,
				LineColumn: cursorio.TextLineColumn{0, 21},
			},
		}); _a != _e {
			t.Fatalf("attr 0 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Fatalf("attr 0 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       22,
				LineColumn: cursorio.TextLineColumn{0, 22},
			},
			Until: cursorio.TextOffset{
				Byte:       29,
				LineColumn: cursorio.TextLineColumn{0, 29},
			},
		}); *_a != _e {
			t.Fatalf("attr 0 value: expected %v, got %v", _e, *_a)
		}

		// Attribute 1: href=https://example.com/script.js
		if np.TagAttr[1] == nil {
			t.Fatal("attribute 1 (href) has nil metadata")
		} else if _a, _e := np.TagAttr[1].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       30,
				LineColumn: cursorio.TextLineColumn{0, 30},
			},
			Until: cursorio.TextOffset{
				Byte:       34,
				LineColumn: cursorio.TextLineColumn{0, 34},
			},
		}); _a != _e {
			t.Fatalf("attr 1 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[1].ValueOffsets == nil {
			t.Fatalf("attr 1 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[1].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       35,
				LineColumn: cursorio.TextLineColumn{0, 35},
			},
			Until: cursorio.TextOffset{
				Byte:       64,
				LineColumn: cursorio.TextLineColumn{0, 64},
			},
		}); *_a != _e {
			t.Fatalf("attr 1 value: expected %v, got %v", _e, *_a)
		}

		// Attribute 2: as=script
		if np.TagAttr[2] == nil {
			t.Fatal("attribute 2 (as) has nil metadata")
		} else if _a, _e := np.TagAttr[2].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       65,
				LineColumn: cursorio.TextLineColumn{0, 65},
			},
			Until: cursorio.TextOffset{
				Byte:       67,
				LineColumn: cursorio.TextLineColumn{0, 67},
			},
		}); _a != _e {
			t.Fatalf("attr 2 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[2].ValueOffsets == nil {
			t.Fatalf("attr 2 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[2].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       68,
				LineColumn: cursorio.TextLineColumn{0, 68},
			},
			Until: cursorio.TextOffset{
				Byte:       74,
				LineColumn: cursorio.TextLineColumn{0, 74},
			},
		}); *_a != _e {
			t.Fatalf("attr 2 value: expected %v, got %v", _e, *_a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head><link rel=\"preload\" href=\"https://example.com/script.js\" as=\"script\"/></head><body></body></html>"; _a != _e {
		t.Fatalf("rendered: expected %q, got %q", _e, _a)
	}
}

func TestReaderTagAttrUnquotedStartingWithSlash(t *testing.T) {
	// Regression test: unquoted attribute value starting with / (like href=/path/)
	// Previously the condition `rawCutset[0] != '/'` prevented these from being processed
	document, documentOffsets, err := Parse(strings.NewReader(`<html><body><a href=/informationen/ target=_blank>link</a></body></html>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.A {
			return
		}

		// Verify all attributes are present
		if len(n.Attr) != 2 {
			t.Fatalf("expected 2 attributes, got %d: %+v", len(n.Attr), n.Attr)
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		}

		// TagAttr length must match Attr length
		if _a, _e := len(np.TagAttr), 2; _a != _e {
			t.Fatalf("TagAttr length (%d) must match Attr length (%d)", _a, _e)
		}

		// Attribute 0: href=/informationen/
		if np.TagAttr[0] == nil {
			t.Fatal("attribute 0 (href) has nil metadata")
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       15,
				LineColumn: cursorio.TextLineColumn{0, 15},
			},
			Until: cursorio.TextOffset{
				Byte:       19,
				LineColumn: cursorio.TextLineColumn{0, 19},
			},
		}); _a != _e {
			t.Fatalf("attr 0 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Fatalf("attr 0 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       20,
				LineColumn: cursorio.TextLineColumn{0, 20},
			},
			Until: cursorio.TextOffset{
				Byte:       35,
				LineColumn: cursorio.TextLineColumn{0, 35},
			},
		}); *_a != _e {
			t.Fatalf("attr 0 value: expected %v, got %v", _e, *_a)
		}

		// Attribute 1: target=_blank
		if np.TagAttr[1] == nil {
			t.Fatal("attribute 1 (target) has nil metadata")
		} else if _a, _e := np.TagAttr[1].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       36,
				LineColumn: cursorio.TextLineColumn{0, 36},
			},
			Until: cursorio.TextOffset{
				Byte:       42,
				LineColumn: cursorio.TextLineColumn{0, 42},
			},
		}); _a != _e {
			t.Fatalf("attr 1 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[1].ValueOffsets == nil {
			t.Fatalf("attr 1 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[1].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       43,
				LineColumn: cursorio.TextLineColumn{0, 43},
			},
			Until: cursorio.TextOffset{
				Byte:       49,
				LineColumn: cursorio.TextLineColumn{0, 49},
			},
		}); *_a != _e {
			t.Fatalf("attr 1 value: expected %v, got %v", _e, *_a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><a href=\"/informationen/\" target=\"_blank\">link</a></body></html>"; _a != _e {
		t.Fatalf("rendered: expected %q, got %q", _e, _a)
	}
}

func TestReaderTagAttrUnquotedSingleSlash(t *testing.T) {
	// Regression test: unquoted attribute value that is just "/" followed by ">"
	// E.g., href=/> where the value is "/" not "/>"
	// This was causing the internal o attribute to be incorrectly included in the attribute value
	document, documentOffsets, err := Parse(strings.NewReader(`<a href=/>hello</a>`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	visitNode(document, func(n *html.Node) {
		if n.DataAtom != atom.A {
			return
		}

		// Verify attribute is present
		if len(n.Attr) != 1 {
			t.Fatalf("expected 1 attribute, got %d: %+v", len(n.Attr), n.Attr)
		}

		// Verify the attribute value is just "/"
		if _a, _e := n.Attr[0].Key, "href"; _a != _e {
			t.Fatalf("attr key: expected %q, got %q", _e, _a)
		}
		if _a, _e := n.Attr[0].Val, "/"; _a != _e {
			t.Fatalf("attr value: expected %q, got %q", _e, _a)
		}

		np, ok := documentOffsets.GetNodeMetadata(n)
		if !ok {
			t.Fatal("expected metadata")
		}

		// TagAttr length must match Attr length
		if _a, _e := len(np.TagAttr), 1; _a != _e {
			t.Fatalf("TagAttr length (%d) must match Attr length (%d)", _a, _e)
		}

		// Attribute: href=/
		if np.TagAttr[0] == nil {
			t.Fatal("attribute 0 (href) has nil metadata")
		} else if _a, _e := np.TagAttr[0].KeyOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       3,
				LineColumn: cursorio.TextLineColumn{0, 3},
			},
			Until: cursorio.TextOffset{
				Byte:       7,
				LineColumn: cursorio.TextLineColumn{0, 7},
			},
		}); _a != _e {
			t.Fatalf("attr 0 key: expected %v, got %v", _e, _a)
		} else if np.TagAttr[0].ValueOffsets == nil {
			t.Fatalf("attr 0 value: expected non-nil, got nil")
		} else if _a, _e := np.TagAttr[0].ValueOffsets, (cursorio.TextOffsetRange{
			From: cursorio.TextOffset{
				Byte:       8,
				LineColumn: cursorio.TextLineColumn{0, 8},
			},
			Until: cursorio.TextOffset{
				Byte:       9,
				LineColumn: cursorio.TextLineColumn{0, 9},
			},
		}); *_a != _e {
			t.Fatalf("attr 0 value: expected %v, got %v", _e, *_a)
		}
	})

	var rendered = &bytes.Buffer{}
	err = html.Render(rendered, document)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else if _a, _e := rendered.String(), "<html><head></head><body><a href=\"/\">hello</a></body></html>"; _a != _e {
		t.Fatalf("rendered: expected %q, got %q", _e, _a)
	}
}

func TestReaderDropNUL(t *testing.T) {
	document, _, err := Parse(strings.NewReader("<html><head><body>l\x00r</html>"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// previous implementation did not match the upstream tokenizer which drops NUL

	if _a, _e := document.FirstChild.LastChild.FirstChild.Data, "lr"; _a != _e {
		t.Fatalf("expected text content to be 'lr', got %q", _a)
	}
}
