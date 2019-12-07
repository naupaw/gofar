package schema

import (
	"log"
	"reflect"
	"regexp"

	"github.com/graphql-go/graphql"
	tagparser "github.com/moznion/go-struct-custom-tag-parser"
	"github.com/pedox/gofar/server/model"
)

//makeModel - create models from schema.yaml and collected to []graphQLModels
func (schema Schema) makeModel(modelName string, modelFields Model) {
	fields := graphql.Fields{}
	gqlType := graphql.NewObject(
		graphql.ObjectConfig{
			Name:   modelName,
			Fields: fields,
		},
	)

	// Store to GraphQL Models
	schema.graphQLModels[modelName] = gqlType
	// and store to models to!
	schema.models[modelName] = model.Model{
		Name:   modelName,
		Fields: map[string]model.Field{},
	}

	// Automatically add id fields
	modelFields["id"] = "string `unique:\"true\"`"

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

		if fieldName[0:2] == "__" {
			options := map[string]interface{}{}
			for name, val := range typeData.(map[interface{}]interface{}) {
				options[name.(string)] = val
			}
			modelFields[fieldName] = options
		}

		if outputTypeData != nil {
			fields[fieldName] = &graphql.Field{
				Type: outputTypeData,
			}

		}

	}

	//CALL CreateModel
	for _, mod := range schema.loadedModules {
		mod.CreateModel(schema.models[modelName])
	}

	if schema.Debug {
		log.Println(":: Initialize", modelName)
	}
}

func getFieldTypeData(typeData string) (outputType string, props map[string]string) {
	props = map[string]string{}
	//Golang doesn't support backtick (`) escape right now :(
	rule := `(^[A-Za-z0-9]+)\s+`
	rule2 := `(.+)`
	regex := regexp.MustCompile(rule + "`" + rule2 + "`")

	var res = regex.FindStringSubmatch(typeData)

	if len(res) > 2 {
		result, err := tagparser.Parse(res[2], true)
		if err != nil {
			log.Fatalf("unexpected error has come: %s", err)
		}
		return res[1], result
	}
	return typeData, props
}

func (schema Schema) appendSchemaProps(modelName string, fieldName string, typeData string, subType *string) string {
	field := schema.models[modelName].Fields[fieldName]

	typeData, props := getFieldTypeData(typeData)
	if typeData != "slice" {
		if subType != nil {
			if *subType == "_SLICE_" {
				props["relation"] = "hasMany"
			}
			if *subType == "_MODEL_" {
				props["relation"] = "hasOne"
			}
		}
		field.Type = typeData
		field.Props = props
	}

	schema.models[modelName].Fields[fieldName] = field

	return typeData
}

//Get Supported Type Data
func (schema Schema) getTypeData(modelName string, fieldName string, typeData string, subType *string) graphql.Output {
	typeData = schema.appendSchemaProps(modelName, fieldName, typeData, subType)

	switch typeData {
	case "string", "text":
		return graphql.String
	case "number":
		return graphql.Int
	case "datetime":
		return graphql.DateTime
	case "boolean":
		return graphql.Boolean
	case "slice":
		slice := "_SLICE_"
		return graphql.NewList(schema.getTypeData(modelName, fieldName, *subType, &slice))
	default:

		if subType == nil {
			model := "_MODEL_"
			subType = &model
		}

		if fields, ok := schema.Models[typeData]; ok {
			if _, ok := schema.graphQLModels[typeData]; !ok {
				schema.makeModel(typeData, fields)
				schema.appendSchemaProps(modelName, fieldName, typeData, subType)
				return schema.graphQLModels[typeData]
			}
			schema.appendSchemaProps(modelName, fieldName, typeData, subType)
			return schema.graphQLModels[typeData]
		}
		return nil
	}
}
