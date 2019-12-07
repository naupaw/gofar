package schema

import (
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/pedox/gofar/server/resolve"
)

func (schema Schema) resolve(fieldName string, param graphql.ResolveParams, fields map[string]interface{}, parentFieldName *string, parentFields *map[string]interface{}) (map[string]interface{}, error) {
	fieldSet := map[string]resolve.ResolveType{}

	if _, ok := fields["id"]; !ok {
		fields["id"] = resolve.Primitive
	}

	//Pre defined foreign Key
	for name, typeData := range fields {
		if reflect.TypeOf(typeData).Kind() == reflect.Map {
			fieldSet[name] = resolve.Relation
			fieldSet[name+"_id"] = resolve.Primitive
		} else {
			fieldSet[name] = resolve.Primitive
		}
	}

	for _, mod := range schema.loadedModules {
		res := resolve.Resolve{
			Param:      param,
			FieldName:  fieldName,
			FieldTypes: fieldSet,
			Fields:     fields,
		}
		fields = mod.Query(res)
	}

	for name, resolveType := range fieldSet {
		if resolveType == resolve.Relation {
			param.Args["ID"] = fields[name+"_id"]
			fieldValue, _ := schema.resolve(
				name,
				param,
				fields[name].(map[string]interface{}),
				&fieldName,
				&fields,
			)
			fields[name] = fieldValue
		}
	}

	return fields, nil
}
