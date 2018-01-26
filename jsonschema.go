/*
Package jsonschema parses JSON Schema documents.
The resulting schema can be used to generate source code in any supported language.

The JSON Schema implementation is based on https://tools.ietf.org/html/draft-handrews-json-schema-00.
The validation implementation is based on http://json-schema.org/latest/json-schema-validation.html.

Validations

required: http://json-schema.org/latest/json-schema-validation.html#rfc.section.6.5.3

Generators

go: https://godoc.org/github.com/tfkhsr/jsonschema/golang


Parse a schema into a map of JSON pointers to Schemas (Index):

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
	"encoding/json"
	"fmt"
	"regexp"
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

	// Camel-cased name
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
	s.Name = nameFromPointer(pointer)
	s.JSONName = jsonNameFromPointer(pointer)

	(*idx)[pointer] = s
}

// Creates a new instance conforming to the schema
func (s *Schema) NewInstance(idx *Index) (interface{}, error) {
	switch s.Type {
	case "ref":
		sch, err := resolveRefToSchema(s, idx)
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

// Parse converts a raw JSON schema document to an Index of Schemas
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

// returns a schema or referenced schema
func resolveRefToSchema(s *Schema, idx *Index) (*Schema, error) {
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
func nameFromPointer(pointer string) string {
	p := strings.Split(pointer, "/")
	return nameFromStrings(p[len(p)-1])
}

// creates a go friendly name from string parts
func nameFromStrings(parts ...string) string {
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
