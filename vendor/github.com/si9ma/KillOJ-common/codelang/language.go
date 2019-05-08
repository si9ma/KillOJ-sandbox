// programing language
package codelang

import "errors"

var UnknownLangErr = errors.New("unknown language")

const (
	C = iota
	CPP
	JAVA
	GOLANG
)

type Language struct {
	Code     int
	Name     string
	FileName string
}

var (
	LangC    = Language{C, "c", "main.c"}
	LangCPP  = Language{CPP, "cpp", "main.cpp"}
	LangJava = Language{JAVA, "java", "Main.java"}
	LangGo   = Language{GOLANG, "go", "main.go"}
)

var codeMap = map[int]Language{
	C:      LangC,
	CPP:    LangCPP,
	JAVA:   LangJava,
	GOLANG: LangGo,
}

var nameMap = map[string]Language{
	"c":    LangC,
	"cpp":  LangCPP,
	"java": LangJava,
	"go":   LangGo,
}

func GetLangByCode(code int) (*Language, error) {
	if val, ok := codeMap[code]; ok {
		return &val, nil
	}

	return nil, UnknownLangErr
}

func GetLangByName(name string) (*Language, error) {
	if val, ok := nameMap[name]; ok {
		return &val, nil
	}

	return nil, UnknownLangErr
}
