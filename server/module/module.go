package module

//Module basic module interface
type Module interface {
	ModuleName() string
	ModuleLoaded()
	LoadedSchema()
	CreateModel(modelName string, model map[string]interface{})
}
