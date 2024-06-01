package parser

import (
	"strings"

	"github.com/Astemirdum/si/internal/ast"
)

func (m *Module) Transform(scope *ast.Scope) *ast.Module {
	module := &ast.Module{
		ModuleTypeDefs: []*ast.TypeDef{},
		LocalTypes:     []*ast.TypeDef{},
		Functions:      []*ast.Function{},
		Scope:          scope,
		Pos:            m.Pos,
	}

	for _, td := range m.TypeDefs {
		module.LocalTypes = append(module.LocalTypes, td.Transform(module))
	}

	for _, f := range m.Functions {
		module.Functions = append(module.Functions, f.Transform(module))
	}

	return module
}

func (t *TypeDef) Transform(scope ast.ScopeLike) *ast.TypeDef {
	typ := t.Type.Transform(scope)

	return &ast.TypeDef{
		Alias: t.Ident,
		Type:  typ,
	}
}

func (f *Function) Transform(scope ast.ScopeLike) *ast.Function {
	childScope := ast.NewScopeFromParent(scope)

	fn := &ast.Function{
		Name:        f.Declarator.Ident,
		Params:      []*ast.Variable{},
		Variadic:    f.Variadic,
		Body:        []ast.StatementLike{},
		OnlyDeclare: f.OnlyDeclare,

		Scope: childScope,
		Pos:   f.Pos,
	}

	fn.ReturnType = f.Declarator.Type.Transform(fn)

	if f.Params != nil {
		for _, param := range f.Params {
			fn.Params = append(fn.Params, &ast.Variable{
				Ident:   param.Ident,
				Type:    param.Type.Transform(fn),
				IsParam: true,
			})
		}
	}

	// in case of function declaration, we don't have body
	if f.Body != nil {
		fn.Body = append(fn.Body, f.Body.Transform(fn))
	}

	return fn
}

func (s *Stmt) Transform(scope ast.ScopeLike) []ast.StatementLike {
	switch {
	case s.DeclStmt != nil:
		return []ast.StatementLike{s.DeclStmt.Transform(scope)}
	case s.AssignStmt != nil:
		return []ast.StatementLike{s.AssignStmt.Transform(scope)}
	case s.ExprStmt != nil:
		return []ast.StatementLike{s.ExprStmt.Transform(scope)}
	case s.ReturnStmt != nil:
		return []ast.StatementLike{s.ReturnStmt.Transform(scope)}
	case s.CompoundStmt != nil:
		return []ast.StatementLike{s.CompoundStmt.Transform(scope)}
	case s.IfStmt != nil:
		return []ast.StatementLike{s.IfStmt.Transform(scope)}
	case s.WhileStmt != nil:
		return []ast.StatementLike{s.WhileStmt.Transform(scope)}
	case s.ForStmt != nil:
		return []ast.StatementLike{s.ForStmt.Transform(scope)}
	case s.ContinueStmt != nil:
		return []ast.StatementLike{s.ContinueStmt.Transform(scope)}
	case s.BreakStmt != nil:
		return []ast.StatementLike{s.BreakStmt.Transform(scope)}
	default:
		panic("unknown statement")
	}
}

func (a *ExprStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	return &ast.ExprStmt{
		Expr:  a.Expr.Transform(scope),
		Scope: scope,
		Pos:   a.Pos,
	}
}

func (a *DeclStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	ds := &ast.DeclStmt{
		Ident: a.Declarator.Ident,
		Type:  a.Declarator.Type.Transform(scope),
		Scope: scope,
		Pos:   a.Pos,
	}

	if a.Expr != nil {
		ds.Expr = a.Expr.Transform(scope)
	}

	return ds
}

func (a *AssignStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	return &ast.AssignStmt{
		Left:  a.Left.Transform(scope),
		Right: a.Right.Transform(scope),
		Scope: scope,
		Pos:   a.Pos,
	}
}

func (r *ReturnStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	var expr ast.ExpressionLike

	if r.Expr != nil {
		expr = r.Expr.Transform(scope)
	}

	return &ast.ReturnStmt{
		Expr:  expr,
		Scope: scope,
		Pos:   r.Pos,
	}
}

func (r *ContinueStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	return &ast.ContinueStmt{
		Scope: scope,
		Pos:   r.Pos,
	}
}

func (r *BreakStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	return &ast.BreakStmt{
		Scope: scope,
		Pos:   r.Pos,
	}
}

func (c *CompoundStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	childScope := ast.NewScopeFromParent(scope)

	block := &ast.Block{
		Stmts: []ast.StatementLike{},
		Scope: childScope,
		Pos:   c.Pos,
	}

	for _, stmt := range c.Stmts {
		stmt := stmt.Transform(block)
		block.Stmts = append(block.Stmts, stmt...)
	}

	return block
}

func (i *IfStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	var el []ast.StatementLike
	if i.Else != nil {
		el = i.Else.Transform(scope)
	}

	return &ast.IfStmt{
		Condition: i.Condition.Transform(scope),
		Then:      i.Then.Transform(scope),
		Else:      el,
		Scope:     scope,
		Pos:       i.Pos,
	}
}

func (w *WhileStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	return &ast.WhileStmt{
		Condition: w.Condition.Transform(scope),
		Body:      w.Body.Transform(scope),
		Scope:     scope,
		Pos:       w.Pos,
	}
}

func (f *ForStmt) Transform(scope ast.ScopeLike) ast.StatementLike {
	var post ast.StatementLike
	if f.AssignPost != nil {
		post = f.AssignPost.Transform(scope)
	} else if f.ExprPost != nil {
		post = f.ExprPost.Transform(scope)
	}

	return &ast.ForStmt{
		Init:      f.Init.Transform(scope),
		Condition: f.Condition.Transform(scope),
		Post:      post,
		Body:      f.Body.Transform(scope),
		Scope:     scope,
		Pos:       f.Pos,
	}
}

func (e *Expr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	return e.LogicalExpr.Transform(scope)
}

func (le *LogicalExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if le.Op == "" {
		return le.Left.Transform(scope)
	}

	return &ast.BinaryOp{
		Left:  le.Left.Transform(scope),
		Op:    le.Op,
		Right: le.Right.Transform(scope),
		Scope: scope,
		Pos:   le.Pos,
	}
}

func (io *InclusiveOrExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if io.Op == "" {
		return io.Left.Transform(scope)
	}

	return &ast.BinaryOp{
		Left:  io.Left.Transform(scope),
		Op:    io.Op,
		Right: io.Right.Transform(scope),
		Scope: scope,
		Pos:   io.Pos,
	}
}

func (ea *ExclusiveOrExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if ea.Op == "" {
		return ea.Left.Transform(scope)
	}

	return &ast.BinaryOp{
		Left:  ea.Left.Transform(scope),
		Op:    ea.Op,
		Right: ea.Right.Transform(scope),
		Scope: scope,
		Pos:   ea.Pos,
	}
}

func (ae *AndExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if ae.Op == "" {
		return ae.Left.Transform(scope)
	}

	return &ast.BinaryOp{
		Left:  ae.Left.Transform(scope),
		Op:    ae.Op,
		Right: ae.Right.Transform(scope),
		Scope: scope,
		Pos:   ae.Pos,
	}
}

func (ee *EqualityExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if ee.Op == "" {
		return ee.Left.Transform(scope)
	}

	return &ast.BinaryOp{
		Left:  ee.Left.Transform(scope),
		Op:    ee.Op,
		Right: ee.Right.Transform(scope),
		Scope: scope,
		Pos:   ee.Pos,
	}
}

func (ce *ComparisonExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if ce.Op == "" {
		return ce.Left.Transform(scope)
	}

	return &ast.BinaryOp{
		Left:  ce.Left.Transform(scope),
		Op:    ce.Op,
		Right: ce.Right.Transform(scope),
		Scope: scope,
		Pos:   ce.Pos,
	}
}

func (ae *ShiftExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	head := ae.Head.Transform(scope)

	for _, tail := range ae.Tail {
		head = &ast.BinaryOp{
			Left:  head,
			Op:    tail.Op,
			Right: tail.Expr.Transform(scope),
			Scope: scope,
			Pos:   tail.Pos,
		}
	}

	return head
}

func (ae *AddExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	head := ae.Head.Transform(scope)

	for _, tail := range ae.Tail {
		head = &ast.BinaryOp{
			Left:  head,
			Op:    tail.Op,
			Right: tail.Expr.Transform(scope),
			Scope: scope,
			Pos:   tail.Pos,
		}
	}

	return head
}

func (me *MulExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	head := me.Head.Transform(scope)

	for _, tail := range me.Tail {
		head = &ast.BinaryOp{
			Left:  head,
			Op:    tail.Op,
			Right: tail.Expr.Transform(scope),
			Scope: scope,
			Pos:   tail.Pos,
		}
	}

	return head
}

func (ce *CastingExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if ce.Type == nil {
		return ce.Expr.Transform(scope)
	}

	return &ast.CastingOp{
		Type:  ce.Type.Transform(scope),
		Expr:  ce.Expr.Transform(scope),
		Scope: scope,
		Pos:   ce.Pos,
	}
}

func (pe *PrefixExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if pe.Next != nil {
		return pe.Next.Transform(scope)
	}

	return &ast.UnaryOp{
		Op:    pe.Op,
		Expr:  pe.Expr.Transform(scope),
		Scope: scope,
		Pos:   pe.Pos,
	}
}

func (pe *PostfixExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	next := pe.Next.Transform(scope)
	return pe.Postfix.Transform(scope, next)
}

func (pet *PostfixExprTick) Transform(scope ast.ScopeLike, ae ast.ExpressionLike) ast.ExpressionLike {
	if pet.Op != "" {
		return &ast.UnaryOp{
			Op:        pet.Op,
			Expr:      pet.Expr.Transform(scope, ae),
			IsPostfix: true,
			Scope:     scope,
			Pos:       pet.Pos,
		}
	}

	return ae
}

func (ae *AccessorExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	head := ae.Head.Transform(scope)

	for _, tail := range ae.Tail {
		deref := false
		if tail.Op == "->" {
			deref = true
		}

		head = &ast.AccessorOp{
			Expr:        head,
			Field:       tail.Field,
			Dereference: deref,
			Scope:       scope,
			Pos:         tail.Pos,
		}
	}

	return head
}

func (ie *IndexExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	head := ie.Head.Transform(scope)

	for _, tail := range ie.Tail {
		head = &ast.IndexOp{
			Expr:      head,
			IndexExpr: tail.Index.Transform(scope),
			Scope:     scope,
			Pos:       tail.Pos,
		}
	}

	return head
}

func (ue *UnaryExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	switch {
	case ue.FnCallExpr != nil:
		return ue.FnCallExpr.Transform(scope)
	case ue.PrimaryExpr != nil:
		return ue.PrimaryExpr.Transform(scope)
	default:
		panic("unknown unary expression")
	}
}

func (soe *SizeOfExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	if soe.Head != nil {
		return soe.Head.Transform(scope)
	}

	var typ *ast.Type
	if soe.Type != nil {
		typ = soe.Type.Transform(scope)
	}

	var expr ast.ExpressionLike
	if soe.Expr != nil {
		expr = soe.Expr.Transform(scope)
	}

	return &ast.SizeOfOp{
		Type:  typ,
		Expr:  expr,
		Scope: scope,
		Pos:   soe.Pos,
	}
}

func (fce *FnCallExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	args := make([]ast.ExpressionLike, len(fce.Args))

	for i, arg := range fce.Args {
		args[i] = arg.Transform(scope)
	}

	return &ast.FnCallOp{
		Ident: fce.Ident,
		Args:  args,
		Scope: scope,
		Pos:   fce.Pos,
	}
}

func (pe *PrimaryExpr) Transform(scope ast.ScopeLike) ast.ExpressionLike {
	switch {
	case pe.Variable != "":
		if pe.Variable == "true" || pe.Variable == "false" {
			return &ast.ConstantBoolOp{
				Constant: pe.Variable,
				Scope:    scope,
				Pos:      pe.Pos,
			}
		}

		return &ast.LoadOp{
			Name:  pe.Variable,
			Scope: scope,
			Pos:   pe.Pos,
		}
	case pe.Null != "":
		return &ast.ConstantNullOp{
			Constant: pe.Null,
			Scope:    scope,
			Pos:      pe.Pos,
		}
	case pe.Char != "":
		return &ast.ConstantCharOp{
			Constant: pe.Char,
			Scope:    scope,
			Pos:      pe.Pos,
		}
	case pe.Number != "":
		return &ast.ConstantNumberOp{
			Sign:     pe.Sign,
			Constant: pe.Number,
			Scope:    scope,
			Pos:      pe.Pos,
		}
	case pe.String != nil:
		sb := strings.Builder{}
		for _, el := range pe.String.Parts {
			if strings.HasPrefix(el, "\\") {
				switch el[1:] {
				case "n":
					sb.WriteByte('\n')
				case "r":
					sb.WriteByte('\r')
				case "t":
					sb.WriteByte('\t')
				case `"`:
					sb.WriteByte('"')
				default:
					sb.WriteString(el)
				}
			} else {
				sb.WriteString(el)
			}
		}

		return &ast.ConstantStringOp{
			Constant: sb.String(),
			Scope:    scope,
			Pos:      pe.Pos,
		}
	case pe.Expr != nil:
		return pe.Expr.Transform(scope)
	default:
		panic("unknown primary expression")
	}
}

func (t *Type) Transform(scope ast.ScopeLike) *ast.Type {
	var typ *ast.Type

	if t.Basic != "" {
		typ = ast.NewTypeBasic(scope, t.Pos, ast.BasicType(t.Basic))
	} else if t.Struct != nil {
		fields := make([]*ast.StructField, 0, len(t.Struct.Fields))
		for _, f := range t.Struct.Fields {
			fields = append(fields, &ast.StructField{
				Ident: f.Ident,
				Type:  f.Type.Transform(scope),
			})
		}

		typ = ast.NewTypeStruct(scope, t.Pos, fields...)
	} else if t.Alias != "" {
		typ = ast.NewTypeAlias(scope, t.Pos, t.Alias)
	}

	for range t.Pointers {
		typ = typ.NewPointer()
	}

	for _, l := range t.Lengths {
		typ = typ.NewArray(l)
	}

	return typ
}
