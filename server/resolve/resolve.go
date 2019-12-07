package resolve

import "github.com/graphql-go/graphql"

type ResolveType int

const (
	Primitive ResolveType = iota
	Relation
)

//Resolve query resolver
type Resolve struct {
	FieldName  string
	Param      graphql.ResolveParams
	FieldTypes map[string]ResolveType
	Fields     map[string]interface{}
}
