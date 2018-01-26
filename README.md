# jsonschema

A Go package that parses JSON Schema documents and generates go types including validations

[![CircleCI](https://circleci.com/gh/tfkhsr/jsonschema.svg?style=svg)](https://circleci.com/gh/tfkhsr/jsonschema)

## Features

* Parses schema documents based on https://tools.ietf.org/html/draft-handrews-json-schema-00
* Supports schema validation based on http://json-schema.org/latest/json-schema-validation.html
* Creates a schema lookup index based on JSON Pointers
* Generates source code for any supported language (currently only Go)
* No dependencies on external packages
* Test suite with shared schema fixtures
* Library and standalone compiler binary `jsonschemac`

## GoDoc

Godoc is available from https://godoc.org/github.com/tfkhsr/jsonschema.

## Install

To install as library run:

```
go get -u github.com/tfkhsr/jsonschema
```

To install the standalone compiler binary `jsonschemac` run:

```
go get -u github.com/tfkhsr/jsonschema/cmd/jsonschemac
```
