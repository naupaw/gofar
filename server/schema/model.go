package schema

import (
	"log"
	"reflect"
	"regexp"

	"github.com/graphql-go/graphql"
	tagparser "github.com/moznion/go-struct-custom-tag-parser"
)

//makeModel - create models from schema.yaml and collected to []GraphQLModels
func (schema Schema) makeModel(modelName string, modelFields Model) {
	fields := graphql.Fields{}
	gqlType := graphql.NewObject(
		graphql.ObjectConfig{
			Name:   modelName,
			Fields: fields,
		},
	)
	// Store to GraphQL Models
	schema.GraphQLModels[modelName] = gqlType

	// Automatically add ID fields
	fields["ID"] = &graphql.Field{
		Type: schema.getTypeData(modelName, "ID", "string `unique:\"true\"`", nil),
	}

	// Loop trough modelFields
	for fieldName, typeData := range modelFields {
		var outputTypeData graphql.Output = nil
		// Get All Fields except magic fields (starts with __)
		if fieldName[0:2] != "__" {
			v := reflect.ValueOf(typeData)
			switch v.Kind() {
			// If type is slice [fields]
			case reflect.Slice:
				s := reflect.ValueOf(typeData)
				subtype := s.Index(0).Interface().(string)
				outputTypeData = schema.getTypeData(modelName, fieldName, "slice", &subtype)
			default:
				outputTypeData = schema.getTypeData(modelName, fieldName, v.String(), nil)
			}
		}

		if outputTypeData != nil {
			fields[fieldName] = &graphql.Field{
				Type: outputTypeData,
			}
		}

	}

	log.Println(":: Initialize", modelName)
}

func getFieldTypeData(typeData string) (outputType string, props map[string]string) {
	props = map[string]string{}
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
	return typeData, props
}

//Get Supported Type Data
func (schema Schema) getTypeData(modelName string, fieldName string, typeData string, subType *string) graphql.Output {

	typeData, props := getFieldTypeData(typeData)
	if typeData != "slice" {
		// fmt
		//Define typedata and field prop
		schema.Models[modelName][fieldName] = map[string]interface{}{
			"type":  typeData,
			"props": props,
		}
	}

	switch typeData {
	case "string":
		return graphql.String
	case "number":
		return graphql.Int
	case "datetime":
		return graphql.DateTime
	case "slice":
		return graphql.NewList(schema.getTypeData(modelName, fieldName, *subType, nil))
	default:
		if fields, ok := schema.Models[typeData]; ok {
			if _, ok := schema.GraphQLModels[typeData]; !ok {
				schema.makeModel(typeData, fields)
				return schema.GraphQLModels[typeData]
			}
			return schema.GraphQLModels[typeData]
		}
		return nil
	}
}
