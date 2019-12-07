package schema

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/pedox/gofar/server/model"
	"github.com/pedox/gofar/server/module"
)

//GraphQLConfig - GraphQL config
type GraphQLConfig struct {
	Path       string
	Playground string
}

//Model model lists
type Model map[string]interface{}

type GraphQLModels map[string]*graphql.Object

//Schema - schema
type Schema struct {
	Name           string                            `yaml:"name"`
	Version        string                            `yaml:"version"`
	GraphQL        GraphQLConfig                     `yaml:"graphql"`
	Modules        map[string]map[string]interface{} `yaml:"modules"`
	Models         map[string]Model                  `yaml:"models"`
	Port           int                               `yaml:"port"`
	Debug          bool                              `yaml:"debug"`
	models         map[string]model.Model
	graphQLModels  GraphQLModels
	queryFields    graphql.Fields
	mutationFields graphql.Fields
	compiedSchema  graphql.Schema
	loadedModules  map[string]module.Module
}

func (schema Schema) loadModule() {
	moduleKeys := map[string]module.Module{}
	listModules := []module.Module{
		module.NewMYSQLModule(),
	}

	for _, mod := range listModules {
		moduleKeys[mod.ModuleName()] = mod
	}

	for name, config := range schema.Modules {
		if mod, ok := moduleKeys[name]; ok {
			if schema.Debug {
				fmt.Println("> Loaded module", name)
			}
			mod.ModuleLoaded(config)
			schema.loadedModules[name] = mod
		}
	}

}

//Initialize - initialize schema
func (schema Schema) Initialize() graphql.Schema {
	schema.graphQLModels = GraphQLModels{}
	schema.queryFields = graphql.Fields{}
	schema.mutationFields = graphql.Fields{}
	schema.compiedSchema = graphql.Schema{}
	schema.loadedModules = map[string]module.Module{}
	schema.models = map[string]model.Model{}

	schema.loadModule()

	for name, fields := range schema.Models {
		if _, ok := schema.graphQLModels[name]; ok == false {
			schema.makeModel(name, fields)
		}
	}

	query, mutation := schema.makeOperation()

	var graphQLSchema, err = graphql.NewSchema(
		graphql.SchemaConfig{
			Query:    query,
			Mutation: mutation,
		},
	)

	if err != nil {
		fmt.Println(err)
	}

	return graphQLSchema
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

func (schema Schema) ModuleEvent(mdEvent func()) {
	// mdEvent()
}
