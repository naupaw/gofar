package main

import (
	"log"
	"reflect"

	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
)

func makeModel(name string, fields map[interface{}]interface{}, depth int) {
	objectFields := graphql.Fields{}
	gqlType := graphql.NewObject(
		graphql.ObjectConfig{
			Name:   name,
			Fields: objectFields,
		},
	)
	dataTypes[name] = gqlType

	objectFields["ID"] = &graphql.Field{
		Type: getTypeData("string", nil, depth+1),
	}

	for field, typeData := range fields {
		fieldName := strcase.ToLowerCamel(field.(string))
		if field.(string)[0:2] != "__" {
			v := reflect.ValueOf(typeData)
			switch v.Kind() {
			case reflect.Slice:
				s := reflect.ValueOf(typeData)
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

	log.Println(":: Initialize", name)
}

func makeFieldList(name string, data graphql.Output) graphql.Output {
	objectFields := graphql.Fields{
		"page": &graphql.Field{
			Type: graphql.Int,
		},
		"lastPage": &graphql.Field{
			Type: graphql.Int,
		},
		"total": &graphql.Field{
			Type: graphql.Int,
		},
		"perPage": &graphql.Field{
			Type: graphql.Int,
		},
		"data": &graphql.Field{
			Type: graphql.NewList(data),
		},
	}
	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   strcase.ToCamel(name + "List"),
			Fields: objectFields,
		},
	)
}

func makeCollection(col graphql.Fields) graphql.Fields {
	for collection, fields := range dataTypes {
		collectionName := strcase.ToLowerCamel(collection)
		col[collectionName] = &graphql.Field{
			Description: collectionName + " Single data",
			Type:        fields,
			Args: graphql.FieldConfigArgument{
				"ID": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (res interface{}, err error) {
				return map[string]interface{}{
					"ID": "jsad781n2k3jncz8x-asdjnui13hn-123unc9aus9d",
				}, nil
			},
		}

		col[collectionName+"List"] = &graphql.Field{
			Description: collectionName + " Datasets",
			Type:        makeFieldList(collection, fields),
			Args: graphql.FieldConfigArgument{
				"page": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"perPage": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(p graphql.ResolveParams) (res interface{}, err error) {
				dataList := []map[string]interface{}{}

				dataList = append(dataList, map[string]interface{}{
					"title":   "Judul",
					"content": "konten di sini",
				})

				page, _ := p.Args["page"]

				return map[string]interface{}{
					"page":     page,
					"lastPage": 1,
					"total":    10,
					"perPage":  10,
					"data":     dataList,
				}, nil
			},
		}

	}

	return col
}

func makeQuery(schema MainSchema) *graphql.Object {

	aboutType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "About",
			Fields: graphql.Fields{
				"version": &graphql.Field{
					Type: graphql.String,
				},
				"name": &graphql.Field{
					Type: graphql.String,
				},
			},
		},
	)

	collections := graphql.Fields{
		"about": &graphql.Field{
			Description: "Tentang aplikasi ini",
			Type:        aboutType,
			Resolve: func(p graphql.ResolveParams) (res interface{}, err error) {
				return map[string]interface{}{
					"version": schema.Version,
					"name":    schema.Name,
				}, nil
			},
		},
	}

	collections = makeCollection(collections)

	return graphql.NewObject(
		graphql.ObjectConfig{
			Name:   "Query",
			Fields: collections,
		},
	)
}

func defineSchema() {
	for name, fields := range models {
		if _, ok := dataTypes[name]; ok == false {
			makeModel(name, fields.(map[interface{}]interface{}), 1)
		}
	}
}
