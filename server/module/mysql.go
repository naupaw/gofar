package module

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	pluralize "github.com/gertd/go-pluralize"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
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
	db *sql.DB
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

	db, err := sql.Open("mysql", mysqlConn)
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

func getType(typeData string) string {
	switch typeData {
	case "string":
		return "VARCHAR(255)"
	case "number":
		return "INT"
	case "TIMESTAMP", "TINYINT", "DATE":
		return typeData
	default:
		return ""
	}
}

func createInsertStatement(statements []string, fieldName string, typeDate string) []string {
	statements = append(
		statements,
		fmt.Sprintf(" `%s` %s ", strcase.ToSnake(fieldName), typeDate),
	)
	return statements
}

func (m *MysqlModule) CreateModel(model model.Model) {
	pluralize := pluralize.NewClient()
	tableName := pluralize.Plural(strcase.ToLowerCamel(model.Name))

	insertStatement := []string{}

	insertStatement = createInsertStatement(insertStatement, "id", getType(m.IDDataType()))

	for name, field := range model.Fields {
		// tag := ""
		// if val, ok := field.Props["types"]; ok {
		// 	tag = fmt.Sprintf(`gorm:"%s"`, val)
		// }

		tx, err := m.db.Begin()
		if err != nil {
			fmt.Println(err)
		}
		defer tx.Rollback()

		stmt, err := tx.Prepare(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName))
		if err != nil {
			fmt.Println(err)
		}
		defer stmt.Close()
		stmt.Exec()

		if err := tx.Commit(); err != nil {
			fmt.Println(err)
		}

		typeDat := getType(field.Type)

		if name != "ID" {
			if rel, ok := field.Props["relation"]; ok {
				if rel == "hasOne" {
					insertStatement = createInsertStatement(insertStatement, name+"_id", getType(m.IDDataType()))
				}
			} else {
				if typeDat != "" {
					insertStatement = createInsertStatement(insertStatement, name, typeDat)
				}
			}
		}
	}

	insertStatement = createInsertStatement(insertStatement, "created_at", "TIMESTAMP")
	insertStatement = createInsertStatement(insertStatement, "updated_at", "TIMESTAMP")

	sqlInsert := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS `%s` ( %s ) ENGINE=InnoDB DEFAULT CHARSET=latin1",
		tableName, strings.Join(insertStatement, ","),
	)

	// fmt.Println(sqlInsert)

	tx, err := m.db.Begin()
	if err != nil {
		fmt.Println(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(sqlInsert)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()
	stmt.Exec()

	if err := tx.Commit(); err != nil {
		fmt.Println(err)
	}
}
