package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
)

// Query Resolver
func resolver(modelName string, fields map[string]interface{}, p graphql.ResolveParams, parent string) map[string]interface{} {
	fieldKeys := make([]string, 0, len(fields))
	nestedFields := make([]string, 0, len(fields))
	for key, typ := range fields {
		if reflect.TypeOf(typ).String()[0:4] == "map[" {
			nestedFields = append(nestedFields, key)
		} else {
			fieldKeys = append(fieldKeys, key)
		}
	}

	id, _ := p.Args["ID"].(string)

	idField := "id"

	// Run from parent only
	if parent == "" {
		fmt.Printf(
			"RUN SYNTAX SELECT %s FROM %s WHERE %s = \"%s\" \n\n",
			strings.Join(fieldKeys, ", "),
			modelName,
			idField,
			id,
		)
		fmt.Printf(
			"AND THEN RUN PROCEDURAL SCAN FOR NESTED FIELDS %s\n\n",
			strings.Join(nestedFields, ", "),
		)

		// out, _ := json.Marshal(fields)

		md := Models[strcase.ToCamel(modelName)].(map[interface{}]interface{})
		for _, key := range fieldKeys {
			if key == "ID" {
				fields[key] = "string"
			} else {
				typeData, _ := getFieldTypeData(md[key].(string))
				fields[key] = typeData
			}
		}
		for _, key := range nestedFields {
			fields[key] = resolver(key, fields[key].(map[string]interface{}), p, parent)
		}
		// fmt.Println("Output", string(out))
	}

	return fields
}
