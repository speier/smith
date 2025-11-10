package core

// Node represents an element in the UI tree
type Node struct {
	Type       string            // box, text, input, scrollarea
	ID         string            // element id
	Classes    []string          // css classes
	Attributes map[string]string // custom attributes
	Styles     *ComputedStyle    // computed styles
	Children   []*Node           // child nodes
	Content    string            // text content
	Parent     *Node             // parent node

	// Layout results
	X      int // computed x position
	Y      int // computed y position
	Width  int // computed width
	Height int // computed height
}

// ComputedStyle holds the final computed styles for a node
type ComputedStyle struct {
	// Layout
	Width    string // "100%", "50", "auto"
	Height   string // "100%", "10", "auto"
	Display  string // "block", "flex"
	FlexDir  string // "row", "column"
	Flex     string // "1", "0"
	Position string // "static", "fixed"
	Top      int
	Bottom   int
	Left     int
	Right    int

	// Spacing
	PaddingTop    int
	PaddingRight  int
	PaddingBottom int
	PaddingLeft   int
	MarginTop     int
	MarginRight   int
	MarginBottom  int
	MarginLeft    int

	// Visual
	Color      string // hex color
	BgColor    string // hex background color
	Border     bool   // has border
	BorderChar string // "single", "rounded", "double"
	TextAlign  string // "left", "center", "right"
}

// NewNode creates a new node with default styles
func NewNode(nodeType string) *Node {
	return &Node{
		Type:       nodeType,
		Attributes: make(map[string]string),
		Children:   []*Node{},
		Styles: &ComputedStyle{
			Width:      "auto",
			Height:     "auto",
			Display:    "block",
			FlexDir:    "column",
			Flex:       "0",
			Position:   "static",
			BorderChar: "single",
			TextAlign:  "left",
		},
	}
}

// AddChild adds a child node
func (n *Node) AddChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// FindByID finds a node by ID (depth-first search)
func (n *Node) FindByID(id string) *Node {
	if n.ID == id {
		return n
	}
	for _, child := range n.Children {
		if found := child.FindByID(id); found != nil {
			return found
		}
	}
	return nil
}
