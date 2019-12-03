package main

import (
	"github.com/graphql-go/graphql"
)

//MainSchema - main application Schema
type GraphQLConfig struct {
	Path       string
	Playground string
}

type Model map[string]interface{}

type MainSchema struct {
	Name    string           `json:"name"`
	Version string           `json:"version"`
	GraphQL GraphQLConfig    `json:"version" yaml:"graphql"`
	Models  map[string]Model `json:"models"`
	Port    int              `yaml:"port"`
}

var dataTypes map[string]*graphql.Object
var kindType = map[string]string{}

//ModelLists lists of models
var ModelLists map[string]Model

//SchemaManager - initialize schema
func SchemaManager(mainSchema MainSchema) graphql.Schema {
	dataTypes = map[string]*graphql.Object{}
	ModelLists = mainSchema.Models
	defineSchema()
	var schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query: makeQuery(mainSchema),
			// Mutation: mutationType,
		},
	)

	return schema
}
