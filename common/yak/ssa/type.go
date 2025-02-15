package ssa

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/samber/lo"
	"github.com/yaklang/yaklang/common/utils"
	"golang.org/x/exp/slices"
)

func init() {
	BasicTypes[ErrorType].method["Error"] = NewFunctionType(
		"error.Error",
		[]Type{BasicTypes[ErrorType]},
		[]Type{BasicTypes[String]},
		false,
	)
}
func IsObjectType(t Type) bool {
	switch t.GetTypeKind() {
	case ObjectTypeKind, SliceTypeKind, MapTypeKind, StructTypeKind:
		return true
	default:
		return false
	}
}

type Type interface {
	String() string
	RawString() string
	GetTypeKind() TypeKind

	// set/get method
	SetMethod(map[string]*FunctionType)
	GetMethod(id string) *FunctionType
	GetAllKey() []string
}
type Types []Type // each value can have multiple type possible

// return true  if org != typs
// return false if org == typs
func (org Types) Compare(typs Types) bool {
	if len(org) == 0 && len(typs) != 0 {
		return true
	}
	return slices.CompareFunc(org, typs, func(org, typ Type) int {
		if org.String() == typs.String() {
			return 0
		}
		return 1
	}) != 0
}

func (t Types) String() string {
	return strings.Join(
		lo.Map(t, func(typ Type, _ int) string {
			if typ == nil {
				return "nil"
			} else {
				return typ.String()
			}
		}),
		", ",
	)
}

func (t Types) Equal(typs Types) bool {
	if len(t) != len(typs) {
		return false
	}
	return reflect.DeepEqual(t, typs)
}

func (t Types) Contains(typ Types) bool {
	if len(t) == 0 {
		return false
	}
	targetMap := lo.SliceToMap(typ, func(typ Type) (Type, struct{}) {
		return typ, struct{}{}
	})
	for _, tt := range t {
		if _, ok := targetMap[tt]; ok {
			return true
		}
	}
	return false
}

func (t Types) IsType(kind TypeKind) bool {
	for _, typ := range t {
		if typ.GetTypeKind() == kind {
			return true
		}
	}
	return false
}

// basic type
type TypeKind int

const (
	Number TypeKind = iota
	String
	Boolean
	UndefinedType // undefined is nil in golang
	Null          //
	Any           // any type
	ChanTypeKind
	ErrorType

	ObjectTypeKind
	SliceTypeKind
	MapTypeKind
	StructTypeKind

	InterfaceTypeKind
	FunctionTypeKind
)

type BasicType struct {
	Kind TypeKind
	name string

	method map[string]*FunctionType
}

func (b *BasicType) String() string {
	return b.name
}

func (b *BasicType) RawString() string {
	return b.name
}

func (b *BasicType) GetTypeKind() TypeKind {
	return b.Kind
}
func (b *BasicType) GetMethod(id string) *FunctionType {
	if v, ok := b.method[id]; ok {
		return v
	} else {
		return nil
	}
}
func (b *BasicType) SetMethod(method map[string]*FunctionType) {
	b.method = method
}
func (b *BasicType) GetAllKey() []string {
	return lo.Keys(b.method)
}

var _ Type = (*BasicType)(nil)

var BasicTypes = []*BasicType{
	Number:        {Number, "number", make(map[string]*FunctionType, 0)},
	String:        {String, "string", make(map[string]*FunctionType, 0)},
	Boolean:       {Boolean, "boolean", make(map[string]*FunctionType, 0)},
	UndefinedType: {UndefinedType, "undefined", make(map[string]*FunctionType, 0)},
	Null:          {Null, "null", make(map[string]*FunctionType, 0)},
	Any:           {Any, "any", make(map[string]*FunctionType, 0)},
	ErrorType:     {ErrorType, "error", make(map[string]*FunctionType, 0)},
}

func GetType(i any) Type {
	if typ := GetTypeByStr(reflect.TypeOf(i).String()); typ != nil {
		return typ
	} else {
		panic("undefined type")
	}
}
func GetTypeByStr(typ string) Type {
	switch typ {
	case "uint", "uint8", "byte", "uint16", "uint32", "uint64", "int", "int8", "int16", "int32", "int64", "uintptr":
		return BasicTypes[Number]
	case "float", "float32", "float64", "double", "complex128", "complex64":
		return BasicTypes[Number]
	case "string":
		return BasicTypes[String]
	case "bool":
		return BasicTypes[Boolean]
	case "interface {}", "var":
		return BasicTypes[Any]
	case "error":
		return BasicTypes[ErrorType]
	default:
		return nil
	}
}

// ====================== alias type
type AliasType struct {
	elem   Type
	method map[string]*FunctionType
	Name   string
}

var _ Type = (*AliasType)(nil)

func NewAliasType(name string, elem Type) *AliasType {
	return &AliasType{
		elem:   elem,
		method: make(map[string]*FunctionType),
		Name:   name,
	}
}

func (a *AliasType) SetMethod(m map[string]*FunctionType) {
	a.method = m
}

func (a *AliasType) GetMethod(id string) *FunctionType {
	if v, ok := a.method[id]; ok {
		return v
	} else {
		return nil
	}
}

func (b *AliasType) GetAllKey() []string {
	return lo.Keys(b.method)
}

func (a *AliasType) String() string {
	if a.Name != "" {
		return a.Name
	} else {
		return a.RawString()
	}
}

func (a *AliasType) RawString() string {
	return fmt.Sprintf("type %s (%s)", a.Name, a.elem)
}

func (a *AliasType) GetTypeKind() TypeKind {
	return a.elem.GetTypeKind()
}

// ====================== interface type
type InterfaceType struct {
	method map[string]*FunctionType
	name   string
}

func NewInterfaceType(name string) *InterfaceType {
	return &InterfaceType{
		method: make(map[string]*FunctionType),
		name:   name,
	}
}

var _ Type = (*InterfaceType)(nil)

func (i *InterfaceType) SetMethod(m map[string]*FunctionType) {
	i.method = m
}

func (i *InterfaceType) GetMethod(id string) *FunctionType {
	if v, ok := i.method[id]; ok {
		return v
	} else {
		return nil
	}
}
func (b *InterfaceType) GetAllKey() []string {
	return lo.Keys(b.method)
}

func (i *InterfaceType) GetTypeKind() TypeKind {
	return InterfaceTypeKind
}

func (i *InterfaceType) String() string {
	if i.name != "" {
		return i.name
	} else {
		return i.RawString()
	}
}

func (i *InterfaceType) RawString() string {
	return fmt.Sprintf("type %s interface{}", i.name)
}

// ====================== chan type
type ChanType struct {
	elem   Type
	method map[string]*FunctionType
}

var _ (Type) = (*ChanType)(nil)

func (c *ChanType) SetMethod(m map[string]*FunctionType) {
	c.method = m
}
func (c *ChanType) GetMethod(id string) *FunctionType {
	return c.method[id]
}
func (b *ChanType) GetAllKey() []string {
	return lo.Keys(b.method)
}

func (c *ChanType) GetTypeKind() TypeKind {
	return ChanTypeKind
}

func NewChanType(elem Type) *ChanType {
	return &ChanType{
		elem: elem,
	}
}

func (c ChanType) String() string {
	return fmt.Sprintf("chan %s", c.elem)
}

func (c ChanType) RawString() string {
	return c.String()
}

// ==================== interface type
type ObjectType struct {
	Name       string
	Kind       TypeKind
	Len        int
	Key        []Value
	keyTypes   []Type
	FieldTypes []Type

	AnonymousField []*ObjectType

	Combination bool // function multiple return will combined to struct
	// VariadicPara bool // function last variadic parameter will become slice

	method map[string]*FunctionType

	KeyTyp    Type
	FieldType Type
}

func (i *ObjectType) GetTypeKind() TypeKind {
	return i.Kind
}

func (i *ObjectType) GetMethod(id string) *FunctionType {
	if v, ok := i.method[id]; ok {
		return v
	} else {
		return nil
	}
}

func (i *ObjectType) SetMethod(m map[string]*FunctionType) {
	i.method = m
}

func (b *ObjectType) GetAllKey() []string {
	return append(lo.Keys(b.method), lo.Map(b.Key, func(v Value, _ int) string { return v.String() })...)
}

var _ (Type) = (*ObjectType)(nil)

func (i *ObjectType) SetName(name string) {
	i.Name = name
}

func NewObjectType() *ObjectType {
	return &ObjectType{
		Kind:       ObjectTypeKind,
		Key:        make([]Value, 0),
		keyTypes:   make([]Type, 0),
		FieldTypes: make([]Type, 0),
		method:     make(map[string]*FunctionType, 0),
	}
}

// for slice build
func NewSliceType(elem Type) *ObjectType {
	i := NewObjectType()
	i.Kind = SliceTypeKind
	i.KeyTyp = BasicTypes[Number]
	i.FieldType = elem
	return i
}

func NewMapType(key, field Type) *ObjectType {
	i := NewObjectType()
	i.KeyTyp = key
	i.FieldType = field
	i.Kind = MapTypeKind
	return i
}

func NewStructType() *ObjectType {
	i := NewObjectType()
	i.Kind = StructTypeKind
	return i
}

func (itype ObjectType) String() string {
	if itype.Combination {
		return strings.Join(
			lo.Map(
				itype.FieldTypes,
				func(t Type, _ int) string { return t.String() },
			),
			", ",
		)
	}
	if itype.Name != "" {
		return itype.Name
	}
	return itype.RawString()
}

func (itype ObjectType) RawString() string {
	ret := ""
	switch itype.Kind {
	case SliceTypeKind:
		// map[int]T
		if itype.Len == 0 {
			ret += fmt.Sprintf("[]%s", itype.FieldType.String())
		} else {
			ret += fmt.Sprintf("[%d]%s", itype.Len, itype.FieldType.String())
		}
	case MapTypeKind:
		// map[T]U
		// if len(itype.keyType) == 1 && len(itype.Field) == 1 {
		keyTyp := itype.KeyTyp
		if utils.IsNil(keyTyp) {
			keyTyp = BasicTypes[Any]
		}
		fieldType := itype.FieldType
		if utils.IsNil(fieldType) {
			fieldType = BasicTypes[Any]
		}
		ret += fmt.Sprintf("map[%s]%s", keyTyp.String(), fieldType.String())
		// } else {
		// 	panic("this interface type not map")
		// }
	case StructTypeKind:
		// map[string](T/U/xx)
		ret += fmt.Sprintf(
			"struct {%s}",
			strings.Join(
				lo.Map(itype.FieldTypes, func(field Type, _ int) string { return field.String() }),
				",",
			),
		)
	case ObjectTypeKind:
		ret += "object{}"
	}
	return ret
}

// for struct build
func (s *ObjectType) AddField(key Value, field Type) {
	s.Key = append(s.Key, key)
	keyTyp := key.GetType()
	s.keyTypes = append(s.keyTypes, keyTyp)
	if field == nil {
		field = BasicTypes[Any]
	}
	s.FieldTypes = append(s.FieldTypes, field)
}

// return (field-type, key-type)
func (s *ObjectType) GetField(key Value) Type {
	switch s.Kind {
	case SliceTypeKind, MapTypeKind:
		if key.GetType() == s.KeyTyp {
			return s.FieldType
		}
	case StructTypeKind:
		getField := func(o *ObjectType) Type {
			if index := slices.IndexFunc(o.Key, func(v Value) bool { return v.String() == key.String() }); index != -1 {
				return o.FieldTypes[index]
			} else {
				return nil
			}
		}
		if t := getField(s); t != nil {
			return t
		}
		for _, obj := range s.AnonymousField {
			if t := getField(obj); t != nil {
				return t
			}
		}
	}
	return nil
}

// ===================== Finish simply
func (s *ObjectType) Finish() {
	if s.Kind != ObjectTypeKind {
		return
	}
	fieldTypes := lo.UniqBy(s.FieldTypes, func(t Type) TypeKind { return t.GetTypeKind() })
	keyTypes := lo.UniqBy(s.keyTypes, func(t Type) TypeKind { return t.GetTypeKind() })
	if len(keyTypes) == 1 {
		if len(fieldTypes) == 1 {
			// map[T]U
			if keyTypes[0].GetTypeKind() == Number {
				// map[number]T ==> []T slice
				// TODO: check increasing
				s.Kind = SliceTypeKind
				s.KeyTyp = BasicTypes[Number]
				s.FieldType = fieldTypes[0]
			} else {
				// Map
				s.Kind = MapTypeKind
				s.KeyTyp = keyTypes[0]
				s.FieldType = fieldTypes[0]
			}
			// s.keyType = keyType
			// s.Field = field
		} else if keyTypes[0].GetTypeKind() == String || keyTypes[0].GetTypeKind() == Number {
			s.Kind = MapTypeKind
			s.KeyTyp = BasicTypes[String]
			s.FieldType = BasicTypes[Any]
		}
	}
}

type FunctionType struct {
	Name       string
	ReturnType Type
	Parameter  Types
	FreeValue  map[string]bool
	IsVariadic bool
}

var _ Type = (*FunctionType)(nil)

func (f *FunctionType) GetMethod(string) *FunctionType {
	return nil
}

func (f *FunctionType) SetMethod(m map[string]*FunctionType) {
}
func (b *FunctionType) GetAllKey() []string {
	return []string{}
}

func CalculateType(ts []Type) Type {
	if len(ts) == 0 {
		return BasicTypes[Null]
	} else if len(ts) == 1 {
		return ts[0]
	} else {
		i := NewStructType()
		for index, typ := range ts {
			i.AddField(NewConst(index), typ)
		}
		i.Finish()
		i.Combination = true
		// i.SetLen(NewConst(len(ts)))
		i.Len = len(ts)
		return i
	}
}

func NewFunctionType(name string, Parameter []Type, ReturnType []Type, IsVariadic bool) *FunctionType {
	f := &FunctionType{
		Name:       name,
		Parameter:  Parameter,
		IsVariadic: IsVariadic,
	}
	f.ReturnType = CalculateType(ReturnType)
	return f
}

func (s *FunctionType) SetFreeValue(fv map[string]bool) {
	s.FreeValue = fv
}

func (s *FunctionType) SetName(name string) {
	s.Name = name
}

func (s *FunctionType) String() string {
	if s.Name != "" {
		return s.Name
	}
	return s.RawString()
}

func (s *FunctionType) RawString() string {
	str := ""
	if s.IsVariadic {
		str += "..."
	}

	return fmt.Sprintf(
		"(%s %s) -> %s",
		strings.Join(
			lo.Map(s.Parameter, func(t Type, _ int) string { return t.String() }),
			",",
		),
		str,
		s.ReturnType,
	)
}

func (s *FunctionType) GetParamString() string {
	ret := ""
	for index, t := range s.Parameter {
		if index == len(s.Parameter)-1 {
			if s.IsVariadic {
				if obj, ok := ToObjectType(t); ok && obj.Kind == SliceTypeKind {
					// last
					ret += "..." + obj.FieldType.String()
				}
			} else {
				ret += t.String()

			}
		} else {
			ret += t.String() + ", "
		}
	}
	return ret
}

func (s *FunctionType) GetTypeKind() TypeKind {
	return FunctionTypeKind
}
