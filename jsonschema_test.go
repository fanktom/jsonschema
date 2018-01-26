package jsonschema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/tfkhsr/jsonschema/fixture"
)

func TestWithDefinitions(t *testing.T) {
	idx, err := Parse([]byte(fixture.TestSchemaWithDefinitions))
	if err != nil {
		panic(err)
	}

	table := map[string]string{
		"#/definitions/movie":                                  "object",
		"#/definitions/movie/properties/id":                    "string",
		"#/definitions/movie/properties/name":                  "string",
		"#/definitions/movie/properties/year":                  "integer",
		"#/definitions/movie/properties/actor":                 "object",
		"#/definitions/movie/properties/actor/properties/id":   "string",
		"#/definitions/movie/properties/actor/properties/name": "string",
		"#/definitions/movie/properties/categories":            "ref",
		"#/definitions/categories":                             "array",
		"#/definitions/categories/items":                       "string",
	}
	for p, tp := range table {
		s, ok := (*idx)[p]
		if !ok {
			t.Fatalf("index does not contain pointer %v", p)
		}
		if s.Type != tp {
			t.Fatalf("type of schema with pointer %v is not %v but %v", p, tp, s.Type)
		}
		if s.Pointer != p {
			t.Fatalf("pointer of schema should be %v but is %v", p, s.Pointer)
		}
	}
}

func TestDirectSchema(t *testing.T) {
	idx, err := Parse([]byte(fixture.TestSchemaDirect))
	if err != nil {
		panic(err)
	}

	table := map[string]string{
		"#/properties/id":                    "string",
		"#/properties/name":                  "string",
		"#/properties/year":                  "integer",
		"#/properties/actor":                 "object",
		"#/properties/actor/properties/id":   "string",
		"#/properties/actor/properties/name": "string",
		"#/properties/categories":            "ref",
	}
	for p, tp := range table {
		s, ok := (*idx)[p]
		if !ok {
			t.Fatalf("index does not contain pointer %v", p)
		}
		if s.Type != tp {
			t.Fatalf("type of schema with pointer %v is not %v but %v", p, tp, s.Type)
		}
		if s.Pointer != p {
			t.Fatalf("pointer of schema should be %v but is %v", p, s.Pointer)
		}
	}
}

func TestWithNestedDefinitions(t *testing.T) {
	idx, err := Parse([]byte(fixture.TestSchemaWithNestedDefinitions))
	if err != nil {
		panic(err)
	}

	table := map[string]string{
		"#/definitions/movie":                                   "object",
		"#/definitions/movie/properties/id":                     "string",
		"#/definitions/movie/properties/name":                   "string",
		"#/definitions/movie/definitions/actor":                 "object",
		"#/definitions/movie/definitions/actor/properties/id":   "string",
		"#/definitions/movie/definitions/actor/properties/name": "string",
	}
	for p, tp := range table {
		s, ok := (*idx)[p]
		if !ok {
			t.Fatalf("index does not contain pointer %v", p)
		}
		if s.Type != tp {
			t.Fatalf("type of schema with pointer %v is not %v but %v", p, tp, s.Type)
		}
		if s.Pointer != p {
			t.Fatalf("pointer of schema should be %v but is %v", p, s.Pointer)
		}
	}
}

func TestGenerateNewInstanceJSON(t *testing.T) {
	idx, err := Parse([]byte(fixture.TestSchemaWithDefinitions))
	if err != nil {
		panic(err)
	}

	s := (*idx)["#/definitions/movie"]
	inst, err := s.NewInstance(idx)
	if err != nil {
		t.Fatal(err)
	}
	jsn, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	o := `{
  "actor": {
    "id": "string",
    "name": "string"
  },
  "categories": [
    "string"
  ],
  "id": "string",
  "name": "string",
  "year": 42
}`
	if string(jsn) != o {
		t.Fatalf("sample json should be '%s' but is '%s'", o, string(jsn))
	}
}

// compiles the given code, runs it and returns the response
func compileAndRun(code []byte) (string, error) {
	const name = "tmp"
	os.RemoveAll(name)
	err := os.Mkdir(name, 0700)
	if err != nil {
		return "", err
	}

	// write src
	err = ioutil.WriteFile(name+"/main.go", code, 0700)
	if err != nil {
		return "", err
	}

	// compile
	cmd := exec.Command("sh", "-c", "go build")
	cmd.Dir = name
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, out)
	}

	// execute
	cmd = exec.Command("sh", "-c", "./"+name)
	cmd.Dir = name
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, out)
	}

	os.RemoveAll(name)
	return string(out), nil
}
