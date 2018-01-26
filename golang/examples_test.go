package golang

import (
	"fmt"

	"github.com/tfkhsr/jsonschema"
)

// Generate a complete go package with all types and validations
func ExamplePackageSrc() {
	schema := `
{
	"definitions": {
		"user": {
			"type": "object",
			"properties": {
				"id": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"roles": {
					"$ref": "#/definitions/roles"
				}
			}
		},
		"roles": {
			"type": "array",
			"items": {
				"$ref": "#/definitions/role"
			}
		},
		"role": {
			"type": "object",
			"properties": {
				"name": {
					"type": "string"
				}
			},
			"required": ["name"]
		}
	}
}
`
	// parse into index
	idx, err := jsonschema.Parse([]byte(schema))
	if err != nil {
		panic(err)
	}

	// generate package source
	src, err := PackageSrc(idx, "main")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", src)
	// Output:
	//package main
	//
	//import (
	//	"errors"
	//)
	//
	//type Role struct {
	//	Name *string `json:"name,omitempty"`
	//}
	//
	//type Roles []Role
	//
	//type User struct {
	//	ID    *string `json:"id,omitempty"`
	//	Name  *string `json:"name,omitempty"`
	//	Roles *Roles  `json:"roles,omitempty"`
	//}
	//
	//func (t *Role) Validate() error {
	//	if t.Name == nil {
	//		return errors.New("invalid role: missing name")
	//	}
	//
	//	return nil
	//}
	//
	//func (t *Roles) Validate() error {
	//	for _, a := range *t {
	//		err := a.Validate()
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	return nil
	//}
	//
	//func (t *User) Validate() error {
	//	err := t.Roles.Validate()
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}
	//
	//func newString(s string) *string {
	//	return &s
	//}
	//
	//func newInt(i int) *int {
	//	return &i
	//}
	//
	//func newFloat(f float64) *float64 {
	//	return &f
	//}
	//
	//func newBool(b bool) *bool {
	//	return &b
	//}

}

// Generate source with all types and validations but no imports and package
func ExampleSrc() {
	schema := `
{
	"definitions": {
		"user": {
			"type": "object",
			"properties": {
				"id": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"roles": {
					"$ref": "#/definitions/roles"
				}
			}
		},
		"roles": {
			"type": "array",
			"items": {
				"$ref": "#/definitions/role"
			}
		},
		"role": {
			"type": "object",
			"properties": {
				"name": {
					"type": "string"
				}
			},
			"required": ["name"]
		}
	}
}
`
	// parse into index
	idx, err := jsonschema.Parse([]byte(schema))
	if err != nil {
		panic(err)
	}

	// generate source
	src, err := Src(idx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", src)
	// Output:
	//type Role struct {
	//	Name *string `json:"name,omitempty"`
	//}
	//
	//type Roles []Role
	//
	//type User struct {
	//	ID    *string `json:"id,omitempty"`
	//	Name  *string `json:"name,omitempty"`
	//	Roles *Roles  `json:"roles,omitempty"`
	//}
	//
	//func (t *Role) Validate() error {
	//	if t.Name == nil {
	//		return errors.New("invalid role: missing name")
	//	}
	//
	//	return nil
	//}
	//
	//func (t *Roles) Validate() error {
	//	for _, a := range *t {
	//		err := a.Validate()
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	return nil
	//}
	//
	//func (t *User) Validate() error {
	//	err := t.Roles.Validate()
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}
	//
	//func newString(s string) *string {
	//	return &s
	//}
	//
	//func newInt(i int) *int {
	//	return &i
	//}
	//
	//func newFloat(f float64) *float64 {
	//	return &f
	//}
	//
	//func newBool(b bool) *bool {
	//	return &b
	//}

}
