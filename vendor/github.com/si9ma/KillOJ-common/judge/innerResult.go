package judge

const (
	CompileResType = "COMPILE"
	RunResType     = "RUN"
)

const (
	SUCCESS = int64(iota)
	FAIL
)

// result struct in inner system
type InnerResult struct {
	// for compile result and run result
	ID         string `json:"id"`
	ResultType string `json:"resType"`
	Status     int64  `json:"status"`
	Errno      int64  `json:"errno,omitempty"`
	Message    string `json:"msg,omitempty"`
	StdErr     string `json:"stderr,omitempty"`    // stderr from inner process
	TimeLimit  int64  `json:"timelimit,omitempty"` // time limit in ms

	// only for run result
	Runtime  int64  `json:"runtime,omitempty"`  // time usage in ms
	Memory   int64  `json:"memory,omitempty"`   // memory usage in KB
	MemLimit int64  `json:"memlimit,omitempty"` // memory limit in KB
	Input    string `json:"input,omitempty"`
	Output   string `json:"output,omitempty"`
	Expected string `json:"expected,omitempty"`
}

func (i InnerResult) ToOuterResult() OuterResult {
	outerResult := OuterResult{
		ID:        i.ID,
		StdErr:    i.StdErr,
		Runtime:   i.Runtime,
		Memory:    i.Memory,
		Input:     i.Input,
		Output:    i.Output,
		Expected:  i.Expected,
		TimeLimit: i.TimeLimit,
		MemLimit:  i.MemLimit,
	}

	// calculate Message and Status
	if i.Status == FAIL {
		outerResult.Status = GetStatusByErrNo(i.Errno)
		outerResult.Message = GetOuterErrorMsgByErrNo(i.Errno)
	} else {
		outerResult.Status = AcceptedStatus
	}

	return outerResult
}
