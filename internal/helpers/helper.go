package helpers

import (
	"encoding/json"
	"fmt"
)

// PrettyPrintStruct prints prettily a struct
func PrettyPrintStruct(d any) {
	s, _ := json.MarshalIndent(d, "", "\t")
	fmt.Print(string(s))
}

// PrettyPrintStructResponse prints prettily a struct and returns it as a string
func PrettyPrintStructResponse(d any) string {
	s, _ := json.Marshal(d)
	return string(s)
}

// PointerValue returns a pointer value
func PointerValue[T any](p T) *T {
	return &p
}

// GetValue returns safely a pointer value
func GetValue[T any](v *T) T {
	if v != nil {
		return *v
	}
	return *new(T)
}

func StructArrayToAnyArray[T any](s []T) []any {
	var a []any
	for _, v := range s {
		a = append(a, v)
	}
	return a
}
