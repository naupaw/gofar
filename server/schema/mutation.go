package schema

import "github.com/graphql-go/graphql"

func (schema Schema) makeCreateMutation(modelName string, graphQLField *graphql.Object) {
	schema.mutationFields["Create"+modelName] = &graphql.Field{
		Description: "Create New " + modelName,
		Type:        graphQLField,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		// Resolve: schema.makeResolve(graphQLField),
	}
}

func (schema Schema) makeDeleteMutation(modelName string, graphQLField *graphql.Object) {
	schema.mutationFields["Delete"+modelName] = &graphql.Field{
		Description: "Delete " + modelName,
		Type:        graphQLField,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		// Resolve: schema.makeResolve(graphQLField),
	}
}

func (schema Schema) makeEditMutation(modelName string, graphQLField *graphql.Object) {
	schema.mutationFields["Edit"+modelName] = &graphql.Field{
		Description: "Edit " + modelName,
		Type:        graphQLField,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		// Resolve: schema.makeResolve(graphQLField),
	}
}
