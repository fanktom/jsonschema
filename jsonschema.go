/*
Package jsonschema parses JSON Schema documents and generates go types including validations.

The JSON Schema implementation is based on https://tools.ietf.org/html/draft-handrews-json-schema-00.
The validation implementation is based on http://json-schema.org/latest/json-schema-validation.html (supported validations: "required").

Parse a schema into a "main" package including imports, e.g. for saving to a file:

	schema := []byte(`...`)
	src, err := GenerateGoSrcPackage(schema, "main")
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
	src, err := GenerateGoSrc(schema)
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

Further, a schema can be parsed into a map of JSON pointers to Schemas (Index):

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
	idx, err := Parse(schema)
	if err != nil {
		panic(err)
	}

	// idx now contains:
	// "#/definitions/user"      : *Schema{...}
	// "#/definitions/user/id"   : *Schema{...}
	// "#/definitions/user/name" : *Schema{...}

*/
package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"regexp"
	"sort"
	"strings"
)

// Maps JSON pointers to Schemas
type Index map[string]*Schema

// A schema is a part of a schema document tree
type Schema struct {
	// Optional Title
	Title string `json:"title"`

	// Optional Description
	Description string `json:"description"`

	// JSON pointer as defined in https://tools.ietf.org/html/rfc6901
	Pointer string `json:"pointer"`

	// JSON pointer without #/definitions/ part
	PointerName string

	// Go friendly name
	Name string `json:"name"`

	// JSON friendly name
	JSONName string `json:"jsonName"`

	// Type as defined in http://json-schema.org/latest/json-schema-core.html#rfc.section.4.2
	Type string `json:"type"`

	// Definitions as defined in http://json-schema.org/latest/json-schema-validation.html#rfc.section.7.1
	Definitions Index `json:"definitions"`

	// Properties as defined in http://json-schema.org/latest/json-schema-validation.html#rfc.section.6.18
	Properties Index `json:"properties"`

	// Items as defined in http://json-schema.org/latest/json-schema-validation.html#rfc.section.6.9
	Items *Schema `json:"items"`

	// Reference as defined in http://json-schema.org/latest/json-schema-core.html#rfc.section.8
	Ref string `json:"$ref"`

	// Validation properties
	Required []string `json:"required"`
}

// parse traverses the schema document tree to collect information and structure
func (s *Schema) parse(idx *Index, pointer string) {
	if len(s.Definitions) > 0 {
		for name, sch := range s.Definitions {
			sch.parse(idx, pointer+"/definitions/"+name)
		}
	}
	if len(s.Properties) > 0 {
		for name, sch := range s.Properties {
			sch.parse(idx, pointer+"/properties/"+name)
		}
	}
	if s.Items != nil {
		s.Items.parse(idx, pointer+"/items")
	}
	if pointer == "#" {
		return
	}
	if s.Ref != "" {
		s.Type = "ref"
	}

	s.Pointer = pointer
	s.PointerName = strings.Replace(pointer, "#/definitions/", "", 1)
	s.Name = goNameFromPointer(pointer)
	s.JSONName = jsonNameFromPointer(pointer)

	(*idx)[pointer] = s
}

// Creates a new instance conforming to the schema
func (s *Schema) NewInstance(idx *Index) (interface{}, error) {
	switch s.Type {
	case "ref":
		sch, err := resolvRefToSchema(s, idx)
		if err != nil {
			return nil, err
		}
		return sch.NewInstance(idx)
	case "object":
		m := make(map[string]interface{})
		for name, sch := range s.Properties {
			d, err := sch.NewInstance(idx)
			if err != nil {
				return nil, err
			}
			m[name] = d
		}
		return m, nil
	case "array":
		a := make([]interface{}, 0)
		d, err := s.Items.NewInstance(idx)
		if err != nil {
			return nil, err
		}
		a = append(a, d)
		return a, nil
	case "string":
		return "string", nil
	case "integer":
		return 42, nil
	case "number":
		return 3.14, nil
	case "boolean":
		return true, nil
	case "null":
		return nil, nil
	}
	return nil, nil
}

// Parse converts a raw JSON schema document to an index of schemas
func Parse(b []byte) (*Index, error) {
	var s Schema
	err := json.Unmarshal(b, &s)
	if err != nil {
		return nil, fmt.Errorf("jsonschema: %v", err)
	}

	idx := &Index{}
	s.parse(idx, "#")

	return idx, nil
}

// Generates the go src from a schema without imports and a package
// This allows embedding the code in other code
func GenerateGoSrc(s []byte) ([]byte, error) {
	idx, err := Parse(s)
	if err != nil {
		return nil, err
	}

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

// Generates the complete go src from a schema for a package
func GenerateGoSrcPackage(s []byte, p string) ([]byte, error) {
	src, err := GenerateGoSrc(s)
	if err != nil {
		return nil, err
	}

	w := &bytes.Buffer{}
	fmt.Fprintf(w, `package %v

import (
`, p)
	for _, i := range ImportsForGoSrc(src) {
		fmt.Fprintf(w, "\t\"%s\"\n", i)
	}
	fmt.Fprintf(w, ")\n%s", src)

	return format.Source(w.Bytes())
}

// Returns a list of required Go imports
func ImportsForGoSrc(src []byte) []string {
	i := []string{}
	if strings.Contains(string(src), "errors") {
		i = append(i, "errors")
	}
	if strings.Contains(string(src), "fmt") {
		i = append(i, "fmt")
	}
	sort.Strings(i)
	return i
}

// Generates the formatted go types for all schemas in the index
func generateGoTypes(idx *Index) ([]byte, error) {
	w := bytes.NewBufferString("\n")
	for _, k := range sortedMapKeysbyName(idx) {
		t, err := (*idx)[k].generateGoType(idx)
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
func (s *Schema) generateGoType(idx *Index) ([]byte, error) {
	w := &bytes.Buffer{}
	switch s.Type {
	case "object":
		fmt.Fprintf(w, "type %v struct {\n", s.Name)
		for _, k := range sortedMapKeys(&s.Properties) {
			ref := s.Properties[k].generateGoRef(idx)
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
func (s *Schema) generateGoRef(idx *Index) string {
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
func generateGoTypesValidateFuncs(idx *Index) ([]byte, error) {
	w := bytes.NewBufferString("\n")
	for _, k := range sortedMapKeysbyName(idx) {
		t, err := (*idx)[k].generateGoTypeValidateFunc(idx)
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
func (s *Schema) generateGoTypeValidateFunc(idx *Index) ([]byte, error) {
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
func generateRequiredValidationCheck(idx *Index, s *Schema) ([]byte, error) {
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
func resolvRefToSchema(s *Schema, idx *Index) (*Schema, error) {
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
func sortedMapKeys(m *Index) []string {
	var keys []string
	for k, _ := range *m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// returns map keys sorted by schema names
func sortedMapKeysbyName(m *Index) []string {
	var schemas []*Schema
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
type byName []*Schema

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
