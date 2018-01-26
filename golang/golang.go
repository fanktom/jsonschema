/*
Package golang generates go types including validations.

Parse a jsonschema.Index into a "main" package including imports, e.g. for saving to a file:

	schema := []byte(`...`)
	idx, err := jsonschema.Parse(schema)
	if err != nil {
		panic(err)
	}

	src, err := PackageSrc(idx, "main")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", src)

A schema definition named user:

	schema := []byte(`{
	  "definitions": {
	    "user": {
	      "type": "object",
	      "properties": {
	        "id": { "type": "string" },
	        "name": { "type": "string" },
	      },
	      "required": ["id"]
	    }
	}`)

	idx, err := jsonschema.Parse(schema)
	if err != nil {
		panic(err)
	}

	src, err := Src(idx, "main")
	if err != nil {
		panic(err)
	}

Results in a User type with a Validate method:

	type User struct {
		ID    *string `json:"id,omitempty"`
		Name  *string `json:"name,omitempty"`
	}

	func (t *User) Validate() error {
		if t.ID == nil {
			return errors.New("invalid user: missing id")
		}

		return nil
	}
*/
package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"regexp"
	"sort"
	"strings"

	"github.com/tfkhsr/jsonschema"
)

// Generates go src from an jsonschema.Index without imports and package
func Src(idx *jsonschema.Index) ([]byte, error) {
	typ, err := generateGoTypes(idx)
	if err != nil {
		return nil, err
	}

	vf, err := generateGoTypesValidateFuncs(idx)
	if err != nil {
		return nil, err
	}

	pt, err := generateGoPrimitiveTypesNewFuncs()
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	fmt.Fprintf(w, "%s", typ)
	fmt.Fprintf(w, "%s", vf)
	fmt.Fprintf(w, "%s", pt)

	return format.Source(w.Bytes())
}

// Generates go src for a package including imports and package
func PackageSrc(idx *jsonschema.Index, pack string) ([]byte, error) {
	src, err := Src(idx)
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	fmt.Fprintf(w, `package %v

import (
`, pack)
	for _, i := range Imports(src) {
		fmt.Fprintf(w, "\t\"%s\"\n", i)
	}
	fmt.Fprintf(w, ")\n%s", src)

	return format.Source(w.Bytes())
}

// Returns a list of required imports
func Imports(src []byte) []string {
	i := []string{}
	srcString := string(src)
	if strings.Contains(srcString, "errors") {
		i = append(i, "errors")
	}
	if strings.Contains(srcString, "fmt") {
		i = append(i, "fmt")
	}
	sort.Strings(i)
	return i
}

// Generates the formatted go types for all schemas in the index
func generateGoTypes(idx *jsonschema.Index) ([]byte, error) {
	w := bytes.NewBufferString("\n")
	for _, k := range sortedMapKeysbyName(idx) {
		t, err := generateGoType((*idx)[k], idx)
		if err != nil {
			return nil, err
		}
		if string(t) != "" {
			fmt.Fprintf(w, "%s\n", t)
		}
	}

	return format.Source(w.Bytes())
}

// Generates the type definition for a schema
func generateGoType(s *jsonschema.Schema, idx *jsonschema.Index) ([]byte, error) {
	w := &bytes.Buffer{}
	switch s.Type {
	case "object":
		fmt.Fprintf(w, "type %v struct {\n", s.Name)
		for _, k := range sortedMapKeys(&s.Properties) {
			ref := generateGoRef(s.Properties[k], idx)
			if ref != "" {
				fmt.Fprintf(w, "\t%s\n", ref)
			}
		}
		fmt.Fprintf(w, "}\n")
	case "array":
		p, err := resolvRefToSchema(s.Items, idx)
		if err != nil {
			return nil, err
		}
		typ := s.Items.Type
		if p != s.Items {
			typ = p.Name
		}
		fmt.Fprintf(w, "type %v []%v\n", s.Name, typ)
	}
	return format.Source(w.Bytes())
}

// Generates the inline reference in a type for a schema
func generateGoRef(s *jsonschema.Schema, idx *jsonschema.Index) string {
	switch s.Type {
	case "string":
		return fmt.Sprintf("%v *string `json:\"%v,omitempty\"`", s.Name, s.JSONName)
	case "integer":
		return fmt.Sprintf("%v *int `json:\"%v,omitempty\"`", s.Name, s.JSONName)
	case "number":
		return fmt.Sprintf("%v *float64 `json:\"%v,omitempty\"`", s.Name, s.JSONName)
	case "boolean":
		return fmt.Sprintf("%v *bool `json:\"%v,omitempty\"`", s.Name, s.JSONName)
	case "object":
		return fmt.Sprintf("%v *%v `json:\"%v,omitempty\"`", s.Name, s.Name, s.JSONName)
	case "array":
		return fmt.Sprintf("%v *%v `json:\"%v,omitempty\"`", s.Name, s.Name, s.JSONName)
	case "ref":
		ref := (*idx)[s.Ref]
		return fmt.Sprintf("%v *%v `json:\"%v,omitempty\"`", s.Name, ref.Name, ref.JSONName)
	}
	return ""
}

// Generates the formatted go validate funcs for all types in the index
func generateGoTypesValidateFuncs(idx *jsonschema.Index) ([]byte, error) {
	w := bytes.NewBufferString("\n")
	for _, k := range sortedMapKeysbyName(idx) {
		t, err := generateGoTypeValidateFunc((*idx)[k], idx)
		if err != nil {
			return nil, err
		}
		if string(t) != "" {
			fmt.Fprintf(w, "%s\n", t)
		}
	}

	return format.Source(w.Bytes())
}

// Generates the validate func for a schema
func generateGoTypeValidateFunc(s *jsonschema.Schema, idx *jsonschema.Index) ([]byte, error) {
	w := &bytes.Buffer{}
	fmt.Fprintf(w, "func (t *%v) Validate() error {\n", s.Name)
	errorVarExists := ":"
	switch s.Type {
	case "object":
		checks, err := generateRequiredValidationCheck(idx, s)
		if err != nil {
			return nil, err
		}
		if string(checks) != "" {
			fmt.Fprintf(w, "\t%s\n", checks)
		}

		// Validate() calls of non-primitive type properties
		for _, k := range sortedMapKeys(&s.Properties) {
			p, err := resolvRefToSchema(s.Properties[k], idx)
			if err != nil {
				return nil, err
			}

			if p.Type == "object" || p.Type == "array" {
				fmt.Fprintf(w, "\terr %s= t.%v.Validate()\n", errorVarExists, p.Name)
				fmt.Fprintf(w, "\tif err != nil {\n")
				fmt.Fprintf(w, "\t\treturn err\n")
				fmt.Fprintf(w, "\t}\n")
				errorVarExists = ""
			}
		}
	case "array":
		as, err := resolvRefToSchema(s.Items, idx)
		if err != nil {
			return nil, err
		}

		if as.Type == "object" || as.Type == "array" {
			fmt.Fprintf(w, "\tfor _, a := range *t {\n")
			fmt.Fprintf(w, "\t\t\terr %s= a.Validate()\n", errorVarExists)
			fmt.Fprintf(w, "\t\tif err != nil {\n")
			fmt.Fprintf(w, "\t\t\treturn err\n")
			fmt.Fprintf(w, "\t\t}\n")
			fmt.Fprintf(w, "\t}\n")
			errorVarExists = ""
		}
	default:
		return nil, nil
	}
	fmt.Fprintf(w, "\treturn nil\n")
	fmt.Fprintf(w, "}\n")

	return format.Source(w.Bytes())
}

// generate "required" validation check
func generateRequiredValidationCheck(idx *jsonschema.Index, s *jsonschema.Schema) ([]byte, error) {
	if len(s.Required) == 0 {
		return nil, nil
	}
	w := &bytes.Buffer{}
	for _, p := range s.Required {
		ptr := s.Pointer + "/properties/" + p
		rs := (*idx)[ptr]
		if rs == nil {
			return nil, fmt.Errorf("jsonschema: %v does not exist in index", ptr)
		}
		fmt.Fprintf(w, "if t.%v == nil {\n", rs.Name)
		fmt.Fprintf(w, "\treturn errors.New(\"invalid %v: missing %v\")\n", s.JSONName, p)
		fmt.Fprintf(w, "}\n")
	}
	return w.Bytes(), nil
}

// Generates primitive type new funcs
func generateGoPrimitiveTypesNewFuncs() ([]byte, error) {
	b := bytes.NewBufferString(`
func newString(s string) *string {
	return &s
}

func newInt(i int) *int {
	return &i
}

func newFloat(f float64) *float64 {
	return &f
}

func newBool(b bool) *bool {
	return &b
}
	`)

	return format.Source(b.Bytes())
}

// returns a schema or referenced schema
func resolvRefToSchema(s *jsonschema.Schema, idx *jsonschema.Index) (*jsonschema.Schema, error) {
	if s.Type != "ref" {
		return s, nil
	}
	ref := (*idx)[s.Ref]
	if ref == nil {
		return nil, fmt.Errorf("jsonschema: %v does not exist in index", s.Ref)
	}
	return ref, nil
}

// creates a go friendly name from a JSON pointer
func goNameFromPointer(pointer string) string {
	p := strings.Split(pointer, "/")
	return goNameFromStrings(p[len(p)-1])
}

// creates a go friendly name from string parts
func goNameFromStrings(parts ...string) string {
	name := ""
	re := regexp.MustCompile("{|}|-|_")
	for _, p := range parts {
		c := re.ReplaceAllString(p, "")
		switch c {
		case "id":
			name += "ID"
		case "url":
			name += "URL"
		case "api":
			name += "API"
		default:
			name += strings.Title(c)
		}
	}
	return name
}

// returns a json friendly name from a pointer
func jsonNameFromPointer(pointer string) string {
	p := strings.Split(pointer, "/")
	return p[len(p)-1]
}

// returns map keys sorted by alphapet
func sortedMapKeys(m *jsonschema.Index) []string {
	var keys []string
	for k, _ := range *m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// returns map keys sorted by schema names
func sortedMapKeysbyName(m *jsonschema.Index) []string {
	var schemas []*jsonschema.Schema
	for _, v := range *m {
		schemas = append(schemas, v)
	}
	sort.Sort(byName(schemas))

	var keys []string
	for _, v := range schemas {
		keys = append(keys, v.Pointer)
	}
	return keys
}

// sorter for schema names
type byName []*jsonschema.Schema

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
