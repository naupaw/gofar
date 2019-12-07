package module

import (
	"testing"

	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

func TestDynamicStruct(t *testing.T) {

	instance := dynamicstruct.NewStruct().
		AddField("cot", 0, `json:"int"`).
		AddField("Text", "", `json:"someText"`).
		AddField("Float", 0.0, `json:"double"`).
		AddField("Boolean", false, "").
		AddField("Slice", []int{}, "").
		AddField("Anonymous", "", `json:"-"`).
		Build().
		New()

	t.Log(instance)

}
