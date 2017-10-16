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

	_, err = pretty.Println(v)
	if err != nil {
		log.Fatal(err)
	}

	p := MkP(P{"spec": P{"selector": P{"app": nil}}})
	p.Match(v)
	p.Erase(v)

	q := MkP(P{"spec": P{"app": "doot"}})
	q.Write(v)

	o, err1 := yaml.Marshal(v)
	if err1 != nil {
		log.Fatal(err1)
	}

	fmt.Println(string(o))
}
