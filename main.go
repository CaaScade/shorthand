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

	p := MkP(P{"spec": P{"selector": P{"app": Wild, "doot": Absent}}})
	p.Match(v)
	if p.HasErrors() {
		log.Fatal(pretty.Sprint(p))
	}

	p.Erase(v)

	x, err1 := At(p.Extract(), "spec", "selector", "app")
	_, _ = pretty.Println(p.Extract())
	if err1 != nil {
		log.Fatal(err1)
	}

	q := MkP(P{"spec": P{"app": x}})
	_, err = q.Write(v)
	if err != nil {
		log.Fatal(err)
	}

	o, err1 := yaml.Marshal(v)
	if err1 != nil {
		log.Fatal(err1)
	}

	fmt.Println(string(o))
}
