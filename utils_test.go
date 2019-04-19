package main

import (
	"fmt"
	"testing"
)

func TestToRealStr(t *testing.T) {
	src := "hello\\n"
	dst := escapeString(src)
	srcBytes := []byte(src)
	dstBytes := []byte(dst)
	fmt.Print(srcBytes, dstBytes)
}
