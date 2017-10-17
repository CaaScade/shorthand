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
	// WHEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE
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
	// WHEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE

	var w interface{}
	iso := ServicePortsIso()
	w, err = iso.forward.View(v)
	if err != nil {
		log.Fatal(err)
	}

	switch w.(type) {
	case map[string]interface{}:
		// Continue
	default:
		log.Fatal(pretty.Sprint(w))
	}

	// WHEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE
	var o []byte

	o, err = yaml.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	// WHEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE

	fmt.Println(string(o))

	w, err = iso.backward.View(w)
	if err != nil {
		log.Fatal(err)
	}

	switch w.(type) {
	case map[string]interface{}:
		// Continue
	default:
		log.Fatal(pretty.Sprint(w))
	}

	// WHEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE
	o, err = yaml.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(o))
	// WHEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE
}

// ServicePortsIso doot.
func ServicePortsIso() *Iso {
	return ZoomIso(MkP(P{"kind": "Service", "spec": P{"ports": Wild}}),
		MultiplyIso(Port()))
}

// Port doot.
func Port() *Iso {
	return SequenceIsos(HTTP(), HTTPS())
}

// HTTP doot.
func HTTP() *Iso {
	from := MkP(P{
		"name":     "http",
		"port":     80,
		"protocol": "TCP"})
	split := func(from *Pattern) (*Pattern, error) {
		return ConstPattern("http"), nil
	}

	to := ConstPattern("http")
	unsplit := func(to *Pattern) (*Pattern, error) {
		return MkP(P{
			"name":     "http",
			"port":     80,
			"protocol": "TCP"}), nil
	}

	return &Iso{&Prism{from, split}, &Prism{to, unsplit}}
}

// HTTPS doot.
func HTTPS() *Iso {
	from := MkP(P{
		"name":     "https",
		"port":     443,
		"protocol": "TCP"})
	split := func(from *Pattern) (*Pattern, error) {
		return ConstPattern("https"), nil
	}

	to := ConstPattern("https")
	unsplit := func(to *Pattern) (*Pattern, error) {
		return MkP(P{
			"name":     "https",
			"port":     443,
			"protocol": "TCP"}), nil
	}

	return &Iso{&Prism{from, split}, &Prism{to, unsplit}}
}
