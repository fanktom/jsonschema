package jsonschema

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Generate a complete go package with all types and validations
func ExampleGenerateGoSrcPackage() {
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

	src, err := GenerateGoSrcPackage([]byte(schema), "main")
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
func ExampleGenerateGoSrc() {
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

	src, err := GenerateGoSrc([]byte(schema))
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

// Parse a schema for inspection
func ExampleParse() {
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
	idx, err := Parse([]byte(schema))
	if err != nil {
		panic(err)
	}

	// sort pointers by name
	pointers := make([]string, 0)
	for pointer, _ := range *idx {
		pointers = append(pointers, pointer)
	}
	sort.Strings(pointers)

	// print pointer and Go friendly name
	for _, pointer := range pointers {
		fmt.Printf("%s : %s\n", pointer, (*idx)[pointer].Name)
	}
	// Output:
	// #/definitions/role : Role
	// #/definitions/role/properties/name : Name
	// #/definitions/roles : Roles
	// #/definitions/roles/items : Items
	// #/definitions/user : User
	// #/definitions/user/properties/id : ID
	// #/definitions/user/properties/name : Name
	// #/definitions/user/properties/roles : Roles
}

// Generate a sample JSON document instance conforming to a schema
func ExampleSchema_NewInstance() {
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
	idx, err := Parse([]byte(schema))
	if err != nil {
		panic(err)
	}

	// create go instance
	inst, err := (*idx)["#/definitions/user"].NewInstance(idx)
	if err != nil {
		panic(err)
	}

	// marshal e.g. to json
	raw, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", raw)
	// Output:
	// {
	//   "id": "string",
	//   "name": "string",
	//   "roles": [
	//     {
	//       "name": "string"
	//     }
	//   ]
	// }
}
