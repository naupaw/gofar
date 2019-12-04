package schema

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

//GraphQLConfig - GraphQL config
type GraphQLConfig struct {
	Path       string
	Playground string
}

type DatabaseConfig struct {
	Driver   string
	Username string
	Password string
	Host     string
	Port     int
	Name     string
}

//Model model lists
type Model map[string]interface{}

//Schema - schema
type Schema struct {
	Name          string           `json:"name"`
	Version       string           `json:"version"`
	GraphQL       GraphQLConfig    `json:"graphql" yaml:"graphql"`
	Models        map[string]Model `json:"models"`
	Port          int              `yaml:"port"`
	Database      DatabaseConfig   `yaml:"database"`
	queryFields   graphql.Fields
	GraphQLModels map[string]*graphql.Object
	compiedSchema graphql.Schema
}

//Initialize - initialize schema
func (schema Schema) Initialize() graphql.Schema {
	schema.GraphQLModels = map[string]*graphql.Object{}
	schema.queryFields = graphql.Fields{}
	schema.compiedSchema = graphql.Schema{}

	schema.defineSchema()

	var graphQLSchema, err = graphql.NewSchema(
		graphql.SchemaConfig{
			Query: schema.makeQuery(),
			// Mutation: mutationType,
		},
	)

	if err != nil {
		fmt.Println(err)
	}

	return graphQLSchema
}

//defineSchema - loop and define models to GraphQL Schema
func (schema Schema) defineSchema() {
	for name, fields := range schema.Models {
		if _, ok := schema.GraphQLModels[name]; ok == false {
			schema.makeModel(name, fields)
		}
	}
}

//ExecuteQuery execute GraphQL Query
func ExecuteQuery(query string, variables map[string]interface{}, operationName string, schema graphql.Schema) *graphql.Result {

	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  query,
		OperationName:  operationName,
		VariableValues: variables,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("errors: %v", result.Errors)
	}
	return result
}
