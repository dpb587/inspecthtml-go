package inspecthtml

import "golang.org/x/net/html"

type ParseMetadata struct {
	metadataByNode map[*html.Node]*NodeMetadata
}

func (po *ParseMetadata) GetNodeMetadata(n *html.Node) (*NodeMetadata, bool) {
	v, ok := po.metadataByNode[n]

	return v, ok
}
