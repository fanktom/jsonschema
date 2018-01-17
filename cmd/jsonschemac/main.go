// Command jsonschemac compiles a jsonschema document into go types
package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"gitlab.mi.hdm-stuttgart.de/smu/jsonschema"
)

func main() {
	file := flag.String("file", "schema.json", "json schema file to load")
	pack := flag.String("package", "main", "name for generated package")
	flag.Parse()

	buf, err := ioutil.ReadFile(*file)
	if err != nil {
		panic(err)
	}

	src, err := jsonschema.GenerateGoSrcPackage(buf, *pack)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(src))
}
