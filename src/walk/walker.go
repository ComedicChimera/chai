package walk

import "chai/deps"

// Walker is the construct responsible for performing semantic analysis on files
// as both the top level and expression level
type Walker struct {
	// SrcFile is the file this walker is walking
	SrcFile *deps.ChaiFile
}

// NewWalker creates a new walker for a given file
func NewWalker(f *deps.ChaiFile) *Walker {
	return &Walker{
		SrcFile: f,
	}
}
