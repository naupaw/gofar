package main

import (
	"github.com/graphql-go/graphql"
)

//MainSchema - main application Schema
type GraphQLConfig struct {
	Path       string
	Playground string
}
type MainSchema struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	GraphQL     GraphQLConfig          `json:"version" yaml:"graphql"`
	Collections map[string]interface{} `json:"collections"`
	Port        int                    `yaml:"port"`
}

var dataTypes map[string]*graphql.Object
var kindType = map[string]string{}

//Models models
var Models map[string]interface{}

//SchemaManager - initialize schema
func SchemaManager(mainSchema MainSchema) graphql.Schema {
	dataTypes = map[string]*graphql.Object{}

	Models = mainSchema.Collections

	defineSchema()

	q := makeQuery(mainSchema)

	var schema, _ = graphql.NewSchema(
		graphql.SchemaConfig{
			Query: q,
			// Mutation: mutationType,
		},
	)

	return schema
}
