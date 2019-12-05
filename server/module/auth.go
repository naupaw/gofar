package module

//AuthModule - mysql driver module
func AuthModule() Module {
	module := Module{}
	module.name = "auth"

	module.load = func() {
		// fmt.Println("that supposed to be loaded")
	}

	return module
}
