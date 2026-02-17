package model

// 响应码常量
const (
    CodeSuccess = 200

    // 认证相关 401xx
    CodeUnauthorized  = 40101
    CodeTokenExpired  = 40102
    CodeTooManyLogin  = 40103

    // 权限相关 403xx
    CodePermissionDenied = 40301
    CodeOpForbidden      = 40302

    // 请求参数 400xx
    CodeInvalidParam = 40001

    // 资源不存在 404xx
    CodeDocNotFound      = 40401
    CodeResourceNotFound = 40402

    // 限流 429xx
    CodeTooManyRequests = 42901

    // 服务器错误 500xx
    CodeInternalError = 50000
    CodeDatabaseError = 50001
    CodeUnknownError  = 50099
)

// ResponseCode 错误码接口
type ResponseCode interface {
    Code() int
    Message() string
}

// 标准错误码实现
type stdCode struct {
    code int
    msg  string
}

func (c stdCode) Code() int       { return c.code }
func (c stdCode) Message() string { return c.msg }

// 预定义错误码
var (
    ErrInvalidParam      = stdCode{CodeInvalidParam, "invalid request parameter"}
    ErrUnauthorized      = stdCode{CodeUnauthorized, "unauthorized"}
    ErrTokenExpired      = stdCode{CodeTokenExpired, "token expired"}
    ErrPermissionDenied  = stdCode{CodePermissionDenied, "permission denied"}
    ErrDocNotFound       = stdCode{CodeDocNotFound, "document not found"}
    ErrResourceNotFound  = stdCode{CodeResourceNotFound, "resource not found"}
    ErrTooManyRequests   = stdCode{CodeTooManyRequests, "too many requests"}
    ErrInternalError     = stdCode{CodeInternalError, "internal server error"}
    ErrDatabaseError     = stdCode{CodeDatabaseError, "database error"}
)