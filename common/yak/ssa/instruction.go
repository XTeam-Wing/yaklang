package ssa

import (
	"github.com/samber/lo"
)

func NewJump(to *BasicBlock) *Jump {
	j := &Jump{
		anInstruction: NewInstruction(),
		To:            to,
	}
	return j
}

func NewLoop(cond Value) *Loop {
	l := &Loop{
		anInstruction: NewInstruction(),
		Cond:          cond,
	}
	return l
}

func NewConstInst(c *Const) *ConstInst {
	v := &ConstInst{
		Const:         c,
		anInstruction: NewInstruction(),
	}
	return v
}

func NewUndefined(name string) *Undefined {
	u := &Undefined{
		anInstruction: NewInstruction(),
		anValue:       NewValue(),
	}
	u.SetVariable(name)
	return u
}

func NewBinOpOnly(op BinaryOpcode, x, y Value) *BinOp {
	b := &BinOp{
		anInstruction: NewInstruction(),
		anValue:       NewValue(),
		Op:            op,
		X:             x,
		Y:             y,
	}
	if op >= OpGt && op <= OpIn {
		b.SetType(BasicTypes[Boolean])
	}
	return b
}

func NewBinOp(op BinaryOpcode, x, y Value) Value {
	v := HandlerBinOp(NewBinOpOnly(op, x, y))
	return v
}

func NewUnOpOnly(op UnaryOpcode, x Value) *UnOp {
	u := &UnOp{
		anInstruction: NewInstruction(),
		anValue:       NewValue(),
		Op:            op,
		X:             x,
	}
	return u
}

func NewUnOp(op UnaryOpcode, x Value) Value {
	b := HandlerUnOp(NewUnOpOnly(op, x))
	return b
}

func NewIf(cond Value) *If {
	ifSSA := &If{
		anInstruction: NewInstruction(),
		Cond:          cond,
	}
	return ifSSA
}

func NewSwitch(cond Value, defaultb *BasicBlock, label []SwitchLabel) *Switch {
	sw := &Switch{
		anInstruction: NewInstruction(),
		Cond:          cond,
		DefaultBlock:  defaultb,
		Label:         label,
	}
	return sw
}

func NewReturn(vs []Value) *Return {
	r := &Return{
		anInstruction: NewInstruction(),
		Results:       vs,
	}
	return r
}

func NewTypeCast(typ Type, v Value) *TypeCast {
	t := &TypeCast{
		anInstruction: NewInstruction(),
		anValue:       NewValue(),
		Value:         v,
	}
	t.SetType(typ)
	return t
}

func NewTypeValue(typ Type) *TypeValue {
	t := &TypeValue{
		anInstruction: NewInstruction(),
		anValue:       NewValue(),
	}
	t.SetType(typ)
	return t
}

func NewAssert(cond, msgValue Value, msg string) *Assert {
	a := &Assert{
		anInstruction: NewInstruction(),
		Cond:          cond,
		Msg:           msg,
		MsgValue:      msgValue,
	}
	return a
}

var NextType *ObjectType = nil

func NewNext(iter Value, isIn bool) *Next {
	n := &Next{
		anInstruction: NewInstruction(),
		anValue:       NewValue(),
		Iter:          iter,
		InNext:        isIn,
	}
	if NextType == nil {
		NextType = NewObjectType()
		NextType.Kind = StructTypeKind
		NextType.AddField(NewConst("ok"), BasicTypes[Boolean])
		NextType.AddField(NewConst("key"), BasicTypes[Any])
		NextType.AddField(NewConst("field"), BasicTypes[Any])
	}
	n.SetType(NextType)
	return n
}

func NewErrorHandler(try, catch *BasicBlock) *ErrorHandler {
	e := &ErrorHandler{
		anInstruction: NewInstruction(),
		try:           try,
		catch:         catch,
	}
	// block.AddSucc(try)
	try.Handler = e
	// block.AddSucc(catch)
	catch.Handler = e
	return e
}

func NewParam(variable string, isFreeValue bool, fun *Function) *Parameter {
	p := &Parameter{
		anInstruction: NewInstruction(),
		anValue:       NewValue(),
		variable:      variable,
		IsFreeValue:   isFreeValue,
	}
	p.SetFunc(fun)
	p.SetBlock(fun.EnterBlock)
	p.SetPosition(fun.GetPosition())
	return p
}

func (i *If) AddTrue(t *BasicBlock) {
	i.True = t
	i.GetBlock().AddSucc(t)
}

func (i *If) AddFalse(f *BasicBlock) {
	i.False = f
	i.GetBlock().AddSucc(f)
}

func (l *Loop) Finish(init, step []Value) {
	// check cond
	check := func(v Value) bool {
		if _, ok := ToPhi(v); ok {
			return true
		} else {
			return false
		}
	}

	if b, ok := l.Cond.(*BinOp); ok {
		// if b.Op < OpGt || b.Op > OpNotEq {
		// 	l.NewError(Error, SSATAG, "this condition not compare")
		// }
		if check(b.X) {
			l.Key = b.X
		} else if check(b.Y) {
			l.Key = b.Y
			// } else {
			// l.NewError(Error, SSATAG, "this condition not change")
		}
	}

	if l.Key == nil {
		return
	}
	tmp := lo.SliceToMap(l.Key.GetValues(), func(v Value) (Value, struct{}) { return v, struct{}{} })

	set := func(vs []Value) Value {
		for _, v := range vs {
			if _, ok := tmp[v]; ok {
				return v
			}
		}
		return nil
	}

	l.Init = set(init)
	l.Step = set(step)

	fixupUseChain(l)
}

func (e *ErrorHandler) AddFinal(f *BasicBlock) {
	e.final = f
	e.GetBlock().AddSucc(f)
	f.Handler = e
}

func (e *ErrorHandler) AddDone(d *BasicBlock) {
	e.done = d
	// just mark in instruction
	// e.GetBlock().AddSucc(d)
	d.Handler = e
}
