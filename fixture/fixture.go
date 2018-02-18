// Package fixture provides common schemas for testing and evaluation
package fixture

const (
	// Simple schema with definitions and refs
	TestSchemaWithDefinitions = `
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

	// Schema without definitions
	TestSchemaDirect = `
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

	// Schema with nested definitions
	TestSchemaWithNestedDefinitions = `
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
	// All possible primitive types of a schema
	TestSchemaPrimitiveTypes = `
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

	// Schema with validation: required
	TestSchemaRequiredValidation = `
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
	// Schema with array of objects
	TestSchemaWithArrayOfObjects = `
{
	"definitions": {
    "movies": {
      "type": "array",
			"items": {
				"type": "object",
      	"properties": {
        	"id": {
          	"type": "string"
        	},
        	"name": {
          	"type": "string"
        	},
        	"year": {
          	"type": "integer"
        	}
				}
			}
		}
	}
}
`
)
