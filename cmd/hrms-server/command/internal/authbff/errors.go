package authbff

// 统一错误码（与企业级响应一致）
const (
    ErrOIDCNotConfigured   = "OIDC_NOT_CONFIGURED"
    ErrInvalidCallback     = "INVALID_CALLBACK"
    ErrStateExpired        = "STATE_EXPIRED"
    ErrIDTokenInvalid      = "ID_TOKEN_INVALID"
    ErrSessionExpired      = "SESSION_EXPIRED"
    ErrCSRFCheckFailed     = "CSRF_CHECK_FAILED"
    ErrRefreshFailed       = "REFRESH_FAILED"
)

