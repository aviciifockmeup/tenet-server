package model

// Response 统一响应结构
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(data interface{}) Response {
    return Response{
        Code:    CodeSuccess,
        Message: "success",
        Data:    data,
    }
}

// Error 错误响应
func Error(code int, message string) Response {
    return Response{
        Code:    code,
        Message: message,
    }
}

// ErrorWithCode 使用错误码的响应
func ErrorWithCode(code ResponseCode) Response {
    return Response{
        Code:    code.Code(),
        Message: code.Message(),
    }
}