package module

import (
	"database/sql"
	"fmt"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
	"github.com/pedox/gofar/server/model"
	"github.com/pedox/gofar/server/resolve"
)

//MysqlModule mysql module
type MysqlModule struct {
	db     *sql.DB
	config map[string]interface{}
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
	m.config = config
}

func (m *MysqlModule) LoadedSchema() {

}

func (m *MysqlModule) IDDataType() string {
	return "bigint"
}

func getType(typeData string) string {
	switch typeData {
	case "string":
		return "VARCHAR(255)"
	case "number":
		return "INT"
	case "bigint":
		return "BIGINT"
	case "boolean":
		return "TINYINT(1) DEFAULT 0"
	case "text":
		return "TEXT"
	case "TIMESTAMP", "TINYINT", "DATE":
		return typeData
	default:
		return ""
	}
}

func createInsertStatement(statements []string, fieldName string, typeDate string, extra *string) []string {

	fieldQuery := fmt.Sprintf(" `%s` %s ", strcase.ToSnake(fieldName), typeDate)

	if extra != nil {
		fieldQuery += *extra
	}

	statements = append(
		statements,
		fieldQuery,
	)
	return statements
}

//extractDBExtraWithValue something like db:"default=1"
func extractDBExtraWithValue(field string) string {
	v := strings.Split(field, "=")
	switch v[0] {
	case "default":
		return "DEFAULT " + v[1]
	}
	return ""
}

//extractDBExtra something like db:"unique;primary_key"
func extractDBExtra(field model.Field) string {
	dbExtra := ""
	extraRule := map[string]bool{}

	if val, ok := field.Props["validate"]; ok {
		for _, v := range strings.Split(val, ",") {
			if _, ok := extraRule[v]; !ok {
				switch v {
				case "unique":
					extraRule[v] = true
					dbExtra += " UNIQUE"
					break
				case "required":
					extraRule[v] = true
					dbExtra += " NOT NULL"
					break
				default:
					dbExtra += extractDBExtraWithValue(v)
				}
			}
		}
	}
	return dbExtra
}

func (m *MysqlModule) CreateModel(model model.Model) {
	create := false

	if create {
		pluralize := pluralize.NewClient()
		tableName := pluralize.Plural(strcase.ToLowerCamel(model.Name))

		insertStatement := []string{}

		primaryKey := "UNSIGNED NOT NULL AUTO_INCREMENT UNIQUE PRIMARY KEY"
		insertStatement = createInsertStatement(insertStatement, "id", "BIGINT", &primaryKey)

		for name, field := range model.Fields {
			dbExtra := extractDBExtra(field)

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

			if name != "id" {
				if rel, ok := field.Props["relation"]; ok {
					if rel == "hasOne" {
						insertStatement = createInsertStatement(insertStatement, name+"_id", getType(m.IDDataType()), &dbExtra)
					}
				} else {
					if typeDat != "" {
						insertStatement = createInsertStatement(insertStatement, name, typeDat, &dbExtra)
					}
				}
			}
		}

		timestampExtra := "DEFAULT CURRENT_TIMESTAMP"
		insertStatement = createInsertStatement(insertStatement, "created_at", "TIMESTAMP", &timestampExtra)
		insertStatement = createInsertStatement(insertStatement, "updated_at", "TIMESTAMP", &timestampExtra)

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
}

func (m *MysqlModule) Query(res resolve.Resolve) map[string]interface{} {
	pluralize := pluralize.NewClient()
	tableName := pluralize.Plural(strcase.ToLowerCamel(res.FieldName))
	field := []string{}

	id, _ := res.Param.Args["id"].(int)

	for name, kind := range res.FieldTypes {
		if kind == resolve.Primitive {
			field = append(field, name)
		}
	}

	fmt.Println("ID", id)

	row := m.db.QueryRow(
		fmt.Sprintf(
			"SELECT %s FROM %s WHERE id = ? LIMIT 1",
			strings.Join(field, ", "),
			tableName,
		),
		id,
	)

	columns := make([]interface{}, len(field))
	columnPointers := make([]interface{}, len(field))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	if err := row.Scan(columnPointers...); err != nil {
		fmt.Println("err", err)
		return map[string]interface{}{}
	}

	mx := make(map[string]interface{})
	for i, name := range field {
		val := columnPointers[i].(*interface{})
		mx[name] = *val
	}

	// a := sqlRes["test".(string)]

	fmt.Println("SQLRES", mx)

	//Dummy result
	res.Fields["username"] = "pedox"
	res.Fields["password"] = "secret"
	res.Fields["user_id"] = "11-22-33-44"

	return res.Fields
}

func (m *MysqlModule) Create(res resolve.Resolve) map[string]interface{} {
	pluralize := pluralize.NewClient()
	tableName := pluralize.Plural(strcase.ToLowerCamel(res.FieldName))
	fields := []string{}
	valuesQ := []string{}
	values := []interface{}{}

	tx, err := m.db.Begin()
	if err != nil {
		return nil
	}
	defer tx.Rollback()

	for name, val := range res.Param.Args {
		fields = append(fields, "`"+name+"`")
		values = append(values, val)
		valuesQ = append(valuesQ, "?")
	}

	prepare := fmt.Sprintf(
		"INSERT INTO %s ( %s ) VALUES ( %s )",
		tableName,
		strings.Join(fields, ", "),
		strings.Join(valuesQ, ", "),
	)

	fmt.Println(prepare)

	stmt, err := tx.Prepare(prepare)
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	exres, err := stmt.Exec(values...)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(exres.LastInsertId())

	if err := tx.Commit(); err != nil {
		fmt.Println(err)
	}

	return res.Fields
}

func (m *MysqlModule) Update(res resolve.Resolve) map[string]interface{} {
	return nil
}

func (m *MysqlModule) Delete(res resolve.Resolve) map[string]interface{} {
	return nil
}
