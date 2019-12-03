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
	"github.com/graphql-go/graphql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type GqlParam struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

func executeQuery(query string, variables map[string]interface{}, operationName string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  query,
		OperationName:  operationName,
		VariableValues: variables,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("errors: %v", result.Errors)
	}
	return result
}

func main() {
	e := echo.New()

	schemaFile, err := os.Open("schema.yaml")
	if err != nil {
		fmt.Println(err)
	}
	defer schemaFile.Close()
	byteValue, _ := ioutil.ReadAll(schemaFile)
	var mainSchema MainSchema
	yaml.Unmarshal(byteValue, &mainSchema)

	definePort := fmt.Sprintf(":%d", mainSchema.Port)

	schema := SchemaManager(mainSchema)

	e.Use(middleware.BodyLimit("2M"))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"msg": "gopur v1.0.0",
		})
	})

	e.GET(mainSchema.GraphQL.Playground, echo.WrapHandler(gqlGenHandler.Playground("GraphQL playground", mainSchema.GraphQL.Path)))

	e.POST(mainSchema.GraphQL.Path, func(c echo.Context) (err error) {
		f := new(GqlParam)
		if err = c.Bind(f); err != nil {
			return err
		}

		result := executeQuery(f.Query, f.Variables, f.OperationName, schema)
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

	fmt.Println("--------------------------------------------")
	fmt.Println("GQL Path at", "http://0.0.0.0"+definePort+mainSchema.GraphQL.Path)
	fmt.Println("Playground Start at", "http://0.0.0.0"+definePort+mainSchema.GraphQL.Playground)
	fmt.Println("--------------------------------------------")

	e.Logger.Fatal(e.Start(definePort))
}
