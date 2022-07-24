package wintool

import (
	"os/exec"
	"strings"
)

// ToolCmdBuilder is used to construct the resulting `exec.Cmd`.
type ToolCmdBuilder struct {
	ToolPath     string
	BinPaths     []string
	LibPaths     []string
	IncludePaths []string
}

// newToolCmdBuilder creates a new tool command builder from its VC tool paths.
func newToolCmdBuilder(vctool *VCToolPaths) *ToolCmdBuilder {
	return &ToolCmdBuilder{
		ToolPath:     vctool.ToolPath,
		BinPaths:     []string{vctool.BinPath, vctool.DyLibPath},
		LibPaths:     []string{vctool.LibPath},
		IncludePaths: []string{vctool.IncludePath},
	}
}

// ToCommand converts a tool cmd to an exec.Cmd
func (tc *ToolCmdBuilder) ToCommand() *exec.Cmd {
	cmd := exec.Command(tc.ToolPath)
	cmd.Env = append(cmd.Env, "PATH="+strings.Join(tc.BinPaths, ";"))
	cmd.Env = append(cmd.Env, "LIB="+strings.Join(tc.LibPaths, ";"))
	cmd.Env = append(cmd.Env, "INCLUDE="+strings.Join(tc.IncludePaths, ";"))
	return cmd
}
