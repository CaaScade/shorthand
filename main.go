package main

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	iso := SequenceIsos(ServicePortsIso(), IdentityIso())
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
	return ZoomIso(MkP(P{"kind": "Service", "spec": P{"ports": AnyW}}),
		MultiplyIso(Port()))
}

// Port doot.
func Port() *Iso {
	from := MkP(P{
		"name":     StringW,
		"port":     FloatW,
		"protocol": StringW})
	to := ConstPattern(StringW)

	split := func(from *Pattern) (*Pattern, error) {
		x := from.Extract()

		name := StringAt(x, "name")
		port := FloatAt(x, "port")
		protocol := StringAt(x, "protocol")

		if protocol == "TCP" {
			if name == "http" && port == 80 {
				return ConstPattern("http"), nil
			}

			if name == "https" && port == 443 {
				return ConstPattern("https"), nil
			}

			return ConstPattern(fmt.Sprintf("%s:%v", name, port)), nil
		}

		return ConstPattern(fmt.Sprintf("%s:%v:%s", name, port, protocol)), nil
	}

	unsplit := func(to *Pattern) (*Pattern, error) {
		x := to.ExtractString()
		segments := strings.Split(x, ":")
		l := len(segments)

		var err error
		var name string
		var port float64
		var protocol string

		if l > 0 {
			name = segments[0]
		} else {
			return nil, fmt.Errorf("expected non-empty string")
		}

		if l > 1 {
			port, err = strconv.ParseFloat(segments[1], 64)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse port (%v)", err)
			}
		} else {
			switch name {
			case "http":
				port = 80
			case "https":
				port = 443
			default:
				return nil, fmt.Errorf("dunno port for %s", name)
			}
		}

		if l > 2 {
			protocol = segments[2]
		} else {
			protocol = "TCP"
		}

		return MkP(P{"name": name, "port": port, "protocol": protocol}), nil
	}

	return MkIso(from, to, split, unsplit)
}
