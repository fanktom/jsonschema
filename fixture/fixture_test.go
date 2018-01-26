package fixture

import (
	"encoding/json"
	"testing"
)

func TestFixtureUnmarshal(t *testing.T) {
	fs := map[string]string{
		"TestSchemaWithDefinitions":       TestSchemaWithDefinitions,
		"TestSchemaDirect":                TestSchemaDirect,
		"TestSchemaWithNestedDefinitions": TestSchemaWithNestedDefinitions,
		"TestSchemaPrimitiveTypes":        TestSchemaPrimitiveTypes,
		"TestSchemaRequiredValidation":    TestSchemaRequiredValidation,
	}
	for k, v := range fs {
		var o interface{}
		err := json.Unmarshal([]byte(v), &o)
		if err != nil {
			t.Fatalf("fixture %s does not unmarshal: %s", k, err)
		}
	}
}
