package core

import "fmt"

type HelloAgentsError struct {
	Code    string
	Message string
	Cause   error
}

func (e *HelloAgentsError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *HelloAgentsError) Unwrap() error {
	return e.Cause
}

func NewError(code, message string) *HelloAgentsError {
	return &HelloAgentsError{Code: code, Message: message}
}

func NewErrorWithCause(code, message string, cause error) *HelloAgentsError {
	return &HelloAgentsError{Code: code, Message: message, Cause: cause}
}

var (
	ErrLLM      = NewError("LLM_ERROR", "LLM相关错误")
	ErrAgent    = NewError("AGENT_ERROR", "Agent相关错误")
	ErrConfig   = NewError("CONFIG_ERROR", "配置相关错误")
	ErrTool     = NewError("TOOL_ERROR", "工具相关错误")
	ErrDatabase = NewError("DATABASE_ERROR", "数据库相关错误")
)
