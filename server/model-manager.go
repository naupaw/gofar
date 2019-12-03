package main

import (
	"github.com/graphql-go/graphql"
)

func getTypeData(typeData string, subType *string, depth int) graphql.Output {
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
