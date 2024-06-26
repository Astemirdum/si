package ast

import (
	"strings"

	"github.com/Astemirdum/si/pkg"

	"github.com/llir/llvm/ir"
)

type Scope struct {
	File   *pkg.File
	Level  int
	Parent ScopeLike
}

type ScopeLike interface {
	Current() *Scope
	AddModuleTypeDef(alias string, typ *Type)
	AddLocalType(alias string, typ *Type)
	AddGlobal(v *Variable)
	AddLocal(v *Variable) error
	FindTypeDefByAlias(alias string) *TypeDef
	FindTypeDefByType(typ *Type) *TypeDef
	FindVariable(ident string) *Variable
	FindFunction(ident string) *Function
	CurrentModule() *Module
	CurrentFunction() *Function
	BasicBlock() *ir.Block
	SetBasicBlock(b *ir.Block)
}

func NewScope(file *pkg.File) *Scope {
	return &Scope{
		File:  file,
		Level: 0,
	}
}

func NewScopeFromParent(parent ScopeLike) *Scope {
	return &Scope{
		File:   parent.Current().File,
		Level:  parent.Current().Level + 1,
		Parent: parent,
	}
}

func (s *Scope) PrefixLines(lines []string) {
	parentLevel := 0

	if s.Parent != nil {
		parentLevel = s.Parent.Current().Level
	}

	for i := range lines {
		lines[i] = strings.Repeat("  ", s.Level-parentLevel) + lines[i]
	}
}
