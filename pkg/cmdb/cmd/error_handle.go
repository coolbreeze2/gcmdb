package cmd

import (
	"fmt"
	"goTool/pkg/cmdb"
	"os"
	"strings"
)

const DefaultErrorExitCode = 1

func CheckError(err error) {
	if err == nil {
		return
	}

	switch err.(type) {
	case cmdb.ResourceNotFoundError:
		msg := fmt.Sprintf("Error from server (NotFound): %s", err.Error())
		fatalErrHandler(msg, DefaultErrorExitCode)
	case cmdb.ResourceValidateError:
		msg := fmt.Sprintf("Error from server (ValidateError): %s", err.Error())
		fatalErrHandler(msg, DefaultErrorExitCode)
	case cmdb.ResourceAlreadyExistError:
		msg := fmt.Sprintf("Error from server (AlreadyExistError): %s", err.Error())
		fatalErrHandler(msg, DefaultErrorExitCode)
	case cmdb.ResourceReferencedError:
		msg := fmt.Sprintf("Error from server (ReferencedError): %s", err.Error())
		fatalErrHandler(msg, DefaultErrorExitCode)
	case cmdb.ServerError:
		msg := fmt.Sprintf("Error from server (UnknowError): %s", err.Error())
		fatalErrHandler(msg, DefaultErrorExitCode)
	default:
		msg := err.Error()
		if !strings.HasPrefix(msg, "error: ") {
			msg = fmt.Sprintf("error: %s", msg)
		}
		fatalErrHandler(msg, DefaultErrorExitCode)
	}
}

func fatalErrHandler(msg string, code int) {
	if len(msg) > 0 {
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(os.Stderr, msg)
	}
	os.Exit(code)
}
