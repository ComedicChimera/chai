package wintool

import (
	"os"
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
	addDefaultEnv(cmd)
	return cmd
}

// addDefaultEnv adds the parent processes environment variables to the command.
func addDefaultEnv(cmd *exec.Cmd) {
envloop:
	for _, envv := range os.Environ() {
		envvContent := strings.Split(envv, "=")
		k := envvContent[0]

		for i, cenvv := range cmd.Env {
			cenvvContent := strings.Split(cenvv, "=")
			ck, cv := cenvvContent[0], cenvvContent[1]

			if strings.EqualFold(k, ck) {
				cmd.Env[i] += ";" + cv
				continue envloop
			}
		}

		cmd.Env = append(cmd.Env, envv)
	}
}
