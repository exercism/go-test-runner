package extra_packages

// This file imports the dependencies we want to support on the Go track.
// To add or remove a supported dependency, add or remove an import here
// and then run 'go mod tidy' to update the go.mod and go.sum files.

// Note that some packages are part of a module, in which case the whole
// module will be downloaded as a dependency.
// e.g "golang.org/x/exp/constraints" is a package that is part of the
// "golang.org/x/exp" module. Importing "golang.org/x/exp/constraints"
// makes the whole "golang.org/x/exp/" module be downloaded and referenced
// in the go.mod file.
// This means that if you want to add a module as a dependency
// that is not itself a package, importing any of its sub-packages
// should suffice.

import (
	_ "golang.org/x/exp/constraints" // package of module golang.org/x/exp
	_ "golang.org/x/text"
)
