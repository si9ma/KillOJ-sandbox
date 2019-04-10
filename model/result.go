package model

const (
	CompileResType = "COMPILE"
	RunResType = "RUN"
)

const (
	SUCCESS = iota
	FAIL
)

type Result struct {
	ResultType string	`json:"resType"`
	Status int64	`json:"status"`
	Errno int64 	`json:"errno,omitempty"`
}

type CompileResult struct {
	Result
	Message string		`json:"msg,omitempty"`
}

type RunResult struct {
	Result
	Runtime  int64  `json:"runtime,omitempty"`
	Memory   int64  `json:"memory,omitempty"`
	Status   int32  `json:"status"`
	Error    string `json:"error,omitempty"`
	Input    string `json:"input,omitempty"`
	Output   string `json:"output,omitempty"`
	Expected string `json:"expected,omitempty"`
}
