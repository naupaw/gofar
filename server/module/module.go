package module

import "github.com/pedox/gofar/server/model"

//Module basic module interface
type Module interface {
	ModuleName() string
	IDDataType() string
	ModuleLoaded(map[string]interface{})
	LoadedSchema()
	CreateModel(model model.Model)
}
