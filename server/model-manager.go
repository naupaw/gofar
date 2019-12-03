package main

import (
	"fmt"
	"log"
	"reflect"
	"regexp"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	tagparser "github.com/moznion/go-struct-custom-tag-parser"
)

func makeModel(modelName string, fields map[interface{}]interface{}, depth int) {
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
		Type: getTypeData("string", nil, depth+1),
	}

	for field, typeData := range fields {
		fieldName := strcase.ToLowerCamel(field.(string))
		// Get All Fields except magic fields (starts with __)
		if field.(string)[0:2] != "__" {
			v := reflect.ValueOf(typeData)
			switch v.Kind() {
			// If type is slice [fields]
			case reflect.Slice:
				s := reflect.ValueOf(typeData)
				// get typeData
				subtype := s.Index(0).Interface().(string)
				td := getTypeData("slice", &subtype, depth+1)
				if td != nil {
					objectFields[fieldName] = &graphql.Field{
						Type: td,
					}
				}
			default:
				td := getTypeData(v.String(), nil, depth+1)
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

func getTypeData(typeData string, subType *string, depth int) graphql.Output {

	//Golang doesn't support backtick (`) escape right now :(
	rule := `(^[A-Za-z0-9]+)\s+`
	rule2 := `(.+)`
	regex := regexp.MustCompile(rule + "`" + rule2 + "`")

	var res = regex.FindStringSubmatch(typeData)

	if len(res) > 2 {
		fmt.Printf("typeData=%s prop=%s\n\n", res[1], res[2])
		result, err := tagparser.Parse(res[2], true)
		if err != nil {
			log.Fatalf("unexpected error has come: %s", err)
		}

		for f, prop := range result {
			fmt.Println(f, prop)
		}

		typeData = res[1]
	}

	switch typeData {
	case "string":
		return graphql.String
	case "number":
		return graphql.Int
	case "datetime":
		return graphql.DateTime
	case "slice":
		return graphql.NewList(getTypeData(*subType, nil, depth))
	default:
		if fields, ok := models[typeData]; ok {
			if _, ok := dataTypes[typeData]; !ok {
				makeModel(typeData, fields.(map[interface{}]interface{}), depth)
				return dataTypes[typeData]
			}
			return dataTypes[typeData]
		}
		return nil
	}
}
