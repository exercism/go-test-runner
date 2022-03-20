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
}

var newValueChildrenTestTrees = []string{"singleton"}

func mkTestTree(treeName string) *Tree {
	treeData := testTrees[treeName]
	return New(treeData.root, treeData.children...)
}
