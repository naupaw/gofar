package module

import (
	"github.com/pedox/gofar/server/model"
	"github.com/pedox/gofar/server/resolve"
)

//Module basic module interface
type Module interface {
	ModuleName() string
	IDDataType() string
	ModuleLoaded(map[string]interface{})
	LoadedSchema()
	CreateModel(model model.Model)
	Query(resolve resolve.Resolve) map[string]interface{}
	Create(resolve resolve.Resolve) map[string]interface{}
	Update(resolve resolve.Resolve) map[string]interface{}
	Delete(resolve resolve.Resolve) map[string]interface{}
}
