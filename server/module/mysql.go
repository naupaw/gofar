package module

import (
	"fmt"

	schm "github.com/pedox/gofar/server/schema"
)

//DatabaseMYSQL - mysql driver module
func DatabaseMYSQL() Module {
	module := Module{}
	module.name = "mysql"

	module.load = func() {
		fmt.Println("Mysql Initialized")
	}

	afterModel := func(schema schm.Schema) {
	}

	queryExecute := func(schema schm.Schema) {
	}

	module.afterModel = &afterModel
	module.queryExecute = &queryExecute

	return module
}
