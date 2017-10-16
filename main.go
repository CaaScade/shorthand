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
	telescope := MkP(P{"kind": "Service", "spec": P{"ports": Wild}})
	w, err = Zoom(telescope, Multiply(Sequence(HTTP(), HTTPS()))).View(v)
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

	w, err = Zoom(telescope,
		Multiply(Sequence(FromHTTP(), FromHTTPS()))).View(w)
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

// ServicePorts telescope.
func ServicePorts() *Pattern {
	return MkP(P{"kind": "Service", "spec": P{"ports": Wild}})
}

// HTTPS doot.
func HTTPS() *Prism {
	from := MkP(P{
		"name":     "https",
		"port":     443,
		"protocol": "TCP"})
	split := func(from *Pattern) (*Pattern, error) {
		return ConstPattern("https"), nil
	}

	return &Prism{from, split}
}

// FromHTTPS doot.
func FromHTTPS() *Prism {
	from := ConstPattern("https")
	split := func(from *Pattern) (*Pattern, error) {
		return MkP(P{
			"name":     "https",
			"port":     443,
			"protocol": "TCP"}), nil
	}

	return &Prism{from, split}
}

// HTTP doot.
func HTTP() *Prism {
	from := MkP(P{
		"name":     "http",
		"port":     80,
		"protocol": "TCP"})
	split := func(from *Pattern) (*Pattern, error) {
		return ConstPattern("http"), nil
	}

	return &Prism{from, split}
}

// FromHTTP doot.
func FromHTTP() *Prism {
	from := ConstPattern("http")
	split := func(from *Pattern) (*Pattern, error) {
		return MkP(P{
			"name":     "http",
			"port":     80,
			"protocol": "TCP"}), nil
	}

	return &Prism{from, split}
}
