package exitcode

import (
	"fmt"
	"syscall"
)

// get Signal from exit code
// refer: http://www.tldp.org/LDP/abs/html/exitcodes.html
func GetSignal(exitcode int) (syscall.Signal,error) {
	// max signal value = 31
	if exitcode > 128 && exitcode <= 159 {
		return syscall.Signal(exitcode - 128),nil
	}

	return syscall.SIGABRT,fmt.Errorf("the exit code is not from signal")
}
