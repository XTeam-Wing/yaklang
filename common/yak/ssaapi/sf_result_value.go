package ssaapi

import (
	"github.com/yaklang/yaklang/common/syntaxflow/sfvm"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/orderedmap"
	"github.com/yaklang/yaklang/common/yak/ssa/ssadb"
)

// ======================================== All Value/Variable ========================================

func (r *SyntaxFlowResult) GetVariableNum() int {
	if r == nil {
		return 0
	}
	if r.variable != nil {
		return r.variable.Len()
	}
	r.GetAllVariable()
	return r.variable.Len()
}

func (r *SyntaxFlowResult) GetAllVariable() *orderedmap.OrderedMap {
	if r == nil {
		return nil
	}
	if r.variable != nil {
		return r.variable
	}

	r.variable = orderedmap.New()
	if r.memResult != nil {
		r.memResult.SymbolTable.ForEach(func(name string, value sfvm.ValueOperator) bool {
			r.variable.Set(name, sfvm.ValuesLen(value))
			return true
		})
		for name := range r.memResult.AlertSymbolTable {
			if v, ok := r.variable.Get(name); ok && v.(int) > 0 {
				r.alertVariable = append(r.alertVariable, name)
			}
		}
	}

	if r.dbResult != nil {
		res, err := ssadb.GetResultVariableByID(ssadb.GetDB(), r.GetResultID())
		if err != nil {
			log.Errorf("err: %v", err)
			return nil
		}
		for _, v := range res {
			if v.Name == "_" {
				continue
			}
			r.variable.Set(v.Name, int(v.ValueNum))
			if v.HasRisk {
				r.alertVariable = append(r.alertVariable, v.Name)
			}
		}
		for _, name := range r.dbResult.UnValueVariable {
			if _, ok := r.variable.Get(name); !ok {
				r.variable.Set(name, 0)
			}
		}
	}

	return r.variable
}

func (r *SyntaxFlowResult) GetAllValuesChain() Values {
	var results Values
	m := r.GetAllVariable()
	if m == nil {
		return nil
	}
	m.ForEach(func(name string, value any) {
		vs := r.GetValues(name)
		results = append(results, vs...)
	})
	if len(results) == 0 {
		results = append(results, r.GetUnNameValues()...)
	}
	return results
}

func (r *SyntaxFlowResult) GetValueCount(name string) int {
	if r == nil {
		return 0
	}

	if r.variable == nil {
		r.GetAllVariable()
	}
	if v, ok := r.variable.Get(name); ok {
		if ret, ok := v.(int); ok {
			return ret
		}
	} else if name == "_" {
		return r.GetUnNameValues().Len()
	}
	return 0
}

// ======================================== Single Value ========================================

// Normal value
func (r *SyntaxFlowResult) GetValues(name string) Values {
	if r == nil {
		return nil
	}
	// unname
	if name == "_" {
		return r.GetUnNameValues()
	}
	// cache
	if vs, ok := r.symbol[name]; ok {
		return vs
	}

	// memory
	if r.memResult != nil {
		if v, ok := r.memResult.SymbolTable.Get(name); ok {
			vs := SyntaxFlowVariableToValues(v)
			r.symbol[name] = vs
			return vs
		}
	}
	if r.dbResult != nil {
		vs := r.getValueFromDB(name)
		r.symbol[name] = vs
		return vs
	}
	return nil
}

func (r *SyntaxFlowResult) GetValue(name string, index int) (*Value, error) {
	if r == nil {
		return nil, utils.Errorf("result is nil")
	}

	if name == "_" {
		return r.GetUnNameValues()[index], nil
	}

	if r.dbResult != nil {
		// for new DB data  have index
		id, err := ssadb.GetResultNodeByVariableIndex(ssadb.GetDB(), r.GetResultID(), name, uint(index))
		if err == nil {
			return r.program.NewValueFromAuditNode(id), nil
		}
	}

	// the old DB data and memory data can get by this
	vs := r.GetValues(name)
	if len(vs) > int(index) {
		return vs[index], nil
	} else {
		return nil, utils.Errorf("index out of range")
	}
}

// Alert value
func (r *SyntaxFlowResult) GetAlertVariables() []string {
	if r == nil {
		return nil
	}
	if r.alertVariable == nil {
		r.GetAllVariable()
	}
	return r.alertVariable
}

// UnName value
func (r *SyntaxFlowResult) GetUnNameValues() Values {
	if r == nil {
		return nil
	}
	if r.unName != nil {
		return r.unName
	}
	if r.memResult != nil {
		// memory
		r.unName = SyntaxFlowVariableToValues(r.memResult.UnNameValue)
	} else if r.dbResult != nil {
		// database
		r.unName = r.getValueFromDB("_")
	}
	return r.unName
}

func (r *SyntaxFlowResult) GetResultID() uint {
	if r == nil || r.dbResult == nil {
		return 0
	}
	return r.dbResult.ID
}

func (r *SyntaxFlowResult) getValueFromDB(name string) Values {
	auditNodeIDs, err := ssadb.GetResultNodeByVariable(ssadb.GetDB(), r.GetResultID(), name)
	if err != nil {
		return nil
	}

	vs := make(Values, 0, len(auditNodeIDs))
	for _, nodeID := range auditNodeIDs {
		if v := r.program.NewValueFromAuditNode(nodeID); v != nil {
			vs = append(vs, v)
		}
	}
	return vs
}
