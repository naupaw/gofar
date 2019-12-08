package schema

import (
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/pedox/gofar/server/model"
	"github.com/pedox/gofar/server/resolve"
)

type MutationResolveKind int

const (
	ResolveCreate MutationResolveKind = iota
	ResolveUpdate
	ResolveDelete
)

func checkRule(field model.Field) string {
	if validate, ok := field.Props["validate"]; ok {
		for _, rule := range strings.Split(validate, ",") {
			if rule == "required" {
				return rule
			}
		}
	}
	return ""
}

func (schema Schema) makeCreateMutation(modelName string, graphQLField *graphql.Object) {
	argFields := graphql.FieldConfigArgument{}

	mod := schema.models[modelName]
	for name, field := range mod.Fields {
		if _, ok := field.Props["relation"]; !ok {
			typed := schema.getTypeData(modelName, name, field.Type, nil)
			rule := checkRule(field)
			if name == "id" || rule == "required" {
				typed = graphql.NewNonNull(typed)
			}
			argFields[name] = &graphql.ArgumentConfig{
				Type: typed,
			}
		}
	}

	delete(argFields, "id")

	schema.mutationFields["Create"+modelName] = &graphql.Field{
		Description: "Create New " + modelName,
		Type:        graphQLField,
		Args:        argFields,
		Resolve: schema.makeResolve(graphQLField, func(param resolve.PreResolveParam) (resolve.ResolveParamResult, error) {
			param.FieldName = modelName
			return schema.resolveMutation(ResolveCreate, param)
		}),
	}
}

func (schema Schema) makeDeleteMutation(modelName string, graphQLField *graphql.Object) {
	schema.mutationFields["Delete"+modelName] = &graphql.Field{
		Description: "Delete " + modelName,
		Type:        graphQLField,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: schema.makeResolve(graphQLField, func(param resolve.PreResolveParam) (resolve.ResolveParamResult, error) {
			return schema.resolveMutation(ResolveDelete, param)
		}),
	}
}

func (schema Schema) makeUpdateMutation(modelName string, graphQLField *graphql.Object) {
	argFields := graphql.FieldConfigArgument{}

	mod := schema.models[modelName]
	for name, field := range mod.Fields {
		if _, ok := field.Props["relation"]; !ok {
			typed := schema.getTypeData(modelName, name, field.Type, nil)
			rule := checkRule(field)
			if name == "id" || rule == "required" {
				typed = graphql.NewNonNull(typed)
			}
			argFields[name] = &graphql.ArgumentConfig{
				Type: typed,
			}
		}
	}

	schema.mutationFields["Update"+modelName] = &graphql.Field{
		Description: "Update " + modelName,
		Type:        graphQLField,
		Args:        argFields,
		Resolve: schema.makeResolve(graphQLField, func(param resolve.PreResolveParam) (resolve.ResolveParamResult, error) {
			param.FieldName = modelName
			return schema.resolveMutation(ResolveUpdate, param)
		}),
	}
}

//resolveCreate resolve for single query execution
func (schema Schema) resolveMutation(kind MutationResolveKind, param resolve.PreResolveParam) (resolve.ResolveParamResult, error) {
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
		switch kind {
		case ResolveCreate:
			param.Fields = mod.Create(res)
		case ResolveUpdate:
			param.Fields = mod.Update(res)
		case ResolveDelete:
			param.Fields = mod.Delete(res)
		}
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
				fieldValue, _ := schema.resolveMutation(kind, param)
				param.Fields[name] = fieldValue
			}
		}
	}
	return param.Fields, nil
}
