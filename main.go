package main

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
	"github.com/kr/pretty"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	inputFile := os.Args[1]
	content, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	v := map[string]interface{}{}
	err = yaml.Unmarshal(content, &v)
	if err != nil {
		log.Fatal(err)
	}

	_, _ = pretty.Println(v)

	var w interface{}
	w, err = ServicePorts(Identity()).View(v)

	switch w.(type) {
	case map[string]interface{}:
		// Continue
	default:
		log.Fatal(pretty.Sprint(w))
	}

	o, err1 := yaml.Marshal(v)
	if err1 != nil {
		log.Fatal(err1)
	}

	fmt.Println(string(o))
}
