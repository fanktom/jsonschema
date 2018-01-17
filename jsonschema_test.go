package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const withDefinitions = `
{
	"definitions": {
    "movie": {
      "type": "object",
      "required": ["id", "name"],
      "properties": {
        "id": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "year": {
          "type": "integer"
        },
				"actor": {
					"type": "object",
					"properties": {
        		"id": {
        		  "type": "string"
        		},
        		"name": {
        		  "type": "string"
        		}
					}
				},
				"categories": {
					"$ref": "#/definitions/categories"
				}
      }
    },
		"categories": {
			"type": "array",
			"items": {
				"type": "string"
			}
		}
	}
}
`

const directSchema = `
{
  "type": "object",
  "required": ["id", "name"],
  "properties": {
    "id": {
      "type": "string"
    },
    "name": {
      "type": "string"
    },
    "year": {
      "type": "integer"
    },
		"actor": {
			"type": "object",
			"properties": {
    		"id": {
    		  "type": "string"
    		},
    		"name": {
    		  "type": "string"
    		}
			}
		},
		"categories": {
			"$ref": "#/definitions/categories"
		}
  }
}
`

const withNestedDefinitions = `
{
	"definitions": {
    "movie": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string"
        },
        "name": {
          "type": "string"
        }
      },
			"definitions": {
  		  "actor": {
  		    "type": "object",
  		    "properties": {
  		      "id": {
  		        "type": "string"
  		      },
  		      "name": {
  		        "type": "string"
  		      }
  		    }
  		  }
			}
    }
	}
}
`
const primitiveTypes = `
{
	"definitions": {
		"null": { "type": "null" },
		"boolean": { "type": "boolean" },
		"object": { "type": "object" },
		"array": { "type": "array" },
		"number": { "type": "number" },
		"integer": { "type": "integer" },
		"string": { "type": "string" }
	}
}
`

const requiredValidationSchema = `
{
	"definitions": {
    "movie": {
      "type": "object",
      "required": ["id", "actors"],
      "properties": {
        "id": {
          "type": "string"
        },
				"actors": {
					"type": "array",
					"items": {
						"$ref": "#/definitions/actor"
					}
				}
      }
    },
    "actor": {
      "type": "object",
      "required": ["name", "location"],
      "properties": {
        "name": {
          "type": "string"
        },
				"location": {
					"$ref": "#/definitions/location"
				}
			}
    },
    "location": {
      "type": "object",
      "required": ["name"],
			"properties": {
				"name": {
					"type": "string"
				}
			}
    }
	}
}
`

func TestWithDefinitions(t *testing.T) {
	idx, err := Parse([]byte(withDefinitions))
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
	idx, err := Parse([]byte(directSchema))
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
	idx, err := Parse([]byte(withNestedDefinitions))
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

func TestGenerateWithDefinitions(t *testing.T) {
	idx, err := Parse([]byte(withDefinitions))
	if err != nil {
		panic(err)
	}

	actor := ""
	actor += "type Actor struct {\n"
	actor += "	ID   *string `json:\"id,omitempty\"`\n"
	actor += "	Name *string `json:\"name,omitempty\"`\n"
	actor += "}\n"

	categories := "type Categories []string\n"

	movie := ""
	movie += "type Movie struct {\n"
	movie += "	Actor      *Actor      `json:\"actor,omitempty\"`\n"
	movie += "	Categories *Categories `json:\"categories,omitempty\"`\n"
	movie += "	ID         *string     `json:\"id,omitempty\"`\n"
	movie += "	Name       *string     `json:\"name,omitempty\"`\n"
	movie += "	Year       *int        `json:\"year,omitempty\"`\n"
	movie += "}\n"

	table := map[string]string{
		"#/definitions/movie":                                  movie,
		"#/definitions/movie/properties/id":                    "",
		"#/definitions/movie/properties/name":                  "",
		"#/definitions/movie/properties/year":                  "",
		"#/definitions/movie/properties/actor":                 actor,
		"#/definitions/movie/properties/actor/properties/id":   "",
		"#/definitions/movie/properties/actor/properties/name": "",
		"#/definitions/movie/properties/categories":            "",
		"#/definitions/categories":                             categories,
		"#/definitions/categories/items":                       "",
	}
	for p, g := range table {
		s := (*idx)[p]
		gos, err := s.generateGoType(idx)
		if err != nil {
			t.Fatal(err)
		}
		if string(gos) != g {
			t.Fatalf("struct of %v should be '%v' but is '%s'", p, g, gos)
		}
	}
}

func TestGenerateGoTypes(t *testing.T) {
	idx, err := Parse([]byte(withDefinitions))
	if err != nil {
		panic(err)
	}

	gts, err := generateGoTypes(idx)
	if err != nil {
		t.Fatal(err)
	}

	out := "\n"
	out += "type Actor struct {\n"
	out += "	ID   *string `json:\"id,omitempty\"`\n"
	out += "	Name *string `json:\"name,omitempty\"`\n"
	out += "}\n"
	out += "\n"
	out += "type Categories []string\n"
	out += "\n"
	out += "type Movie struct {\n"
	out += "	Actor      *Actor      `json:\"actor,omitempty\"`\n"
	out += "	Categories *Categories `json:\"categories,omitempty\"`\n"
	out += "	ID         *string     `json:\"id,omitempty\"`\n"
	out += "	Name       *string     `json:\"name,omitempty\"`\n"
	out += "	Year       *int        `json:\"year,omitempty\"`\n"
	out += "}\n"
	out += "\n"

	if string(gts) != out {
		t.Fatalf("invalid output '%s' should be '%v'", gts, out)
	}
}

func TestGenerateWithPrimitiveTypes(t *testing.T) {
	idx, err := Parse([]byte(primitiveTypes))
	if err != nil {
		panic(err)
	}

	table := map[string]string{
		"#/definitions/null":    "",
		"#/definitions/boolean": "Boolean *bool `json:\"boolean,omitempty\"`",
		"#/definitions/object":  "Object *Object `json:\"object,omitempty\"`",
		"#/definitions/array":   "Array *Array `json:\"array,omitempty\"`",
		"#/definitions/number":  "Number *float64 `json:\"number,omitempty\"`",
		"#/definitions/integer": "Integer *int `json:\"integer,omitempty\"`",
		"#/definitions/string":  "String *string `json:\"string,omitempty\"`",
	}
	for p, r := range table {
		s := (*idx)[p]
		ref := s.generateGoRef(idx)
		if ref != r {
			t.Fatalf("ref type of %v should be '%v' but is '%s'", p, r, ref)
		}
	}
}

func TestGenerateGoTypeValidateFuncWithDefinitions(t *testing.T) {
	table := []struct {
		RawSchema string
		Pointer   string
		Error     string
		Code      string
	}{
		{
			withDefinitions,
			"#/definitions/movie", "invalid movie: missing id", `
				m := Movie{}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			withDefinitions,
			"#/definitions/movie", "invalid movie: missing name", `
				m := Movie{ID: newString("foo")}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			withDefinitions,
			"#/definitions/movie", "", `
				m := Movie{ID: newString("foo"), Name: newString("bar")}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			requiredValidationSchema,
			"#/definitions/movie", "invalid movie: missing id", `
				m := Movie{}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			requiredValidationSchema,
			"#/definitions/movie", "invalid movie: missing actors", `
				m := Movie{ID: newString("foo")}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			requiredValidationSchema,
			"#/definitions/movie", "", `
				m := Movie{ID: newString("foo"), Actors: &Actors{}}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			requiredValidationSchema,
			"#/definitions/movie", "invalid actor: missing name", `
				m := Movie{ID: newString("foo"), Actors: &Actors{
					Actor{},
				}}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			requiredValidationSchema,
			"#/definitions/movie", "invalid actor: missing location", `
				m := Movie{ID: newString("foo"), Actors: &Actors{
					Actor{Name: newString("John Snow")},
				}}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			requiredValidationSchema,
			"#/definitions/movie", "invalid location: missing name", `
				m := Movie{ID: newString("foo"), Actors: &Actors{
					Actor{Name: newString("John Snow"), Location: &Location{}},
				}}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
		{
			requiredValidationSchema,
			"#/definitions/movie", "", `
				m := Movie{ID: newString("foo"), Actors: &Actors{
					Actor{Name: newString("John Snow"), Location: &Location{
						Name: newString("Winterfell"),
					}},
				}}
				err := m.Validate()
				if err != nil {
					fmt.Print(err)
				}
			`,
		},
	}
	for _, ts := range table {
		src, err := GenerateGoSrcPackage([]byte(ts.RawSchema), "main")
		if err != nil {
			t.Fatal(err)
		}

		// inject fmt (only needed for test program runs)
		srcs := strings.Replace(string(src), "import (", "import (\n\t\"fmt\"", 1)

		w := bytes.NewBufferString(srcs)
		fmt.Fprintf(w, `
func main() {
%v
}
`, ts.Code)

		out, err := compileAndRun(w.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		if out != ts.Error {
			t.Fatalf("%v should have produced '%v', but produced '%v'", ts, ts.Error, out)
		}
	}
}

func TestGenerateNewInstanceJSON(t *testing.T) {
	idx, err := Parse([]byte(withDefinitions))
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
