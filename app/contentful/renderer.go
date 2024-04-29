package contentful

var BLOCKS = map[string]string{
  "DOCUMENT": "document",
  "PARAGRAPH": "paragraph",

  "HEADING_1": "heading-1",
  "HEADING_2": "heading-2",
  "HEADING_3": "heading-3",
  "HEADING_4": "heading-4",
  "HEADING_5": "heading-5",
  "HEADING_6": "heading-6",

  "OL_LIST": "ordered-list",
  "UL_LIST": "unordered-list",
  "LIST_ITEM": "list-item",

  "HR": "hr",
  "QUOTE": "blockquote",

  "EMBEDDED_ENTRY": "embedded-entry-block",
  "EMBEDDED_ASSET": "embedded-asset-block",
  "EMBEDDED_RESOURCE": "embedded-resource-block",

  "TABLE": "table",
  "TABLE_ROW": "table-row",
  "TABLE_CELL": "table-cell",
  "TABLE_HEADER_CELL": "table-header-cell",
}

var INLINES = map[string]string{
  "HYPERLINK": "hyperlink",
  "ENTRY_HYPERLINK": "entry-hyperlink",
  "ASSET_HYPERLINK": "asset-hyperlink",
  "RESOURCE_HYPERLINK": "resource-hyperlink",
  "EMBEDDED_ENTRY": "embedded-entry-inline",
  "EMBEDDED_RESOURCE": "embedded-resource-inline",
}

type NodeData map[string]interface{}

type Node struct {
  nodeType string;
  data NodeData;
}

type Block struct {
	*Node
	nodeType string
	content []interface{}
}
