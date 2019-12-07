package module

import (
	"fmt"
	"time"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
	"github.com/pedox/gofar/server/model"
)

type BaseModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

//MysqlModule mysql module
type MysqlModule struct {
	db *gorm.DB
}

//NewMYSQLModule - mysql driver module
func NewMYSQLModule() Module {
	return &MysqlModule{}
}

//ModuleName module name
func (m *MysqlModule) ModuleName() string {
	return "mysql"
}

func (m *MysqlModule) ModuleLoaded(config map[string]interface{}) {

	mysqlConn := fmt.Sprintf(
		"%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config["username"],
		config["password"],
		config["host"],
		config["database"],
	)

	db, err := gorm.Open("mysql", mysqlConn)
	if err != nil {
		panic(err)
	}
	m.db = db
}

func (m *MysqlModule) LoadedSchema() {

}

func (m *MysqlModule) IDDataType() string {
	return "string"
}

func getType(typeData string) interface{} {
	switch typeData {
	case "string":
		return ""
	case "number":
		return 0
	default:
		return nil
	}
}

func (m *MysqlModule) CreateModel(model model.Model) {
	pluralize := pluralize.NewClient()
	tableName := pluralize.Plural(strcase.ToLowerCamel(model.Name))
	instance := dynamicstruct.ExtendStruct(BaseModel{})

	// instance.AddField("ID", m.IDDataType(), `gorm:"primary_key"`)

	for name, field := range model.Fields {
		tag := ""
		if val, ok := field.Props["types"]; ok {
			tag = fmt.Sprintf(`gorm:"%s"`, val)
		}

		m.db.DropTableIfExists(tableName)

		typeDat := getType(field.Type)

		if name != "ID" {
			if rel, ok := field.Props["relation"]; ok {
				if rel == "hasOne" {
					instance.AddField(
						strcase.ToCamel(name+"_id"),
						m.IDDataType(),
						tag,
					)
				}
			} else {
				if typeDat != nil {
					instance.AddField(strcase.ToCamel(name), getType(field.Type), tag)
				}
			}
		}
	}

	modelTypes := instance.Build().New()

	m.
		db.
		Debug().
		Table(tableName).CreateTable(modelTypes)

	// fmt.Println(instance)

	// m.db.AutoMigrate()
	// josh, err := json.MarshalIndent(model, " ", "  ")
	// if err != nil {
	// 	fmt.Println("error model", err)
	// } else {
	// 	fmt.Println("model", string(josh))
	// }
}
