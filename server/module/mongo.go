package module

//DatabaseMongoDB - mysql driver module
func DatabaseMongoDB() Module {
	module := Module{}
	module.name = "mongodb"

	module.load = func() {
		// fmt.Println("that supposed to be loaded")
	}

	return module
}
