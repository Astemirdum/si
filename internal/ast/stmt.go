package ast

import (
	"fmt"
	"strings"

	"github.com/Astemirdum/si/pkg"
)

type StatementLikeList []StatementLike

func (s StatementLikeList) String() []string {
	lines := make([]string, 0, len(s))
	for _, stmt := range s {
		lines = append(lines, stmt.String()...)
	}
	return lines
}

func (a *ExprStmt) String() []string {
	return []string{"expr " + a.Expr.String() + ";"}
}

func (a *ExprStmt) Generate() error {
	_, err := a.Expr.Value()

	if err != nil {
		return err
	}

	return nil
}

func (d *DeclStmt) String() []string {
	if d.Expr != nil {
		return []string{"decl " + d.Type.String() + " " + d.Ident + " = " + d.Expr.String() + ";"}
	}

	return []string{"decl " + d.Type.String() + " " + d.Ident + ";"}
}

func (d *DeclStmt) Generate() error {
	typ, err := d.Type.IRType()
	if err != nil {
		return err
	}

	ptr := d.Scope.BasicBlock().NewAlloca(typ)

	v := &Variable{
		Ident: d.Ident,
		Type:  d.Type,
		Ptr:   ptr,
		Pos:   d.Pos,
	}

	err = d.Scope.AddLocal(v)
	if err != nil {
		return pkg.WithPos(err, d.Scope.Current().File, d.Pos)
	}

	if d.Expr != nil {
		expr, err := d.Expr.Value()
		if err != nil {
			return err
		}

		if !expr.Type.Equals(d.Type) {
			return pkg.WithPos(fmt.Errorf("cannot assign %s to %s", expr.Type.String(), d.Type.String()), d.Scope.Current().File, d.Pos)
		}

		d.Scope.BasicBlock().NewStore(expr.Value, ptr)
	}

	return nil
}

func (a *AssignStmt) String() []string {
	return []string{"assign " + a.Left.String() + " = " + a.Right.String() + ";"}
}

func (a *AssignStmt) Generate() error {
	left, err := a.Left.Value()
	if err != nil {
		return err
	}

	right, err := a.Right.Value()
	if err != nil {
		return err
	}

	if !left.Type.Equals(right.Type) {
		return pkg.WithPos(fmt.Errorf("cannot assign %s to %s", right.Type.String(), left.Type.String()), a.Scope.Current().File, a.Pos)
	}

	if left.Ptr == nil {
		return pkg.WithPos(fmt.Errorf("cannot assign to non-variable"), a.Scope.Current().File, a.Pos)
	}

	a.Scope.BasicBlock().NewStore(right.Value, left.Ptr)

	return nil
}

func (r *ReturnStmt) String() []string {
	return []string{"return " + r.Expr.String() + ";"}
}

func (r *ReturnStmt) Generate() error {
	if r.Expr == nil {
		r.Scope.BasicBlock().NewRet(nil)
		return nil
	}

	val, err := r.Expr.Value()

	if err != nil {
		return err
	}

	r.Scope.BasicBlock().NewRet(val.Value)

	return nil
}

func (r *ContinueStmt) String() []string {
	return []string{"continue;"}
}

func (r *ContinueStmt) Generate() error {
	const prevBlocks = 5
	blocks := r.Scope.CurrentFunction().Ptr.Blocks
	if len(blocks) < prevBlocks {
		return fmt.Errorf("invalid break statement")
	}
	breakBlock := blocks[len(blocks)-prevBlocks]
	if !isLoop(breakBlock.Ident()) {
		return fmt.Errorf("invalid break statement in block")
	}
	fmt.Println(blocks, breakBlock)

	r.Scope.BasicBlock().NewBr(breakBlock)

	return nil
}

func isLoop(ident string) bool {
	return strings.HasPrefix(ident, "%for.loop") || strings.HasPrefix(ident, "%while.loop")
}

func (r *BreakStmt) String() []string {
	return []string{"break;"}
}

func (r *BreakStmt) Generate() error {
	const prevBlocks = 4
	blocks := r.Scope.CurrentFunction().Ptr.Blocks
	if len(blocks) < prevBlocks {
		return fmt.Errorf("invalid break statement")
	}
	breakBlock := blocks[len(blocks)-prevBlocks]
	if !isMerge(breakBlock.Ident()) {
		return fmt.Errorf("invalid break statement in block")
	}

	r.Scope.BasicBlock().NewBr(breakBlock)

	return nil
}

func isMerge(ident string) bool {
	return strings.HasPrefix(ident, "%for.merge") || strings.HasPrefix(ident, "%while.merge")
}

func (i *IfStmt) String() []string {
	lines := []string{"if (" + i.Condition.String() + ") then"}

	for _, stmt := range i.Then {
		thenLines := stmt.String()

		// if the statement is a block, don't indent it
		if _, ok := stmt.(*Block); !ok {
			i.Scope.Current().PrefixLines(thenLines)
		}

		lines = append(lines, thenLines...)
	}

	if len(i.Else) > 0 {
		lines = append(lines, "else")

		for _, stmt := range i.Else {
			elseLines := stmt.String()

			// if the statement is a block, don't indent it
			if _, ok := stmt.(*Block); !ok {
				i.Scope.Current().PrefixLines(elseLines)
			}

			lines = append(lines, elseLines...)
		}
	}

	return lines
}

func (i *IfStmt) Generate() error {
	expr, err := i.Condition.Value()
	if err != nil {
		return err
	}

	thenBlock := i.Scope.CurrentFunction().Ptr.NewBlock(i.Scope.CurrentModule().GenerateID("if.then"))
	elseBlock := i.Scope.CurrentFunction().Ptr.NewBlock(i.Scope.CurrentModule().GenerateID("if.else"))
	mergeBlock := i.Scope.CurrentFunction().Ptr.NewBlock(i.Scope.CurrentModule().GenerateID("if.merge"))

	if !expr.Type.IsBool() {
		return pkg.WithPos(fmt.Errorf("cannot use %s as condition", expr.Type.String()), i.Scope.Current().File, i.Pos)
	}

	// entry block
	i.Scope.BasicBlock().NewCondBr(expr.Value, thenBlock, elseBlock)

	// then block
	i.Scope.SetBasicBlock(thenBlock)
	for _, stmt := range i.Then {
		if err := stmt.Generate(); err != nil {
			return err
		}
	}
	// if the last statement in the then block doesn't terminate the block, add a branch to the merge block
	// this is necessary because all basic blocks must terminate
	if i.Scope.BasicBlock().Term == nil {
		i.Scope.BasicBlock().NewBr(mergeBlock)
	}

	// else block
	i.Scope.SetBasicBlock(elseBlock)
	for _, stmt := range i.Else {
		if err := stmt.Generate(); err != nil {
			return err
		}
	}

	if i.Scope.BasicBlock().Term == nil {
		i.Scope.BasicBlock().NewBr(mergeBlock)
	}

	// merge block
	i.Scope.SetBasicBlock(mergeBlock)

	return nil
}

func (w *WhileStmt) String() []string {
	lines := []string{"while (" + w.Condition.String() + ")"}

	for _, stmt := range w.Body {
		bodyLines := stmt.String()

		// if the statement is a block, don't indent it
		if _, ok := stmt.(*Block); !ok {
			w.Scope.Current().PrefixLines(bodyLines)
		}

		lines = append(lines, bodyLines...)
	}

	return lines
}

func (w *WhileStmt) Generate() error {
	entryBlock := w.Scope.CurrentFunction().Ptr.NewBlock(w.Scope.CurrentModule().GenerateID("while.entry"))
	loopBlock := w.Scope.CurrentFunction().Ptr.NewBlock(w.Scope.CurrentModule().GenerateID("while.loop"))
	mergeBlock := w.Scope.CurrentFunction().Ptr.NewBlock(w.Scope.CurrentModule().GenerateID("while.merge"))

	w.Scope.BasicBlock().NewBr(entryBlock)

	// entry block
	w.Scope.SetBasicBlock(entryBlock)
	expr, err := w.Condition.Value()
	if err != nil {
		return err
	}

	if !expr.Type.IsBool() {
		return pkg.WithPos(fmt.Errorf("cannot use %s as condition", expr.Type.String()), w.Scope.Current().File, w.Pos)
	}

	w.Scope.BasicBlock().NewCondBr(expr.Value, loopBlock, mergeBlock)

	// loop block
	w.Scope.SetBasicBlock(loopBlock)
	for _, stmt := range w.Body {
		if err := stmt.Generate(); err != nil {
			return err
		}
	}

	if w.Scope.BasicBlock().Term == nil {
		w.Scope.BasicBlock().NewBr(entryBlock)
	}

	// merge block
	w.Scope.SetBasicBlock(mergeBlock)

	return nil
}

func (f *ForStmt) String() []string {
	lines := []string{"for ("}

	if f.Init != nil {
		lines = append(lines, f.Init.String()...)
	}

	lines = append(lines, f.Condition.String()+";")

	if f.Post != nil {
		lines = append(lines, f.Post.String()...)
	}

	lines = append(lines, ")")

	for _, stmt := range f.Body {
		bodyLines := stmt.String()

		// if the statement is a block, don't indent it
		if _, ok := stmt.(*Block); !ok {
			f.Scope.Current().PrefixLines(bodyLines)
		}

		lines = append(lines, bodyLines...)
	}

	return lines
}

func (f *ForStmt) Generate() error {
	entryBlock := f.Scope.CurrentFunction().Ptr.NewBlock(f.Scope.CurrentModule().GenerateID("for.entry"))
	loopBlock := f.Scope.CurrentFunction().Ptr.NewBlock(f.Scope.CurrentModule().GenerateID("for.loop"))
	mergeBlock := f.Scope.CurrentFunction().Ptr.NewBlock(f.Scope.CurrentModule().GenerateID("for.merge"))

	// Generate init statement
	if f.Init != nil {
		if err := f.Init.Generate(); err != nil {
			return err
		}
	}

	// Entry block
	f.Scope.BasicBlock().NewBr(entryBlock)
	f.Scope.SetBasicBlock(entryBlock)

	expr, err := f.Condition.Value()
	if err != nil {
		return err
	}

	if !expr.Type.IsBool() {
		return pkg.WithPos(fmt.Errorf("cannot use %s as condition", expr.Type.String()), f.Scope.Current().File, f.Pos)
	}

	f.Scope.BasicBlock().NewCondBr(expr.Value, loopBlock, mergeBlock)

	// Loop block
	f.Scope.SetBasicBlock(loopBlock)
	for _, stmt := range f.Body {
		if err := stmt.Generate(); err != nil {
			return err
		}
	}

	// Generate post statement
	if f.Post != nil {
		if err := f.Post.Generate(); err != nil {
			return err
		}
	}

	if f.Scope.BasicBlock().Term == nil {
		f.Scope.BasicBlock().NewBr(entryBlock)
	}

	// Merge block
	f.Scope.SetBasicBlock(mergeBlock)

	return nil
}
