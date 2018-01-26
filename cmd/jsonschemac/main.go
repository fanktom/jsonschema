// Command jsonschemac compiles a jsonschema document into go types
package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/tfkhsr/jsonschema"
	"github.com/tfkhsr/jsonschema/golang"
)

func main() {
	file := flag.String("file", "schema.json", "json schema file to load")
	pack := flag.String("package", "main", "name for generated package")
	gen := flag.String("generator", "go", "generator to use")
	flag.Parse()

	// read schema
	buf, err := ioutil.ReadFile(*file)
	if err != nil {
		panic(err)
	}

	// parse schema
	idx, err := jsonschema.Parse(buf)
	if err != nil {
		panic(err)
	}

	// generate src
	var src []byte
	switch *gen {
	case "go":
		src, err = golang.PackageSrc(idx, *pack)
	default:
		err = fmt.Errorf("unknown generator: %s", *gen)
	}

	if err != nil {
		panic(err)
	}

	// print src
	fmt.Println(string(src))
}
