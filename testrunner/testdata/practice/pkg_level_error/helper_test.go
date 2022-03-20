package pov

// source: problem-specification repo - POV exercise

type TestTreeData struct {
	root     string
	children []*Tree
}

var testTrees = map[string]TestTreeData{
	"singleton": {
		root:     "x",
		children: nil,
	},
	"parent and one sibling": {
		root:     "parent",
		children: []*Tree{New("x"), New("sibling")},
	},
	"parent and kids": {
		root:     "parent",
		children: []*Tree{New("x", New("kid-0"), New("kid-1"))},
	},
}

var newValueChildrenTestTrees = []string{"singleton", "parent and one sibling", "parent and kids"}

func mkTestTree(treeName string) *Tree {
	treeData := testTrees[treeName]
	return New(treeData.root, treeData.children...)
}
