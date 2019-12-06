package schema

import (
	"fmt"
	"reflect"

	ast "github.com/graphql-go/graphql/language/ast"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
)

func makeFieldList(name string, data graphql.Output) graphql.Output {
	objectFields := graphql.Fields{
		"page": &graphql.Field{
			Type: graphql.Int,
		},
		"lastPage": &graphql.Field{
			Type: graphql.Int,
		},
		"total": &graphql.Field{
			Type: graphql.Int,
		},
		"perPage": &graphql.Field{
			Type: graphql.Int,
		},
		"data": &graphql.Field{
			Type: graphql.NewList(data),
		},
	}
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   strcase.ToCamel(name + "List"),
			Fields: objectFields,
		},
	)
}

//https://github.com/graphql-go/graphql/issues/157#issuecomment-506439064
func (schema Schema) selectedFieldsFromSelections(p graphql.ResolveParams, fieldName string, selections []ast.Selection, parent bool) (selected map[string]interface{}, err error) {
	selected = map[string]interface{}{}

	for _, s := range selections {
		switch s := s.(type) {
		case *ast.Field:
			if s.SelectionSet == nil {
				if _, ok := selected[s.Name.Value]; !ok {
					selected[s.Name.Value] = "-"
				}
			} else {
				//@todo must have s.Name.Value_id
				selected[s.Name.Value], err = schema.selectedFieldsFromSelections(p, s.Name.Value, s.SelectionSet.Selections, false)
				if err != nil {
					return
				}
			}
		case *ast.FragmentSpread:
			n := s.Name.Value
			frag, ok := p.Info.Fragments[n]
			if !ok {
				err = fmt.Errorf("no fragment found with name %v", n)
				return
			}
			selected[s.Name.Value], err = schema.selectedFieldsFromSelections(p, s.Name.Value, frag.GetSelectionSet().Selections, false)
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("found unexpected selection type %v", s)
			return
		}
	}

	schema.preResolver(fieldName, selected, p, true)

	if parent == true {
		// jsn, _ := json.Marshal(selected)
		// for name := range selected {
		selected, _ = schema.resolve(fieldName, p, selected, nil, nil)
		// fmt.Println(string(jsn))
		// }
	}

	return
}

//makeResolve make resolve functions
func (schema Schema) makeResolve(fields *graphql.Object) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (res interface{}, err error) {
		fieldASTs := p.Info.FieldASTs
		if len(fieldASTs) == 0 {
			return nil, fmt.Errorf("ResolveParams has no fields")
		}
		fieldName := fieldASTs[0].Name.Value
		return schema.selectedFieldsFromSelections(p, fieldName, fieldASTs[0].SelectionSet.Selections, true)
	}
}

// Query Resolver
func (schema Schema) preResolver(modelName string, fields map[string]interface{}, p graphql.ResolveParams, parent bool) map[string]interface{} {
	fieldKeys := make([]string, 0, len(fields))
	nestedFields := make([]string, 0, len(fields))
	for key, typ := range fields {
		if reflect.TypeOf(typ).String()[0:4] == "map[" {
			nestedFields = append(nestedFields, key)
		} else {
			fieldKeys = append(fieldKeys, key)
		}
	}

	// Run from parent only
	if parent == true {
		md := schema.Models[modelName]
		for _, key := range fieldKeys {
			if key == "ID" {
				fields[key] = "string"
			} else {
				if _, ok := md[key]; ok {
					fields[key] = md[key].(map[string]interface{})["type"].(string)
				}
			}
		}

		for _, key := range nestedFields {
			fields[key] = schema.preResolver(key, fields[key].(map[string]interface{}), p, false)
		}
	}
	return fields
}

func (schema Schema) makeSingleQuery(modelName string, graphQLField *graphql.Object) {
	schema.queryFields[modelName] = &graphql.Field{
		Description: modelName + " Single data",
		Type:        graphQLField,
		Args: graphql.FieldConfigArgument{
			"ID": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: schema.makeResolve(graphQLField),
	}

}

func (schema Schema) makePagingQuery(modelName string, graphQLField *graphql.Object) {
	schema.queryFields[modelName+"List"] = &graphql.Field{
		Description: modelName + " Datasets",
		Type:        makeFieldList(modelName, graphQLField),
		Args: graphql.FieldConfigArgument{
			"page": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"perPage": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
		},
		Resolve: func(p graphql.ResolveParams) (res interface{}, err error) {
			dataList := []map[string]interface{}{}

			dataList = append(dataList, map[string]interface{}{
				"title":   "Judul",
				"content": "konten di sini",
			})

			page, _ := p.Args["page"]

			return map[string]interface{}{
				"page":     page,
				"lastPage": 1,
				"total":    10,
				"perPage":  10,
				"data":     dataList,
			}, nil
		},
	}
}

func (schema Schema) makeQueryFields() {
	for modelName, graphQLField := range schema.graphQLModels {
		// Single Node
		schema.makeSingleQuery(modelName, graphQLField)
		// Paging Node
		schema.makePagingQuery(modelName, graphQLField)
	}
}

func (schema Schema) makeQuery() *graphql.Object {
	schema.graphQLModels["aboutType"] = graphql.NewObject(
		graphql.ObjectConfig{
			Name: "About",
			Fields: graphql.Fields{
				"version": &graphql.Field{
					Type: graphql.String,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	schema.queryFields["about"] = &graphql.Field{
		Description: "Tentang aplikasi ini",
		Type:        schema.graphQLModels["aboutType"],
		Resolve: func(p graphql.ResolveParams) (res interface{}, err error) {
			return map[string]interface{}{
				"version": schema.Version,
				"name":    schema.Name,
			}, nil
		},
	}

	schema.makeQueryFields()

	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   "Query",
			Fields: schema.queryFields,
		},
	)
}
