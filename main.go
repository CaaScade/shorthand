package main

import (
	"fmt"
	"github.com/kr/pretty"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func steal() {
	inputDir := os.Args[1]
	outputDir := os.Args[2]
	err := StealYamlFiles(inputDir, outputDir)
	if err != nil {
		log.Fatal(err)
	}
}

func check() {
	inputDir := os.Args[1]
	outputDir := os.Args[2]
	failDir := os.Args[3]
	paths, err := YamlPathsInDir(inputDir)
	if err != nil {
		log.Fatal(err)
	}

	iso := MultiplyIso(SequenceIsos(ServicePortsIso(), IdentityIso()))

	for _, path := range paths {
		_, _ = pretty.Println(path)
		var relPath, pristine, transformed, reverted string
		pristine, transformed, reverted, err = RoundTrip(path, iso)
		relPath, err = filepath.Rel(inputDir, path)
		if err != nil {
			log.Fatal(err)
		}

		err = WriteResults(
			relPath,
			outputDir,
			failDir,
			pristine,
			transformed,
			reverted,
			err)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	//steal()
	check()
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
