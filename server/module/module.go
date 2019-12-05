package module

import (
	"log"

	schm "github.com/pedox/gofar/server/schema"
)

//Module the module struct
type Module struct {
	name            string
	load            func()
	afterModel      *func(schema schm.Schema)
	queryExecute    *func(schema schm.Schema)
	mutationExecute *func(schema schm.Schema)
}

//Modules Module collections
type Modules map[string]Module

func listModule() []Module {
	// Append module here
	return []Module{
		DatabaseMYSQL(),
		DatabaseMongoDB(),
		AuthModule(),
	}
}

func (m Modules) appendModule(module Module) {
	m[module.name] = module
}

//Load - load modules
func (m Module) Load() Module {
	m.load()
	log.Println(">>", "module", m.name, "loaded")
	return m
}

//LoadModule - load all module that you neededs
func LoadModule(names []string) Modules {
	log.Println(">>", "Initialize Module")

	moduleKey := Modules{}
	activeModules := Modules{}
	for _, m := range listModule() {
		moduleKey.appendModule(m)
	}

	for _, name := range names {
		if _, ok := moduleKey[name]; ok {
			activeModules.appendModule(moduleKey[name].Load())
		}
	}

	return activeModules
}
