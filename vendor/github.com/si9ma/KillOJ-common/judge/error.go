package judge

import (
	"github.com/si9ma/KillOJ-common/lang"
	"github.com/si9ma/KillOJ-common/tip"
)

// 11xx: compile error
const (
	INNER_COMPILER_ERR = 1101 // error from real compiler
	OUTER_COMPILER_ERR = 1102 // error from our outer compiler
	COMPILE_TIMEOUT    = 1103
)

// 1[2-4]xx: container error
const (
	// 12xx: error from outermost process (the process to start container)
	RUNNER_ERR = 1201

	// 13xx: container error (error from container)
	CONTAINER_ERR = 1301

	// 14xx: run error (error from program run in container)
	APP_ERR                   = 1401
	WRONG_ANSWER_ERR          = 1402
	OUT_OF_MEMORY_ERR         = 1403
	RUN_TIMEOUT_ERR           = 1404
	BAD_SYSTEMCALL_ERR        = 1405
	NO_ENOUGH_PID_ERR         = 1406
	JAVA_SECURITY_MANAGER_ERR = 1407
)

type errAdapter struct {
	// inner error message
	// lang code -> message
	InnerMsg    tip.Tip
	OuterStatus Status // the status expose to user

	// the message expose to user,
	// lang code -> message
	OuterMsg tip.Tip
}

// System error
var systemErrAdapter = errAdapter{
	InnerMsg:    tip.SystemErrorTip,
	OuterStatus: SystemErrorStatus,
	OuterMsg:    tip.SystemErrorTip,
}

// map errno to errAdapter
var errAdapterMapping = map[int64]errAdapter{
	RUNNER_ERR:         systemErrAdapter,
	CONTAINER_ERR:      systemErrAdapter,
	OUTER_COMPILER_ERR: systemErrAdapter,

	INNER_COMPILER_ERR: {
		InnerMsg:    tip.CompileErrorTip,
		OuterStatus: CompileErrorStatus,
		OuterMsg:    tip.CompileErrorTip,
	},
	COMPILE_TIMEOUT: {
		InnerMsg:    tip.CompileTimeOutTip,
		OuterStatus: CompileErrorStatus,
		OuterMsg:    tip.CompileTimeOutTip,
	},
	APP_ERR: {
		InnerMsg:    tip.RuntimeErrorTip,
		OuterStatus: RuntimeErrorStatus,
		OuterMsg:    tip.RuntimeErrorTip,
	},
	RUN_TIMEOUT_ERR: {
		InnerMsg:    tip.RunTimeOutTip,
		OuterStatus: RunTimeOutStatus,
		OuterMsg:    tip.RunTimeOutTip,
	},
	WRONG_ANSWER_ERR: {
		InnerMsg:    tip.WrongAnswerErrorTip,
		OuterStatus: WrongAnswerStatus,
		OuterMsg:    tip.WrongAnswerErrorTip,
	},
	OUT_OF_MEMORY_ERR: {
		InnerMsg:    tip.OOMErrorTip,
		OuterStatus: OOMStatus,
		OuterMsg:    tip.OOMErrorTip,
	},

	// InnerMsg and OuterMsg is different
	BAD_SYSTEMCALL_ERR: {
		InnerMsg:    tip.BadSysErrorTip,
		OuterStatus: RuntimeErrorStatus,
		OuterMsg:    tip.RuntimeErrorTip,
	},
	NO_ENOUGH_PID_ERR: {
		InnerMsg:    tip.NoEnoughPidErrorTip,
		OuterStatus: RuntimeErrorStatus,
		OuterMsg:    tip.RuntimeErrorTip,
	},
	JAVA_SECURITY_MANAGER_ERR: {
		InnerMsg:    tip.JavaSecurityManagerErrorTip,
		OuterStatus: RuntimeErrorStatus,
		OuterMsg:    tip.RuntimeErrorTip,
	},
}

// return empty string when don't exist
func GetInnerErrorMsgByErrNo(errno int64) string {
	lan := lang.GetLangFromEnv()
	if adapter, ok := errAdapterMapping[errno]; ok {
		if val, ok := adapter.InnerMsg[lan]; ok {
			return val
		}
	}

	return ""
}

// return empty string when don't exist
func GetOuterErrorMsgByErrNo(errno int64) string {
	lan := lang.GetLangFromEnv()
	if adapter, ok := errAdapterMapping[errno]; ok {
		if val, ok := adapter.OuterMsg[lan]; ok {
			return val
		}
	}

	return ""
}

// return empty status when don't exist
func GetStatusByErrNo(errno int64) Status {
	if adapter, ok := errAdapterMapping[errno]; ok {
		return adapter.OuterStatus
	}

	return NullStatus
}
