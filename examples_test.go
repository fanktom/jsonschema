package jsonschema

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Parse a schema into an Index of Schemas
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

	// marshal to json
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
