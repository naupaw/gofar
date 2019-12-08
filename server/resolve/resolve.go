package resolve

import "github.com/graphql-go/graphql"

type ResolveType int

const (
	Primitive ResolveType = iota
	Relation
)

type ResolveParamResult map[string]interface{}

type PreResolveParamCallback func(PreResolveParam) (ResolveParamResult, error)

//Resolve query resolver
type Resolve struct {
	FieldName  string
	Param      graphql.ResolveParams
	FieldTypes map[string]ResolveType
	Fields     map[string]interface{}
}

type PreResolveParam struct {
	FieldName       string
	Param           graphql.ResolveParams
	Fields          map[string]interface{}
	ParentFieldName *string
	ParentFields    *map[string]interface{}
}
