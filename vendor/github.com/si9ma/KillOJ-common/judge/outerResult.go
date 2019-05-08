package judge

// the result expose to user
type OuterResult struct {
	// for compile result and run result
	ID        string `json:"id"`
	Status    Status `json:"status"`
	Message   string `json:"msg,omitempty"`
	StdErr    string `json:"stderr,omitempty"`    // stderr from inner process
	TimeLimit int64  `json:"timelimit,omitempty"` // time limit in ms

	// only for run result
	Runtime         int64  `json:"runtime,omitempty"`  // time usage in ms
	Memory          int64  `json:"memory,omitempty"`   // memory usage in KB
	MemLimit        int64  `json:"memlimit,omitempty"` // memory limit in KB
	Input           string `json:"input,omitempty"`
	Output          string `json:"output,omitempty"`
	Expected        string `json:"expected,omitempty"`
	TestCaseNum     int    `json:"test_case_num,omitempty"`
	SuccessTestCase int    `json:"success_test_case,omitempty"`

	//
	IsComplete bool `json:"iscomplete"`
}
