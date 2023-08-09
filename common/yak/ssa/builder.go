package ssa

import (
	"fmt"
	"go/types"
	"strconv"
	"strings"

	yak "github.com/yaklang/yaklang/common/yak/antlr4yak/parser"
	"github.com/yaklang/yaklang/common/yak/antlr4yak/yakvm"
)

type builder struct {
	*Function
	next   *builder
	target *target // for break and continue
}

var (
	basicTypes = make(map[string]*types.Basic)
)

func init() {
	for _, basic := range types.Typ {
		basicTypes[basic.Name()] = basic
	}
}

// entry point
func (b *builder) build(ast *yak.YaklangParser) {
	// ast.StatementList()
	b.buildStatementList(ast.StatementList().(*yak.StatementListContext))
}

// statement list
func (b *builder) buildStatementList(stmtlist *yak.StatementListContext) {
	for _, stmt := range stmtlist.AllStatement() {
		if stmt, ok := stmt.(*yak.StatementContext); ok {
			b.buildStatement(stmt)
		}
	}
}

func (b *builder) buildStatement(stmt *yak.StatementContext) {
	//TODO: decalear Variable Expression

	// assign Expression
	if s, ok := stmt.AssignExpressionStmt().(*yak.AssignExpressionStmtContext); ok {
		b.buildAssignExpressionStmt(s)
		return
	}

	// expression
	if s, ok := stmt.ExpressionStmt().(*yak.ExpressionStmtContext); ok {
		b.buildExpressionStmt(s)
	}

	//TODO: Block
	//TODO: try Stmt

	//TODO: empty

	// if stmt
	if s, ok := stmt.IfStmt().(*yak.IfStmtContext); ok {
		b.buildIfStmt(s, nil)
	}

	//TODO: Switch stmt

	//TODO: for range stmt

	// for stmt
	if s, ok := stmt.ForStmt().(*yak.ForStmtContext); ok {
		b.buildForStmt(s)
	}

	// break stmt
	if _, ok := stmt.BreakStmt().(*yak.BreakStmtContext); ok {
		if _break := b.target._break; _break != nil {
			b.emitJump(_break)
		} else {
			panic("unexpection break stmt")
		}
	}
	// return stmt
	if s, ok := stmt.ReturnStmt().(*yak.ReturnStmtContext); ok {
		b.buildReturnStmt(s)
	}
	// continue stmt
	if _, ok := stmt.ContinueStmt().(*yak.ContinueStmtContext); ok {
		if _continue := b.target._continue; _continue != nil {
			b.emitJump(_continue)
		} else {
			panic("unexpection continue stmt")
		}
	}

	//TODO: fallthrough stmt
	//TODO: include stmt
	//TODO: defer stmt
	//TODO: go stmt
	//TODO: assert stmt

}

//TODO: try stmt

// expression stmt
func (b *builder) buildExpressionStmt(stmt *yak.ExpressionStmtContext) {
	if s, ok := stmt.Expression().(*yak.ExpressionContext); ok {
		b.buildExpression(s)
	}
}

// assign expression stmt
func (b *builder) buildAssignExpressionStmt(stmt *yak.AssignExpressionStmtContext) {
	s := stmt.AssignExpression()
	if s == nil {
		return
	}
	if i, ok := s.(*yak.AssignExpressionContext); ok {
		b.buildAssignExpression(i)
	}
}

// TODO: include stmt
// TODO: defer stmt
// TODO: go stmt
// return stmt
func (b *builder) buildReturnStmt(stmt *yak.ReturnStmtContext) {
	if list, ok := stmt.ExpressionList().(*yak.ExpressionListContext); ok {
		values := b.buildExpressionList(list)
		b.emitReturn(values)
	} else {
		b.emitReturn(nil)
	}
}

// for stmt
func (b *builder) buildForStmt(stmt *yak.ForStmtContext) {
	//	    ...enter...
	//	    // for first expre in here
	//      jump loop.header
	// loop.header:
	//      // for stmt cond in here
	//      If [cond] true -> loop.body, false -> loop.exit
	// loop.body:
	//      // for body block in here
	// loop.latch:                           (target of continue)
	//      // for third expr in here
	//      jump loop.header
	// loop.exit:				              (target of break)
	//      jump rest
	// rest:
	//      ...rest.code....

	// current := f.currentBlock
	enter := b.currentBlock
	header := b.newBasicBlockUnSealed("loop.header")

	body := b.newBasicBlock("loop.body")
	exit := b.newBasicBlock("loop.exit")
	latch := b.newBasicBlock("loop.latch")
	var endThird *yak.ForThirdExprContext
	endThird = nil

	var cond Value
	if e, ok := stmt.Expression().(*yak.ExpressionContext); ok {
		// if only expression; just build expression in header;
		cond = b.buildExpression(e)
	} else if condition, ok := stmt.ForStmtCond().(*yak.ForStmtCondContext); ok {
		if first, ok := condition.ForFirstExpr().(*yak.ForFirstExprContext); ok {
			// first expression is initialization, in enter block
			b.currentBlock = enter
			b.buildForFirstExpr(first)
		}
		if expr, ok := condition.Expression().(*yak.ExpressionContext); ok {
			// build expression in header
			b.currentBlock = header
			cond = b.buildExpression(expr)
		} else {
			// not found expression; default is true
			cond = NewConst(true)
		}

		if third, ok := condition.ForThirdExpr().(*yak.ForThirdExprContext); ok {
			// third exprssion in latch block, when loop.body builded
			endThird = third
		}
	}
	// jump enter->header
	b.currentBlock = enter
	b.emitJump(header)
	// build if in header end; to exit or body
	b.currentBlock = header
	ifssa := b.emitIf(cond)
	ifssa.AddFalse(exit)
	ifssa.AddTrue(body)

	//  build body
	b.currentBlock = body
	if block, ok := stmt.Block().(*yak.BlockContext); ok {

		b.target = &target{
			tail:      b.target, // push
			_break:    exit,
			_continue: latch,
		}

		b.buildBlock(block)      // block can create block
		b.target = b.target.tail // pop
		// // f.currentBlock is end block in body
		// body = b.currentBlock
	}
	// jump body -> latch
	b.emitJump(latch)

	// build latch
	b.currentBlock = latch
	if endThird != nil {
		// build third expression in loop.body end
		b.buildForThirdExpr(endThird)
	}
	// jump latch -> header
	b.emitJump(header)

	// now header sealed
	header.Sealed()

	rest := b.newBasicBlock("")
	// jump exit -> rest
	b.currentBlock = exit
	b.emitJump(rest)
	// continue in rest code
	b.currentBlock = rest
}

// for first expr
func (f *builder) buildForFirstExpr(state *yak.ForFirstExprContext) {
	if ae, ok := state.AssignExpression().(*yak.AssignExpressionContext); ok {
		f.buildAssignExpression(ae)
	}
	if e, ok := state.Expression().(*yak.ExpressionContext); ok {
		f.buildExpression(e)
	}
}

// for third expr
func (f *builder) buildForThirdExpr(state *yak.ForThirdExprContext) {
	if ae, ok := state.AssignExpression().(*yak.AssignExpressionContext); ok {
		f.buildAssignExpression(ae)
	}
	if e, ok := state.Expression().(*yak.ExpressionContext); ok {
		f.buildExpression(e)
	}
}

//TODO: for range stmt

// TODO: switch stmt

// if stmt
func (b *builder) buildIfStmt(state *yak.IfStmtContext, done *BasicBlock) {
	//	    ...enter...
	//      // if stmt cond in here
	//      If [cond] true -> if.true, false -> if.elif
	// if.true:
	//      // if-true-body block in here
	//      jump if.done
	// if.elif:
	//      // if-elif cond in here         (this build in "elif" and "else if")
	//      If [cond] true -> if.elif_true, false -> if.false
	// if.elif_true:
	//      // if-elif-true-body block in here
	//      jump if.done
	// if.false:
	//      // if-elif-false-body block in here
	//      jump if.done
	// if.done:				              (target of all if block)
	//      jump rest
	// rest:
	//      ...rest.code....

	// condition
	cond := b.buildExpression(state.Expression(0).(*yak.ExpressionContext))
	// if instruction
	ifssa := b.emitIf(cond)
	isOutIf := false
	if done == nil {
		done = b.newBasicBlock("if.done")
		isOutIf = true
	}

	// create true block
	trueBlock := b.newBasicBlock("if.true")
	ifssa.AddTrue(trueBlock)

	// build true block
	b.currentBlock = trueBlock
	b.buildBlock(state.Block(0).(*yak.BlockContext))
	b.emitJump(done)

	// handler "elif"
	previf := ifssa
	// add elif block to prev-if false
	for index := range state.AllElif() {
		// create false block
		if previf.False == nil {
			previf.AddFalse(b.newBasicBlock("if.elif"))
		}
		// in false block
		b.currentBlock = previf.False
		// build condition
		cond := b.buildExpression(state.Expression(index + 1).(*yak.ExpressionContext))
		// if instruction
		currentif := b.emitIf(cond)
		// create true block
		trueBlock := b.newBasicBlock("if.true")
		currentif.AddTrue(trueBlock)
		// build true block
		b.currentBlock = trueBlock
		b.buildBlock(state.Block(index + 1).(*yak.BlockContext))
		// jump to done
		b.emitJump(done)
		// for next elif
		previf = currentif
	}

	// hanlder "else" and "else if "
	if elseStmt, ok := state.ElseBlock().(*yak.ElseBlockContext); ok {
		if elseblock, ok := elseStmt.Block().(*yak.BlockContext); ok {
			// "else"
			// create false block
			falseBlock := b.newBasicBlock("if.false")
			previf.AddFalse(falseBlock)

			// build false block
			b.currentBlock = falseBlock
			b.buildBlock(elseblock)
			b.emitJump(done)
		} else if elifstmt, ok := elseStmt.IfStmt().(*yak.IfStmtContext); ok {
			// "else if"
			// create elif block
			elifBlock := b.newBasicBlock("if.elif")
			previf.AddFalse(elifBlock)

			// build elif block
			b.currentBlock = elifBlock
			b.buildIfStmt(elifstmt, done)
		}
	} else {
		previf.AddFalse(done)
	}
	b.currentBlock = done
	if isOutIf {
		// in exit if; set rest block
		rest := b.newBasicBlock("")
		b.emitJump(rest)

		// continue rest code
		b.currentBlock = rest
	}
}

// block
func (b *builder) buildBlock(stmt *yak.BlockContext) {
	if s, ok := stmt.StatementList().(*yak.StatementListContext); ok {
		b.buildStatementList(s)
	}
}

// assign expression
func (b *builder) buildAssignExpression(stmt *yak.AssignExpressionContext) {
	if op, op2 := stmt.AssignEq(), stmt.ColonAssignEq(); op != nil || op2 != nil {
		// right value
		var rvalues []Value
		if ri, ok := stmt.ExpressionList().(*yak.ExpressionListContext); ok {
			rvalues = b.buildExpressionList(ri)
		}

		// left
		var lvalues []LeftValue
		if li, ok := stmt.LeftExpressionList().(*yak.LeftExpressionListContext); ok {
			lvalues = b.buildLeftExpressionList(op2 != nil, li)
		}

		// assign
		if len(rvalues) == len(lvalues) {
			for i := range rvalues {
				lvalues[i].Assign(rvalues[i], b.Function)
			}
		}
	}

	if stmt.PlusPlus() != nil { // ++
		lvalue := b.buildLeftExpression(false, stmt.LeftExpression().(*yak.LeftExpressionContext))
		rvalue := b.emitArith(yakvm.OpAdd, lvalue.GetValue(b.Function), NewConst(int64(1)))
		lvalue.Assign(rvalue, b.Function)
	} else if stmt.SubSub() != nil { // --
		lvalue := b.buildLeftExpression(false, stmt.LeftExpression().(*yak.LeftExpressionContext))
		rvalue := b.emitArith(yakvm.OpSub, lvalue.GetValue(b.Function), NewConst(int64(1)))
		lvalue.Assign(rvalue, b.Function)
	}

	if op, ok := stmt.InplaceAssignOperator().(*yak.InplaceAssignOperatorContext); ok {
		rvalue := b.buildExpression(stmt.Expression().(*yak.ExpressionContext))
		lvalue := b.buildLeftExpression(false, stmt.LeftExpression().(*yak.LeftExpressionContext))
		var opcode yakvm.OpcodeFlag
		switch op.GetText() {
		case "+=":
			opcode = yakvm.OpAdd
		case "-=":
			opcode = yakvm.OpSub
		case "*=":
			opcode = yakvm.OpMul
		case "/=":
			opcode = yakvm.OpDiv
		case "%=":
			opcode = yakvm.OpMod
		case "<<=":
			opcode = yakvm.OpShl
		case ">>=":
			opcode = yakvm.OpShr
		case "&=":
			opcode = yakvm.OpAnd
		case "&^=":
			opcode = yakvm.OpAndNot
		case "|=":
			opcode = yakvm.OpOr
		case "^=":
			opcode = yakvm.OpXor

		}
		rvalue = b.emitArith(opcode, lvalue.GetValue(b.Function), rvalue)
		lvalue.Assign(rvalue, b.Function)
	}
}

//TODO: declear variable expression

// left expression list
func (f *builder) buildLeftExpressionList(forceAssign bool, stmt *yak.LeftExpressionListContext) []LeftValue {
	exprs := stmt.AllLeftExpression()
	valueLen := len(exprs)
	values := make([]LeftValue, valueLen)
	for i, e := range exprs {
		if e, ok := e.(*yak.LeftExpressionContext); ok {
			values[i] = f.buildLeftExpression(forceAssign, e)
		}
	}
	return values
}

// left  expression
func (f *builder) buildLeftExpression(forceAssign bool, stmt *yak.LeftExpressionContext) LeftValue {
	if s := stmt.Identifier(); s != nil {
		if v := f.readVariable(s.GetText()); v != nil {
			// when v exist
			switch v := v.(type) {
			case *Field:
				return v
			case *Parameter:
			default:
			}
		} else if !forceAssign && f.CanBuildFreeValue(s.GetText()) {
			field := f.parent.newField(s.GetText())
			f.FreeValues = append(f.FreeValues, field)
			f.parent.writeVariable(s.GetText(), field)
			f.writeVariable(s.GetText(), field)
			return field
		}
		return &IdentifierLV{
			variable: s.GetText(),
		}
	}
	if s, ok := stmt.Expression().(*yak.ExpressionContext); ok {
		expr := f.buildExpression(s)
		if expr == nil {
			panic("leftexpression expression is nil")
		}

		if s, ok := stmt.LeftSliceCall().(*yak.LeftSliceCallContext); ok {
			index := f.buildLeftSliceCall(s)
			if expr, ok := expr.(*Interface); ok {
				return f.emitField(expr, index)
			} else {
				panic("leftexprssion exprssion is not interface")
			}
		}

		//TODO: leftMemberCall
	}
	return nil
}

//TODO: left member call

// left slice call
func (f *builder) buildLeftSliceCall(stmt *yak.LeftSliceCallContext) Value {
	if s, ok := stmt.Expression().(*yak.ExpressionContext); ok {
		return f.buildExpression(s)
	}
	return nil
}

// expression
func (b *builder) buildExpression(stmt *yak.ExpressionContext) Value {
	//TODO: typeliteral expression

	// literal
	if s := stmt.Literal(); s != nil {
		//TODO: literal
		i, _ := strconv.ParseInt(s.GetText(), 10, 64)
		return NewConst(int64(i))
	}

	// anonymous function decl
	if s, ok := stmt.AnonymousFunctionDecl().(*yak.AnonymousFunctionDeclContext); ok {
		return b.buildAnonymouseFunctionDecl(s)
	}
	//TODO: panic

	//TODO: RECOVER

	// identifier
	if s := stmt.Identifier(); s != nil { // 解析变量
		text := s.GetText()
		if ret := b.readVariable(text); ret != nil {
			return ret
		} else if b.CanBuildFreeValue(text) {
			return b.BuildFreeValue(text)
		}
		fmt.Printf("debug undefine value: %v\n", s.GetText())
		panic("undefine value")
	}

	getExpr := func(index int) Value {
		if s, ok := stmt.Expression(index).(*yak.ExpressionContext); ok {
			return b.buildExpression(s)
		}
		return nil
	}
	//TODO: member call

	// slice call
	if s, ok := stmt.SliceCall().(*yak.SliceCallContext); ok {
		expr, ok := getExpr(0).(*Interface)
		if !ok {
			panic("expression slice need expression")
		}
		keys := b.buildSliceCall(s)
		if len(keys) == 1 {
			return b.emitField(expr, keys[0])
		} else if len(keys) == 2 {
			return b.emitInterfaceSlice(expr, keys[0], keys[1], nil)
		} else if len(keys) == 3 {
			return b.emitInterfaceSlice(expr, keys[0], keys[1], keys[2])
		} else {
			panic("")
		}
	}

	// function call
	if s, ok := stmt.FunctionCall().(*yak.FunctionCallContext); ok {
		v := getExpr(0)
		if v != nil {
			return b.buildFunctionCall(s, v)
		} else {
			panic("call target is nil")
		}
	}

	//TODO: parent expression

	//TODO: instance code

	// make expression
	if s, ok := stmt.MakeExpression().(*yak.MakeExpressionContext); ok {
		return b.buildMakeExpression(s)
	}
	//TODO: unary operator expression

	// //TODO: 二元运算（位运算全面优先于数字运算，数字运算全面优先于高级逻辑运算）
	// | expression bitBinaryOperator ws* expression

	// // 普通数学运算 done
	// | expression multiplicativeBinaryOperator ws* expression
	// | expression additiveBinaryOperator ws* expression
	// | expression comparisonBinaryOperator ws* expression

	// //TODO: 高级逻辑
	// | expression '&&' ws* expression
	// | expression '||' ws* expression
	// | expression 'not'? 'in' expression
	// | expression '<-' expression
	// | expression '?' ws* expression ws* ':' ws* expression
	// ;

	type op interface {
		GetText() string
	}
	getBinaryOp := func() op {
		if op := stmt.AdditiveBinaryOperator(); op != nil {
			return op
		}
		if op := stmt.MultiplicativeBinaryOperator(); op != nil {
			return op
		}
		if op := stmt.ComparisonBinaryOperator(); op != nil {
			return op
		}
		return nil
	}

	if op := getBinaryOp(); op != nil {
		op0 := getExpr(0)
		op1 := getExpr(1)
		if op0 == nil || op1 == nil {
			panic("additive binary operator need two expression")
		}
		var opcode yakvm.OpcodeFlag
		switch op.GetText() {
		// AdditiveBinaryOperator
		case "+":
			opcode = yakvm.OpAdd
		case "-":
			opcode = yakvm.OpSub

		// MultiplicativeBinaryOperator
		case "*":
			opcode = yakvm.OpMul
		case "/":
			opcode = yakvm.OpDiv
		case "%":
			opcode = yakvm.OpMod

		// ComparisonBinaryOperator
		case ">":
			opcode = yakvm.OpGt
		case "<":
			opcode = yakvm.OpLt
		case "<=":
			opcode = yakvm.OpLtEq
		case ">=":
			opcode = yakvm.OpGtEq
		case "!=", "<>":
			opcode = yakvm.OpNotEq
		case "==":
			opcode = yakvm.OpEq
		}
		return b.emitArith(opcode, op0, op1)
	}
	return nil
}

// paren expression

// make expression
func (b *builder) buildMakeExpression(stmt *yak.MakeExpressionContext) Value {
	var typ types.Type
	if s, ok := stmt.TypeLiteral().(*yak.TypeLiteralContext); ok {
		typ = b.buildTypeLiteral(s)
	}
	if typ == nil {
		panic("")
	}

	var exprs []Value
	if s, ok := stmt.ExpressionListMultiline().(*yak.ExpressionListMultilineContext); ok {
		exprs = b.buildExpressionListMultiline(s)
	}
	zero := NewConst(int64(0))
	switch typ.(type) {
	case *types.Slice:
		// fmt.Printf("debug %s %v %d\n", "make slice", typ, num)
		if len(exprs) == 0 {
			return b.emitInterfaceBuild(typ, zero, zero)
		} else if len(exprs) == 1 {
			return b.emitInterfaceBuild(typ, exprs[0], exprs[0])
		} else if len(exprs) == 2 {
			return b.emitInterfaceBuild(typ, exprs[0], exprs[1])
		} else {
			panic("make slice expression error!")
		}
	case *types.Map:
		fmt.Printf("debug %v\n", "make map")
	case *types.Chan:
		fmt.Printf("debug %v\n", "make chan")
	default:
		panic("make unknow type")
	}
	return nil
}

// type literal
func (b *builder) buildTypeLiteral(stmt *yak.TypeLiteralContext) types.Type {
	text := stmt.GetText()
	// var type name
	if b, ok := basicTypes[text]; ok {
		return b
	}

	// slice type literal
	if s, ok := stmt.SliceTypeLiteral().(*yak.SliceTypeLiteralContext); ok {
		return b.buildSliceTypeLiteral(s)
	}

	// map type literal
	if strings.HasPrefix(text, "map") {
		if s, ok := stmt.MapTypeLiteral().(*yak.MapTypeLiteralContext); ok {
			return b.buildMapTypeLiteral(s)
		}
	}

	// chan type literal
	if strings.HasPrefix(text, "chan") {
		if s, ok := stmt.TypeLiteral().(*yak.TypeLiteralContext); ok {
			if typ := b.buildTypeLiteral(s); typ != nil {
				return types.NewChan(types.SendRecv, typ)
			}
		}
	}

	return nil
}

// slice type literal
func (b *builder) buildSliceTypeLiteral(stmt *yak.SliceTypeLiteralContext) types.Type {
	if s, ok := stmt.TypeLiteral().(*yak.TypeLiteralContext); ok {
		if eleTyp := b.buildTypeLiteral(s); eleTyp != nil {
			return types.NewSlice(eleTyp)
		}
	}
	return nil
}

// map type literal
func (b *builder) buildMapTypeLiteral(stmt *yak.MapTypeLiteralContext) types.Type {
	// key
	var keyTyp types.Type
	var valueTyp types.Type
	if s, ok := stmt.TypeLiteral(0).(*yak.TypeLiteralContext); ok {
		keyTyp = b.buildTypeLiteral(s)
	}

	// value
	if s, ok := stmt.TypeLiteral(1).(*yak.TypeLiteralContext); ok {
		valueTyp = b.buildTypeLiteral(s)
	}
	if keyTyp != nil && valueTyp != nil {
		return types.NewMap(keyTyp, valueTyp)
	}

	return nil
}

//TODO: instance code

// anonymous function decl
func (b *builder) buildAnonymouseFunctionDecl(stmt *yak.AnonymousFunctionDeclContext) Value {
	funcName := ""
	if name := stmt.FunctionNameDecl(); name != nil {
		funcName = name.GetText()
	}
	newfunc := b.Package.NewFunctionWithParent(funcName, b.Function)
	b = &builder{
		Function: newfunc,
		next:     b, // push
	}

	if stmt.EqGt() != nil {
		if stmt.LParen() != nil && stmt.RParen() != nil {
			// has param
			// stmt FunctionParamDecl()
			if para, ok := stmt.FunctionParamDecl().(*yak.FunctionParamDeclContext); ok {
				b.buildFunctionParamDecl(para)
			}
		} else {
			// only this param
			newfunc.NewParam(stmt.Identifier().GetText())
		}
		if block, ok := stmt.Block().(*yak.BlockContext); ok {
			// build block
			b.buildBlock(block)
		} else if expression, ok := stmt.Expression().(*yak.ExpressionContext); ok {
			// hanlder expression
			v := b.buildExpression(expression)
			b.emitReturn([]Value{v})
		} else {
			panic("BUG: arrow function need expression or block at least")
		}
	} else {
		// this global function
		if para, ok := stmt.FunctionParamDecl().(*yak.FunctionParamDeclContext); ok {
			b.buildFunctionParamDecl(para)
		}
		if block, ok := stmt.Block().(*yak.BlockContext); ok {
			b.buildBlock(block)
		}
	}
	b.Finish()
	b = b.next // pop

	if funcName != "" {
		b.writeVariable(funcName, newfunc)
	}
	return newfunc
}

// function param decl
func (b *builder) buildFunctionParamDecl(stmt *yak.FunctionParamDeclContext) {
	ellipsis := stmt.Ellipsis() // if has "...",  use array pass this argument
	ids := stmt.AllIdentifier()

	for _, id := range ids {
		b.NewParam(id.GetText())
	}
	if ellipsis != nil {
		//TODO: handler "..." to array
		// param[len(ids)-1]
	}
}

// function call
func (b *builder) buildFunctionCall(stmt *yak.FunctionCallContext, v Value) Value {
	var args []Value
	isDropErr := false
	if s, ok := stmt.OrdinaryArguments().(*yak.OrdinaryArgumentsContext); ok {
		args = b.buildOrdinaryArguments(s)
	}
	if stmt.Wavy() != nil {
		isDropErr = true
	}
	return b.emitCall(v, args, isDropErr)
}

// ordinary argument
func (b *builder) buildOrdinaryArguments(stmt *yak.OrdinaryArgumentsContext) []Value {
	ellipsis := stmt.Ellipsis()
	allexpre := stmt.AllExpression()
	v := make([]Value, 0, len(allexpre))
	for _, expr := range allexpre {
		v = append(v, b.buildExpression(expr.(*yak.ExpressionContext)))
	}
	if ellipsis != nil {
		//TODO: handler "..." to array
		// v[len(v)-1]
	}
	return v
}

// TODO: member call

// slice call
func (b *builder) buildSliceCall(stmt *yak.SliceCallContext) []Value {
	exprs := stmt.AllExpression()
	values := make([]Value, len(exprs))
	if len(exprs) == 0 {
		panic("slicecall expression is zero")
	}
	if len(exprs) > 3 {
		panic("slicecall expression too much")
	}
	for i, expr := range exprs {
		if s, ok := expr.(*yak.ExpressionContext); ok {
			values[i] = b.buildExpression(s)
		}
	}
	return values
}

//TODO: literal

// expression list
func (b *builder) buildExpressionList(stmt *yak.ExpressionListContext) []Value {
	exprs := stmt.AllExpression()
	valueLen := len(exprs)
	values := make([]Value, valueLen)
	for i, e := range exprs {
		if e, ok := e.(*yak.ExpressionContext); ok {
			values[i] = b.buildExpression(e)
		}
	}
	return values
}

// expression list multiline
func (b *builder) buildExpressionListMultiline(stmt *yak.ExpressionListMultilineContext) []Value {
	allexpr := stmt.AllExpression()
	exprs := make([]Value, 0, len(allexpr))
	for _, expr := range allexpr {
		if expr, ok := expr.(*yak.ExpressionContext); ok {
			exprs = append(exprs, b.buildExpression(expr))
		}
	}
	return exprs
}

func (pkg *Package) build() {
	main := pkg.NewFunction("yak-main")
	b := builder{
		Function: main,
		next:     nil,
		target:   nil,
	}
	b.build(pkg.ast)
	b.Finish()
}

func (pkg *Package) Build() { pkg.buildOnece.Do(pkg.build) }

func (prog *Program) Build() {
	for _, pkg := range prog.Packages {
		pkg.Build()
	}
}
