package tip

import (
	"github.com/si9ma/KillOJ-common/lang"
	"golang.org/x/text/language"
)

type Tip map[language.Tag]string

var (
	CompileTimeOutTip = Tip{
		language.Chinese: "编译时间太长，请检查代码!",
		language.English: "compile too long, Please check your code!",
	}

	RunTimeOutTip = Tip{
		language.Chinese: "程序运行超时，请检查代码!",
		language.English: "run program timeout, Please check your code!",
	}

	SystemErrorTip = Tip{
		language.Chinese: "糟糕，判题机异常，请向管理员报告异常。",
		language.English: "Oops, something has gone wrong with the judger. Please report this to administrator.",
	}

	RuntimeErrorTip = Tip{
		language.Chinese: "运行时错误，请检查代码!",
		language.English: "Runtime error, Please check your code!",
	}

	CompileErrorTip = Tip{
		language.Chinese: "编译错误,请检查代码!",
		language.English: "compile fail,Please check your code!",
	}

	WrongAnswerErrorTip = Tip{
		language.Chinese: "结果错误!",
		language.English: "Wrong answer!",
	}

	OOMErrorTip = Tip{
		language.Chinese: "超出内存使用限制!",
		language.English: "Memory Limit Exceeded!",
	}

	BadSysErrorTip = Tip{
		language.Chinese: "非法系统调用!",
		language.English: "Illegal system call!",
	}

	NoEnoughPidErrorTip = Tip{
		language.Chinese: "超出PID最大允许值限制!",
		language.English: "No Enough PID!",
	}

	JavaSecurityManagerErrorTip = Tip{
		language.Chinese: "非法Java操作!",
		language.English: "Illegal Java operation!",
	}

	CompileSuccessTip = Tip{
		language.Chinese: "编译成功",
		language.English: "compile success",
	}

	RunSuccessTip = Tip{
		language.Chinese: "结果正确",
		language.English: "Accepted",
	}
)

func (t Tip) String() string {
	lan := lang.GetLangFromEnv()
	if val, ok := t[lan]; ok {
		return val
	}

	// if not found ,return en version
	return t[language.English]
}

func (t Tip) String2Lang(lang language.Tag) string {
	if val, ok := t[lang]; ok {
		return val
	}

	// if not found ,return english version
	return t[language.English]
}
