package module

import (
	"encoding/json"
	"fmt"
)

//MysqlModule mysql module
type MysqlModule struct {
	hostname string
	username string
	password string
}

//NewMYSQLModule - mysql driver module
func NewMYSQLModule() Module {
	return &MysqlModule{}
}

//ModuleName module name
func (m *MysqlModule) ModuleName() string {
	return "mysql"
}

func (m *MysqlModule) ModuleLoaded() {
	// fmt.Println("howddy !")
}

func (m *MysqlModule) LoadedSchema() {

}

func (m *MysqlModule) CreateModel(modelName string, model map[string]interface{}) {
	josh, err := json.MarshalIndent(model, " ", "  ")
	if err != nil {
		fmt.Println("error model", modelName, err)
	} else {
		fmt.Println("model", modelName, string(josh))
	}
}
