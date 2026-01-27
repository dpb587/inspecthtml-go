package main

import (
	"fmt"
	"os"

	"github.com/dpb587/inspecthtml-go/inspecthtml"
	"golang.org/x/net/html"
)

func main() {
	parsedNode, parsedMetadata, err := inspecthtml.Parse(os.Stdin)
	if err != nil {
		panic(err)
	}

	dumpNode(parsedMetadata, parsedNode, "")
}

func dumpNode(metadata *inspecthtml.ParseMetadata, node *html.Node, indent string) {
	nodeMetadata, hasNodeMetadata := metadata.GetNodeMetadata(node)

	switch node.Type {
	case html.CommentNode:
		if hasNodeMetadata {
			fmt.Fprintf(os.Stdout,
				"%s// CommentToken=%s\n",
				indent,
				nodeMetadata.TokenOffsets.OffsetRangeString(),
			)
		}

		fmt.Fprintf(os.Stdout, "%s<!--%s-->\n", indent, node.Data)
	case html.TextNode:
		if hasNodeMetadata {
			fmt.Fprintf(os.Stdout,
				"%s// TextToken=%s\n",
				indent,
				nodeMetadata.TokenOffsets.OffsetRangeString(),
			)
		}

		fmt.Fprintf(os.Stdout, "%s%s\n", indent, node.Data)
	case html.DoctypeNode:
		if hasNodeMetadata {
			fmt.Fprintf(os.Stdout,
				"%s// DoctypeNode=%s\n",
				indent,
				nodeMetadata.TokenOffsets.OffsetRangeString(),
			)
		}

		fmt.Fprintf(os.Stdout, "%s%s\n", indent, node.Data)
	case html.ElementNode:
		if hasNodeMetadata {
			fmt.Fprintf(os.Stdout,
				"%s// StartTagToken=%s OuterOffsets=%s",
				indent,
				nodeMetadata.TokenOffsets.OffsetRangeString(),
				nodeMetadata.GetOuterOffsets().OffsetRangeString(),
			)

			if inner := nodeMetadata.GetInnerOffsets(); inner != nil {
				fmt.Fprintf(os.Stdout, " InnerOffsets=%s", inner.OffsetRangeString())
			}

			if nodeMetadata.TagSelfClosing {
				fmt.Fprintf(os.Stdout, " SelfClosing")
			}

			fmt.Fprintf(os.Stdout, "\n")
		}

		fmt.Fprintf(os.Stdout, "%s<%s", indent, node.Data)

		if len(node.Attr) > 0 {
			for attrIdx, attr := range node.Attr {
				fmt.Fprintf(os.Stdout, "\n%s  // Attr", indent)

				if hasNodeMetadata {
					if attrMetadata := nodeMetadata.TagAttr[attrIdx]; attrMetadata != nil {
						fmt.Fprintf(os.Stdout, " KeyOffsets=%s", attrMetadata.KeyOffsets.OffsetRangeString())

						if attrMetadata.ValueOffsets != nil {
							fmt.Fprintf(os.Stdout, " ValueOffsets=%s", attrMetadata.ValueOffsets.OffsetRangeString())
						}
					}
				}

				fmt.Fprintf(os.Stdout, "\n%s  %s=%q", indent, attr.Key, attr.Val)
			}

			fmt.Fprintf(os.Stdout, "\n%s", indent)
		}

		fmt.Fprintf(os.Stdout, ">\n")
	}

	for childNode := node.FirstChild; childNode != nil; childNode = childNode.NextSibling {
		dumpNode(metadata, childNode, indent+"  ")
	}

	if node.Type == html.ElementNode {
		if hasNodeMetadata && nodeMetadata.EndTagTokenOffsets != nil {
			fmt.Fprintf(os.Stdout, "%s// EndTagToken=%s\n", indent, nodeMetadata.EndTagTokenOffsets.OffsetRangeString())
		}

		fmt.Fprintf(os.Stdout, "%s</%s>\n", indent, node.Data)
	}
}
