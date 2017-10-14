package main

import (
	//"encoding/json"
	"fmt"
	//"log"
	//"os"
	"reflect"
)

// Using "U" to indicate "unboxed" types.

type KindU int

const (
	AString KindU = iota
	ANumber
	AStruct
	AList
)

type Type struct {
	kind KindU
	/* Field is present IFF kind is AStruct */
	theStruct *Struct
	/* Field is present IFF kind is AList */
	theList *Type
}

type Field struct {
	name  string
	value *Type
}

type Struct []*Field

type Pattern struct {
}

func main() {
	var x float64 = 3.4
	fmt.Println("type:", reflect.TypeOf(x))
}
