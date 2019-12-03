package main

import (
	"log"
	"reflect"
	"regexp"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	tagparser "github.com/moznion/go-struct-custom-tag-parser"
)

func makeModel(modelName string, fields Model) {
	objectFields := graphql.Fields{}
	gqlType := graphql.NewObject(
		graphql.ObjectConfig{
			Name:   modelName,
			Fields: objectFields,
		},
	)
	dataTypes[modelName] = gqlType

	// Automatically add ID fields
	objectFields["ID"] = &graphql.Field{
		Type: getTypeData("string", nil),
	}

	for field, typeData := range fields {
		fieldName := strcase.ToLowerCamel(field)
		// Get All Fields except magic fields (starts with __)
		if field[0:2] != "__" {
			v := reflect.ValueOf(typeData)
			switch v.Kind() {
			// If type is slice [fields]
			case reflect.Slice:
				s := reflect.ValueOf(typeData)
				// get typeData
				subtype := s.Index(0).Interface().(string)
				td := getTypeData("slice", &subtype)
				if td != nil {
					objectFields[fieldName] = &graphql.Field{
						Type: td,
					}
				}
			default:
				td := getTypeData(v.String(), nil)
				if td != nil {
					objectFields[fieldName] = &graphql.Field{
						Type: td,
					}
				}
			}
		}
	}

	log.Println(":: Initialize", modelName)
}

func getFieldTypeData(typeData string) (outputType string, opt map[string]string) {
	opt = map[string]string{}
	//Golang doesn't support backtick (`) escape right now :(
	rule := `(^[A-Za-z0-9]+)\s+`
	rule2 := `(.+)`
	regex := regexp.MustCompile(rule + "`" + rule2 + "`")

	var res = regex.FindStringSubmatch(typeData)

	if len(res) > 2 {
		// fmt.Printf("typeData=%s prop=%s\n\n", res[1], res[2])
		result, err := tagparser.Parse(res[2], true)
		if err != nil {
			log.Fatalf("unexpected error has come: %s", err)
		}

		// for f, prop := range result {
		// 	fmt.Println(f, prop)
		// }
		return res[1], result
	}
	return typeData, opt
}

func getTypeData(typeData string, subType *string) graphql.Output {

	//Golang doesn't support backtick (`) escape right now :(
	typeData, _ = getFieldTypeData(typeData)

	switch typeData {
	case "string":
		return graphql.String
	case "number":
		return graphql.Int
	case "datetime":
		return graphql.DateTime
	case "slice":
		return graphql.NewList(getTypeData(*subType, nil))
	default:
		if fields, ok := ModelLists[typeData]; ok {
			if _, ok := dataTypes[typeData]; !ok {
				makeModel(typeData, fields)
				return dataTypes[typeData]
			}
			return dataTypes[typeData]
		}
		return nil
	}
}
