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
		}); _a != _e {
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
		}); _a != _e {
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
		}); _a != _e {
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
		}); _a != _e {
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
		}); _a != _e {
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
