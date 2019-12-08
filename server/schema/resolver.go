package schema

import (
	"fmt"

	"github.com/graphql-go/graphql"
	ast "github.com/graphql-go/graphql/language/ast"
	"github.com/pedox/gofar/server/resolve"
)

//https://github.com/graphql-go/graphql/issues/157#issuecomment-506439064
func (schema Schema) getResolveFields(param graphql.ResolveParams, fieldName string, selections []ast.Selection, callback resolve.PreResolveParamCallback, parent bool) (selected map[string]interface{}, err error) {
	selected = map[string]interface{}{}

	for _, s := range selections {
		switch s := s.(type) {
		case *ast.Field:
			if s.SelectionSet == nil {
				if _, ok := selected[s.Name.Value]; !ok {
					selected[s.Name.Value] = true
				}
			} else {
				//@todo must have s.Name.Value_id
				selected[s.Name.Value], err = schema.getResolveFields(param, s.Name.Value, s.SelectionSet.Selections, callback, false)
				if err != nil {
					return
				}
			}
		case *ast.FragmentSpread:
			n := s.Name.Value
			frag, ok := param.Info.Fragments[n]
			if !ok {
				err = fmt.Errorf("no fragment found with name %v", n)
				return
			}
			selected[s.Name.Value], err = schema.getResolveFields(param, s.Name.Value, frag.GetSelectionSet().Selections, callback, false)
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("found unexpected selection type %v", s)
			return
		}
	}

	if parent == true {
		// selected, _ = schema.resolve()
		selected, err = callback(resolve.PreResolveParam{
			FieldName:       fieldName,
			Param:           param,
			Fields:          selected,
			ParentFieldName: nil,
			ParentFields:    nil,
		})

		if err != nil {
			return nil, err
		}

	}

	return
}

// makeResolve return for resolve graphql actions
// since graphqlgo doesn't provide fields in default arguments,
// so we decide to added extra functions
func (schema Schema) makeResolve(fields *graphql.Object, callback resolve.PreResolveParamCallback) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (res interface{}, err error) {
		fieldASTs := p.Info.FieldASTs
		if len(fieldASTs) == 0 {
			return nil, fmt.Errorf("ResolveParams has no fields")
		}
		fieldName := fieldASTs[0].Name.Value
		return schema.getResolveFields(p, fieldName, fieldASTs[0].SelectionSet.Selections, callback, true)
	}
}
