package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	gqlGenHandler "github.com/99designs/gqlgen/handler"
	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	gschema "github.com/pedox/gofar/server/schema"
)

type GqlParam struct {
	Query         string                 `json:"query" yaml:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

var banner = `
 _____         __
/ ____|      /  _|
| |  __  ___ | |_ ____ __
| | |_ |/ _ \|  _/ _| __|
| |__| | (_) | || (_|| |
 \_____|\___/|_| \__,|_| v1.0.0

`

func main() {
	e := echo.New()
	e.HideBanner = true

	schemaFile, err := os.Open("schema.yaml")
	if err != nil {
		fmt.Println(err)
	}
	defer schemaFile.Close()
	byteValue, _ := ioutil.ReadAll(schemaFile)
	var schema gschema.Schema
	yaml.Unmarshal(byteValue, &schema)

	definePort := fmt.Sprintf(":%d", schema.Port)
	gqlSchema := schema.Initialize()

	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"msg": "gopur v1.0.0",
		})
	})

	e.GET(schema.GraphQL.Playground, echo.WrapHandler(gqlGenHandler.Playground("GraphQL playground", schema.GraphQL.Path)))

	e.POST(schema.GraphQL.Path, func(c echo.Context) (err error) {
		f := new(GqlParam)
		if err = c.Bind(f); err != nil {
			return err
		}

		result := gschema.ExecuteQuery(f.Query, f.Variables, f.OperationName, gqlSchema)
		return c.JSON(http.StatusOK, result)
	})

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		report, ok := err.(*echo.HTTPError)
		if !ok {
			report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		c.Logger().Error(report)
		c.JSON(report.Code, conjson.NewMarshaler(report, transform.ConventionalKeys()))
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://www.graphqlbin.com"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	fmt.Println(banner)
	fmt.Println("--------------------------------------------")
	fmt.Println("Using Driver MongoDB")
	fmt.Println("--------------------------------------------")
	fmt.Println("GQL Path at", "http://0.0.0.0"+definePort+schema.GraphQL.Path)
	fmt.Println("Playground Start at", "http://0.0.0.0"+definePort+schema.GraphQL.Playground)
	fmt.Println("--------------------------------------------")

	e.Logger.Fatal(e.Start(definePort))
}
