package compiler

import (
	"fmt"

	"github.com/Astemirdum/si/internal/ast"
	"github.com/Astemirdum/si/pkg"
	"github.com/llir/llvm/ir"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
}

func ExprToProgramSi(expr, format string, opts []Option) string {
	declares := JoinDeclares(opts)

	src := fmt.Sprintf(`
%s
i64 printf(i8 *fmt,... );

i64 main() {
	%s
	printf(%s);
	return 0;
}
`, declares, expr, format)

	return src
}

func ExprToProgramC(expr, format string, opts []Option) string {
	headers := JoinHeaders(opts)

	src := fmt.Sprintf(`
#include <stdio.h>
%s

int main() {
	%s
	printf(%s);
	return 0;
}
`, headers, expr, format)

	return src
}

func (suite *Suite) EqualProgramSi(src, expected string, opts ...Option) {
	c := NewCompiler()
	defer c.Destroy()

	result, err := c.RunProgramSi(src, opts...)
	suite.NoError(err)

	suite.Equal(expected, result)
}

func (suite *Suite) ErrorGenerateProgramSi(src, contains string, opts ...Option) {
	c := NewCompiler()
	defer c.Destroy()

	_, err := c.RunProgramSi(src, opts...)
	cre := &GenerateError{}
	suite.ErrorAs(err, &cre)
	suite.Contains(err.Error(), contains)
}

func (suite *Suite) EqualExprSi(expr, format, expected string, opts ...Option) {
	src := ExprToProgramSi(expr, format, opts)
	suite.EqualProgramSi(src, expected, opts...)
}

func (suite *Suite) ErrorGenerateExprSi(expr, contains string, opts ...Option) {
	src := ExprToProgramSi(expr, `""`, opts)
	suite.ErrorGenerateProgramSi(src, contains, opts...)
}

func (suite *Suite) EqualProgramC(src, expected string, opts ...Option) {
	c := NewCompiler()
	defer c.Destroy()

	result, err := c.RunProgramC(src, opts...)
	suite.NoError(err)

	suite.Equal(expected, result)
}

func (suite *Suite) ErrorClangProgramC(src, contains string, opts ...Option) {
	c := NewCompiler()
	defer c.Destroy()

	_, err := c.RunProgramC(src, opts...)
	cre := &ClangRunError{}
	suite.ErrorAs(err, &cre)
	suite.Contains(err.Error(), contains)
}

func (suite *Suite) EqualExprC(expr, format, expected string, opts ...Option) {
	src := ExprToProgramC(expr, format, opts)
	suite.EqualProgramC(src, expected, opts...)
}

func (suite *Suite) ErrorClangExprC(expr, contains string, opts ...Option) {
	src := ExprToProgramC(expr, `""`, opts)
	suite.ErrorClangProgramC(src, contains, opts...)
}

func (suite *Suite) EqualLL(m *ir.Module, expected string, opts ...Option) {
	c := NewCompiler()
	defer c.Destroy()

	result, err := c.RunProgramLL(m, opts)
	suite.NoError(err)

	suite.Equal(expected, result)
}

func (suite *Suite) GetLL(src string) (string, error) {
	c := NewCompiler()
	defer c.Destroy()

	basename := OverrideBasename("main")

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
	return bitCodeStr, nil
}
