package model

//Field model fields
type Field struct {
	Props map[string]string
	Type  string
}

//Model model struct
type Model struct {
	Name    string
	Fields  map[string]Field
	Options map[string]string
}
