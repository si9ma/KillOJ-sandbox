package judge

type Status struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var (
	AcceptedStatus = Status{
		Code: 0,
		Msg:  "Accepted",
	}
	JudingStatus = Status{
		Code: 1,
		Msg:  "Juding",
	}
	RuntimeErrorStatus = Status{
		Code: 2,
		Msg:  "RuntimeError",
	}
	CompileErrorStatus = Status{
		Code: 3,
		Msg:  "CompileError",
	}
	RunTimeOutStatus = Status{
		Code: 4,
		Msg:  "RunTimeOut",
	}
	OOMStatus = Status{
		Code: 5,
		Msg:  "OOM",
	}
	WrongAnswerStatus = Status{
		Code: 6,
		Msg:  "WrongAnswer",
	}
	SystemErrorStatus = Status{
		Code: 7,
		Msg:  "SystemError",
	}
	NullStatus = Status{
		Code: -1,
		Msg:  "",
	}
)
