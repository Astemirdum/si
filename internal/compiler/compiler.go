package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Astemirdum/si/internal/ast"
	"github.com/Astemirdum/si/internal/parser"
	"github.com/Astemirdum/si/pkg"
	"github.com/llir/llvm/ir"
)

type Compiler struct {
	parser    *parser.Parser
	clangPath string
	tmpFolder string
}

func NewCompiler() *Compiler {
	c := new(Compiler)

	c.parser = parser.NewParser()

	path, err := exec.LookPath("clang")
	if err != nil {
		panic(err)
	}
	c.clangPath = path

	dir, err := os.MkdirTemp("", "konstruktor-c")
	if err != nil {
		panic(err)
	}
	c.tmpFolder = dir

	return c
}

func (c *Compiler) Destroy() {
	err := os.RemoveAll(c.tmpFolder)
	if err != nil {
		panic(err)
	}
}

func (c *Compiler) runClang(srcFilePath, binaryFilePath string) (string, error) {
	cmd := exec.Command(c.clangPath, "-Werror", "-Wno-override-module", "-o", binaryFilePath, srcFilePath) //nolint:gosec
	out, err := cmd.CombinedOutput()

	return string(out), err
}

func (c *Compiler) runBinary(binaryFilePath string) (string, error) {
	cmd := exec.Command(binaryFilePath)
	out, err := cmd.CombinedOutput()

	return string(out), err
}

func (c *Compiler) RunProgramSi(src string, opts ...Option) (string, error) {
	basename := OverrideBasename("main", opts...)

	input := pkg.NewFile(basename, src)
	scope := ast.NewScope(input)

	// parsing src to AST
	parsedAst, err := c.parser.ParseFile(input)
	if err != nil {
		return "", NewParseError(err, src)
	}
	transformedAst := parsedAst.Transform(scope)

	// AST -> generate LLVM
	bitCode, err := transformedAst.Generate()
	if err != nil {
		return "", NewGenerateError(err, src)
	}

	bitCodeStr := bitCode.String()

	//fmt.Println(bitCodeStr)

	srcFilePath := filepath.Join(c.tmpFolder, fmt.Sprintf("%s.ll", basename))

	err = os.WriteFile(srcFilePath, []byte(bitCodeStr), 0600)
	if err != nil {
		return "", err
	}

	binaryFilePath := filepath.Join(c.tmpFolder, fmt.Sprintf("%s.bin", basename))

	// compile llvm
	result, err := c.runClang(srcFilePath, binaryFilePath)
	if err != nil {
		return "", NewClangRunError(err, src, result)
	}

	// run bin
	result, err = c.runBinary(binaryFilePath)
	if err != nil {
		return "", NewBinaryRunError(err, src)
	}

	return result, nil
}

func (c *Compiler) RunProgramC(src string, opts ...Option) (string, error) {
	file := OverrideBasename("main", opts...)

	srcFilePath := filepath.Join(c.tmpFolder, fmt.Sprintf("%s.c", file))

	err := os.WriteFile(srcFilePath, []byte(src), 0600)
	if err != nil {
		return "", err
	}

	binaryFilePath := filepath.Join(c.tmpFolder, fmt.Sprintf("%s.bin", file))

	result, err := c.runClang(srcFilePath, binaryFilePath)
	if err != nil {
		return "", NewClangRunError(err, src, result)
	}

	result, err = c.runBinary(binaryFilePath)
	if err != nil {
		return "", NewBinaryRunError(err, src)
	}

	return result, nil
}

func (c *Compiler) RunProgramLL(m *ir.Module, opts []Option) (string, error) {
	file := OverrideBasename("main", opts...)

	buf := &strings.Builder{}
	_, err := m.WriteTo(buf)
	if err != nil {
		return "", err
	}

	src := buf.String()

	srcFilePath := filepath.Join(c.tmpFolder, fmt.Sprintf("%s.ll", file))

	err = os.WriteFile(srcFilePath, []byte(src), 0600)
	if err != nil {
		return "", err
	}

	binaryFilePath := filepath.Join(c.tmpFolder, fmt.Sprintf("%s.bin", file))

	result, err := c.runClang(srcFilePath, binaryFilePath)
	if err != nil {
		return "", NewClangRunError(err, src, result)
	}

	result, err = c.runBinary(binaryFilePath)

	if err != nil {
		return "", NewBinaryRunError(err, src)
	}

	return result, nil
}
