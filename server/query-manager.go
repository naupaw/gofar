package main

import (
	"fmt"

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

func resolveFields(fields *graphql.Object, p graphql.ResolveParams) (map[string]interface{}, error) {
	fieldASTs := p.Info.FieldASTs
	if len(fieldASTs) == 0 {
		return nil, fmt.Errorf("getSelectedFields: ResolveParams has no fields")
	}
	fieldName := fieldASTs[0].Name.Value
	return selectedFieldsFromSelections(p, fieldName, fieldASTs[0].SelectionSet.Selections)
}

//https://github.com/graphql-go/graphql/issues/157#issuecomment-506439064
func selectedFieldsFromSelections(p graphql.ResolveParams, fieldName string, selections []ast.Selection) (selected map[string]interface{}, err error) {
	selected = map[string]interface{}{}
	fmt.Println("fieldName", fieldName)

	for _, s := range selections {
		switch s := s.(type) {
		case *ast.Field:
			if s.SelectionSet == nil {
				selected[s.Name.Value] = "the value that you want!"
			} else {
				selected[s.Name.Value], err = selectedFieldsFromSelections(p, s.Name.Value, s.SelectionSet.Selections)
				if err != nil {
					return
				}
			}
		case *ast.FragmentSpread:
			n := s.Name.Value
			frag, ok := p.Info.Fragments[n]
			if !ok {
				err = fmt.Errorf("getSelectedFields: no fragment found with name %v", n)
				return
			}
			selected[s.Name.Value], err = selectedFieldsFromSelections(p, s.Name.Value, frag.GetSelectionSet().Selections)
			if err != nil {
				return
			}
		default:
			err = fmt.Errorf("getSelectedFields: found unexpected selection type %v", s)
			return
		}
	}

	// for name := range selected {
	// 	selected[s.Name.Value]
	// 	fmt.Println("PROCESSED fields", fieldName, name)
	// }

	return
}

func makeResolve(fields *graphql.Object) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (res interface{}, err error) {

		// if len(fieldASTs) == 0 {
		// 	return nil, fmt.Errorf("getSelectedFields: ResolveParams has no fields")
		// }
		// selections := fieldASTs[0].SelectionSet.Selections

		// for _, s := range selections {
		// 	switch s := s.(type) {
		// 	case *ast.Field:
		// 		fmt.Println("FIELDS", s.Name.Value)
		// 		// if s.SelectionSet == nil {
		// 		// } else {
		// 		// 	fmt.Println("FIELD", s.SelectionSet.Selections)
		// 		// }
		// 	}
		// }
		return resolveFields(fields, p)
		// return map[string]interface{}{
		// 	"ID": "jsad781n2k3jncz8x-asdjnui13hn-123unc9aus9d",
		// }, nil
	}
}

func makeCollection(col graphql.Fields) graphql.Fields {
	for collection, fields := range dataTypes {
		collectionName := strcase.ToLowerCamel(collection)
		col[collectionName] = &graphql.Field{
			Description: collectionName + " Single data",
			Type:        fields,
			Args: graphql.FieldConfigArgument{
				"ID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: makeResolve(fields),
		}

		col[collectionName+"List"] = &graphql.Field{
			Description: collectionName + " Datasets",
			Type:        makeFieldList(collection, fields),
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

	return col
}

func makeQuery(schema MainSchema) *graphql.Object {
	aboutType := graphql.NewObject(
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

	collections := graphql.Fields{
		"about": &graphql.Field{
			Description: "Tentang aplikasi ini",
			Type:        aboutType,
			Resolve: func(p graphql.ResolveParams) (res interface{}, err error) {
				return map[string]interface{}{
					"version": schema.Version,
					"name":    schema.Name,
				}, nil
			},
		},
	}

	collections = makeCollection(collections)

	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   "Query",
			Fields: collections,
		},
	)
}

func defineSchema() {
	for name, fields := range models {
		if _, ok := dataTypes[name]; ok == false {
			makeModel(name, fields.(map[interface{}]interface{}), 1)
		}
	}
}
