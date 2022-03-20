package pov

import (
	"testing"
)

func TestNewNotNil(t *testing.T) {
	for _, treeName := range newValueChildrenTestTrees {
		t.Run(treeName+" not nil", func(t *testing.T) {
			tree := mkTestTree(treeName)
			if tree == nil {
				t.Fatalf("tree should not be nil: %v", treeName)
			}
		})
	}
}
