package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"log"
)

func main() {
	v := map[string]interface{}{}
	blob := `{"doot":{},"boop":{"wat":"yes"}}`
	if err := json.Unmarshal([]byte(blob), &v); err != nil {
		log.Fatal(err)
	}

	var out bytes.Buffer
	if err := json.Indent(&out, []byte(blob), "", "  "); err != nil {
		log.Fatal(err)
	}

	blob1 := out.String()
	fmt.Println(blob1)
	//p := MkPattern(MkField("boop", MkField("wat")))
	p := MkP(P{"boop": P{"wat": nil}})
	p.Match(v)
	//spew.Dump(p)
	fmt.Println(p.Extract())

	v1 := p.Erase(v)
	blob2, err2 := json.MarshalIndent(v1, "", "  ")
	if err2 != nil {
		log.Fatal(err2)
	}
	fmt.Println(string(blob2))
	//spew.Dump(v)

	//q := MkPattern(MkField("b", ValField("p", 1000)))
	q := MkP(P{"b": P{"p": 1000}})
	v2 := q.Write(v1)
	blob3, err3 := json.MarshalIndent(v2, "", "  ")
	if err3 != nil {
		log.Fatal(err3)
	}
	fmt.Println(string(blob3))
	//spew.Dump(v2)
}
