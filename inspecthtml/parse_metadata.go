package inspecthtml

import (
	"github.com/dpb587/cursorio-go/cursorio"
	"golang.org/x/net/html"
)

type ParseMetadata struct {
	metadataByNode map[*html.Node]*NodeMetadata
}

func (po *ParseMetadata) GetNodeMetadata(n *html.Node) (*NodeMetadata, bool) {
	v, ok := po.metadataByNode[n]
	if !ok {
		return nil, false
	}

	if n.Type == html.ElementNode && !v.TagSelfClosing && v.EndTagTokenOffsets == nil {
		// handled on-demand assuming less frequently needed than re-traverse entire tree after rebuildNode
		if n.LastChild != nil {
			if po.metadataByNode[n.LastChild] != nil {
				v.EndTagTokenOffsets = &cursorio.TextOffsetRange{
					From:  po.metadataByNode[n.LastChild].TokenOffsets.Until,
					Until: po.metadataByNode[n.LastChild].TokenOffsets.Until,
				}
			}
		} else if n.NextSibling != nil {
			if po.metadataByNode[n.NextSibling] != nil {
				v.EndTagTokenOffsets = &cursorio.TextOffsetRange{
					From:  po.metadataByNode[n.NextSibling].TokenOffsets.From,
					Until: po.metadataByNode[n.NextSibling].TokenOffsets.From,
				}
			}
		} else if n.Parent != nil {
			if po.metadataByNode[n.Parent] != nil {
				v.EndTagTokenOffsets = &cursorio.TextOffsetRange{
					From:  po.metadataByNode[n.Parent].EndTagTokenOffsets.From,
					Until: po.metadataByNode[n.Parent].EndTagTokenOffsets.From,
				}
			}
		}
	}

	return v, ok
}
