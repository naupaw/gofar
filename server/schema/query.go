package schema

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	"github.com/pedox/gofar/server/resolve"
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

func (schema Schema) makeSingleQuery(modelName string, graphQLField *graphql.Object) {
	schema.queryFields[modelName] = &graphql.Field{
		Description: modelName + " Single data",
		Type:        graphQLField,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: schema.makeResolve(graphQLField, func(param resolve.PreResolveParam) (resolve.ResolveParamResult, error) {
			result, err := schema.resolveSingleQuery(param)
			return result, err
		}),
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

//resolveSingleQuery resolve for single query execution
func (schema Schema) resolveSingleQuery(param resolve.PreResolveParam) (resolve.ResolveParamResult, error) {
	fieldSet := map[string]resolve.ResolveType{}

	if _, ok := param.Fields["id"]; !ok {
		param.Fields["id"] = resolve.Primitive
	}

	//Pre defined foreign Key
	for name, typeData := range param.Fields {
		if reflect.TypeOf(typeData).Kind() == reflect.Map {
			fieldSet[name] = resolve.Relation
			fieldSet[name+"_id"] = resolve.Primitive
		} else {
			fieldSet[name] = resolve.Primitive
		}
	}

	for _, mod := range schema.loadedModules {
		res := resolve.Resolve{
			Param:      param.Param,
			FieldName:  param.FieldName,
			FieldTypes: fieldSet,
			Fields:     param.Fields,
		}
		param.Fields = mod.Query(res)
	}

	//let's evaluate relation query
	for name := range param.Fields {
		if resolveType, ok := fieldSet[name]; ok {
			if resolveType == resolve.Relation {
				param.Param.Args["id"] = param.Fields[name+"_id"]
				param.FieldName = name
				param.Fields = param.Fields[name].(map[string]interface{})
				param.ParentFieldName = &param.FieldName
				param.ParentFields = &param.Fields
				fieldValue, _ := schema.resolveSingleQuery(param)
				param.Fields[name] = fieldValue
			}
		}
	}
	return param.Fields, nil
}
