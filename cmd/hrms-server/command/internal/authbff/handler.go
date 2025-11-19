package authbff

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cube-castle/internal/config"
	"cube-castle/internal/organization/audit"
	reqmw "cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/utils"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BFFHandler struct {
	store        Store
	logger       pkglogger.Logger
	jwtCfg       JWTMintConfig
	devMode      bool
	secureCookie bool
	cookieDomain string
	sessionTTL   time.Duration
	accessTTL    time.Duration
	authMode     string // dev|oidc
	oidc         *OIDCClient
	flows        *AuthFlowStore
	auditor      *audit.AuditLogger
}

func scopedLogger(base pkglogger.Logger, component string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}
	fields := pkglogger.Fields{
		"component": component,
	}
	for k, v := range extra {
		fields[k] = v
	}
	return base.WithFields(fields)
}

func (h *BFFHandler) requestLogger(r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	fields := pkglogger.Fields{}
	for k, v := range extra {
		fields[k] = v
	}
	if action != "" {
		fields["action"] = action
	}
	if r != nil {
		fields["method"] = r.Method
		fields["path"] = r.URL.Path
		fields["requestId"] = reqmw.GetRequestID(r.Context())
	}
	return h.logger.WithFields(fields)
}

func NewBFFHandler(baseLogger pkglogger.Logger, devMode bool, auditor *audit.AuditLogger, jwtConfig *config.JWTConfig) *BFFHandler {
	if baseLogger == nil {
		baseLogger = pkglogger.NewNoopLogger()
	}
	if jwtConfig == nil {
		panic("jwt config is required")
	}
	secure := os.Getenv("SECURE_COOKIES") == "true"
	domain := os.Getenv("COOKIE_DOMAIN")
	ttlStr := os.Getenv("SESSION_TTL")
	ttl := 30 * 24 * time.Hour
	if ttlStr != "" {
		if d, err := time.ParseDuration(ttlStr); err == nil {
			ttl = d
		}
	}
	accessTTL := 10 * time.Minute
	if v := os.Getenv("ACCESS_TOKEN_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			accessTTL = d
		}
	}
	componentLogger := scopedLogger(baseLogger, "authBFF", pkglogger.Fields{"module": "authbff"})

	jwMintAlg := strings.ToUpper(strings.TrimSpace(jwtConfig.MintAlgorithm))
	if jwMintAlg == "" {
		jwMintAlg = strings.ToUpper(strings.TrimSpace(jwtConfig.Algorithm))
	}
	if jwMintAlg == "" {
		jwMintAlg = "RS256"
	}
	if jwMintAlg != "RS256" {
		componentLogger.WithFields(pkglogger.Fields{"alg": jwMintAlg}).Error("JWT_MINT_ALG must be RS256")
		panic("JWT_MINT_ALG must be configured as RS256")
	}

	h := &BFFHandler{
		store:        NewInMemoryStore(),
		logger:       componentLogger,
		jwtCfg:       JWTMintConfig{Secret: jwtConfig.Secret, Issuer: jwtConfig.Issuer, Audience: jwtConfig.Audience, Alg: jwMintAlg},
		devMode:      devMode,
		secureCookie: secure,
		cookieDomain: domain,
		sessionTTL:   ttl,
		accessTTL:    accessTTL,
		authMode:     os.Getenv("AUTH_MODE"),
		flows:        NewAuthFlowStore(),
		auditor:      auditor,
	}
	// 可选：加载RS256私钥用于对外签名（BFF自签短期token）
	if h.jwtCfg.Alg == "RS256" {
		if rawPath := jwtConfig.PrivateKeyPath; rawPath != "" {
			safePath, err := sanitizeAbsolutePath(rawPath)
			if err != nil {
				h.logger.WithFields(pkglogger.Fields{"path": rawPath, "error": err}).Error("RS256 private key path invalid")
				panic(fmt.Errorf("RS256私钥路径无效: %w", err))
			}
			// #nosec G304 -- safePath 已验证为绝对路径，由运维配置提供
			if b, err := os.ReadFile(safePath); err == nil {
				if pk, err := ParseRSAPrivateKeyFromPEM(b); err == nil {
					h.jwtCfg.PrivateKey = pk
					h.jwtCfg.PrivateKeyPEM = b
				} else {
					h.logger.WithFields(pkglogger.Fields{"path": safePath, "error": err}).Error("failed to parse RS256 private key")
					panic(fmt.Errorf("解析RS256私钥失败: %w", err))
				}
			} else {
				h.logger.WithFields(pkglogger.Fields{"path": safePath, "error": err}).Error("failed to read RS256 private key")
				panic(fmt.Errorf("读取RS256私钥失败: %w", err))
			}
		}
		if h.jwtCfg.PrivateKey == nil {
			h.logger.Error("RS256 enabled but JWT_PRIVATE_KEY_PATH not configured")
			panic("RS256 enabled but missing private key")
		}
		h.jwtCfg.KeyID = jwtConfig.KeyID
		if h.jwtCfg.KeyID == "" {
			h.jwtCfg.KeyID = "bff-key-1"
		}
	}
	// 会话存储优先 Redis（可选）
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		if rs, err := NewRedisStore(addr, ttl); err == nil {
			h.store = rs
			h.logger.WithFields(pkglogger.Fields{"redisAddr": addr}).Info("BFF session store initialized")
		} else {
			h.logger.WithFields(pkglogger.Fields{"error": err}).Warn("Redis init failed, fallback to in-memory store")
		}
	}

	// 初始化 OIDC 客户端（若配置完整）
	oidc := NewOIDCClientFromEnv()
	if oidc.IsConfigured() {
		h.oidc = oidc
		if _, err := h.oidc.Discover(); err != nil {
			h.logger.WithFields(pkglogger.Fields{"error": err}).Warn("OIDC discovery failed")
		} else {
			h.logger.WithFields(pkglogger.Fields{"issuer": oidc.cfg.Issuer, "clientId": oidc.cfg.ClientID}).Info("OIDC configured")
		}
	} else {
		h.logger.Info("OIDC not fully configured; simulation mode may be used")
	}
	return h
}

func (h *BFFHandler) SetupRoutes(r chi.Router) {
	r.Get("/auth/login", h.handleLogin)
	r.Get("/auth/callback", h.handleCallback)
	r.Get("/auth/session", h.handleSession)
	r.Post("/auth/refresh", h.handleRefresh)
	r.Post("/auth/logout", h.handleLogout)
	r.Get("/.well-known/oidc", h.handleWellKnown)
	r.Get("/.well-known/jwks.json", h.handleJWKS)
	// 浏览器发起的IdP退出联动（302跳转），用于前端在需要彻底注销IdP会话时调用
	r.Get("/auth/logout", h.handleLogoutRedirect)
}

func (h *BFFHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	logger := h.requestLogger(r, "handleLogin", pkglogger.Fields{
		"redirect": redirect,
		"devMode":  h.devMode,
		"authMode": h.authMode,
	})

	// TODO-TEMPORARY: OIDC集成前的模拟登录，DEV 或 OIDC_SIMULATE 下启用，后续替换为真实IdP跳转。截止：2025-10-17
	if (h.devMode || os.Getenv("OIDC_SIMULATE") == "true") && !h.isOIDCEnabled() {
		sess := h.newSimulatedSession()
		h.store.Set(sess)
		h.setSessionCookies(w, sess)
		logger.WithFields(pkglogger.Fields{"sessionId": sess.ID, "tenantId": sess.TenantID}).Info("simulated login session established")
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	if !h.isOIDCEnabled() {
		if err := utils.WriteError(w, http.StatusNotImplemented, "OIDC_NOT_CONFIGURED", "OIDC未配置，请设置 OIDC_ISSUER/CLIENT_ID/REDIRECT_URI", reqmw.GetRequestID(r.Context()), nil); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to write OIDC_NOT_CONFIGURED response")
		}
		logger.Warn("OIDC not configured; login aborted")
		return
	}
	// 构建授权请求
	doc, err := h.oidc.Discover()
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("OIDC discovery failed during login")
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), "OIDC discovery failed: "+err.Error())
		return
	}
	state := uuid.NewString()
	nonce := randomString(16)
	codeVerifier := randomString(64)
	challenge := BuildCodeChallenge(codeVerifier)
	// 保存flow（10分钟）
	h.flows.Set(&AuthFlowState{State: state, Nonce: nonce, CodeVerifier: codeVerifier, RedirectPath: redirect, CreatedAt: time.Now(), ExpiresAt: time.Now().Add(10 * time.Minute)})
	authURL, err := h.oidc.BuildAuthURL(doc.AuthorizationEndpoint, state, nonce, challenge, redirect)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to build OIDC authorization URL")
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), "build auth url failed: "+err.Error())
		return
	}
	logger.WithFields(pkglogger.Fields{"state": state}).Info("redirecting to OIDC authorization endpoint")
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *BFFHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "/"
	}
	logger := h.requestLogger(r, "handleCallback", pkglogger.Fields{
		"redirect": redirect,
	})
	if (h.devMode || os.Getenv("OIDC_SIMULATE") == "true") && !h.isOIDCEnabled() {
		sess := h.newSimulatedSession()
		h.store.Set(sess)
		h.setSessionCookies(w, sess)
		logger.WithFields(pkglogger.Fields{"sessionId": sess.ID}).Info("completed simulated callback")
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	if !h.isOIDCEnabled() {
		if err := utils.WriteError(w, http.StatusNotImplemented, "OIDC_NOT_CONFIGURED", "未实现回调：请先配置 OIDC", reqmw.GetRequestID(r.Context()), nil); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to write OIDC callback error response")
		}
		logger.Warn("OIDC not configured for callback")
		return
	}
	// 校验参数
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	logger = logger.WithFields(pkglogger.Fields{
		"state": state,
	})
	if code == "" || state == "" {
		logger.Warn("callback missing code or state parameters")
		_ = utils.WriteError(w, http.StatusBadRequest, "INVALID_CALLBACK", "缺少code/state", reqmw.GetRequestID(r.Context()), nil)
		return
	}
	flow, ok := h.flows.Get(state)
	if !ok {
		logger.Warn("state not found or expired during callback")
		_ = utils.WriteError(w, http.StatusUnauthorized, "STATE_EXPIRED", "state已过期或无效", reqmw.GetRequestID(r.Context()), nil)
		return
	}
	// 令牌交换
	doc, err := h.oidc.Discover()
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("OIDC discovery failed during callback")
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), "discovery failed: "+err.Error())
		return
	}
	tr, err := h.oidc.ExchangeCode(doc.TokenEndpoint, code, flow.CodeVerifier)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("OIDC token exchange failed")
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), "token exchange failed: "+err.Error())
		h.logAuthError(r, "OIDC_CALLBACK", "TOKEN_EXCHANGE_FAILED", err.Error(), map[string]any{"state": state})
		return
	}
	// 校验并解析ID Token（开发阶段弱校验）
	claims, err := h.oidc.ValidateIDToken(tr.IDToken, flow.Nonce)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("OIDC id token invalid")
		_ = utils.WriteError(w, http.StatusUnauthorized, "ID_TOKEN_INVALID", err.Error(), reqmw.GetRequestID(r.Context()), nil)
		h.logAuthError(r, "OIDC_CALLBACK", "ID_TOKEN_INVALID", err.Error(), map[string]any{"state": state})
		return
	}
	// 构建会话
	tenant := getStringClaim(claims, "tenantId")
	if tenant == "" {
		tenant = getStringClaim(claims, "tenant_id")
	}
	if tenant == "" {
		tenant = os.Getenv("DEFAULT_TENANT_ID")
	}
	if tenant == "" {
		tenant = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	}
	userID := getStringClaim(claims, "sub")
	userName := getStringClaim(claims, "name")
	userEmail := getStringClaim(claims, "email")
	// 角色/作用域（可选）
	var scopes []string
	if s := getStringClaim(claims, "scope"); s != "" {
		scopes = strings.Fields(s)
	}
	sess := &Session{
		ID:         uuid.NewString(),
		UserID:     userID,
		UserName:   userName,
		UserEmail:  userEmail,
		TenantID:   tenant,
		Roles:      []string{},
		Scopes:     scopes,
		RefreshTok: tr.RefreshToken,
		IDToken:    tr.IDToken,
		CreatedAt:  time.Now().UTC(),
		LastUsedAt: time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(h.sessionTTL),
	}
	h.store.Set(sess)
	h.setSessionCookies(w, sess)
	// 清理flow
	h.flows.Delete(state)
	h.logAuthSuccess(r, tenant, userID, "LOGIN", map[string]any{"scopes": scopes})
	logger.WithFields(pkglogger.Fields{
		"userId":   userID,
		"tenantId": tenant,
		"scopes":   scopes,
	}).Info("OIDC login success")
	// 回跳到发起页
	target := redirect
	if flow.RedirectPath != "" {
		target = flow.RedirectPath
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *BFFHandler) handleSession(w http.ResponseWriter, r *http.Request) {
	sid, _ := r.Cookie("sid")
	logger := h.requestLogger(r, "handleSession", pkglogger.Fields{
		"sessionId": getCookieValue(sid),
	})
	if sid == nil || sid.Value == "" {
		_ = utils.WriteUnauthorized(w, reqmw.GetRequestID(r.Context()))
		return
	}
	sess, ok := h.store.Get(sid.Value)
	if !ok || time.Now().After(sess.ExpiresAt) {
		logger.Warn("session expired or missing")
		_ = utils.WriteError(w, http.StatusUnauthorized, "SESSION_EXPIRED", "会话已过期", reqmw.GetRequestID(r.Context()), nil)
		return
	}
	token, exp, err := MintAccessToken(h.jwtCfg, sess, h.accessTTL)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to mint access token in session handler")
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), err.Error())
		return
	}
	data := map[string]interface{}{
		"accessToken": token,
		"expiresIn":   exp - time.Now().UTC().Unix(),
		"tenantId":    sess.TenantID,
		"user":        map[string]interface{}{"id": sess.UserID, "name": sess.UserName, "email": sess.UserEmail},
		"scopes":      sess.Scopes,
	}
	_ = utils.WriteSuccess(w, data, "Session active", reqmw.GetRequestID(r.Context()))
	logger.WithFields(pkglogger.Fields{"userId": sess.UserID}).Info("session validated")
}

func (h *BFFHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	logger := h.requestLogger(r, "handleRefresh", nil)
	if !h.checkCSRF(w, r) {
		logger.Warn("csrf validation failed during refresh")
		return
	}
	sid, _ := r.Cookie("sid")
	if sid == nil || sid.Value == "" {
		_ = utils.WriteUnauthorized(w, reqmw.GetRequestID(r.Context()))
		return
	}
	sess, ok := h.store.Get(sid.Value)
	if !ok || time.Now().After(sess.ExpiresAt) {
		logger.Warn("session expired during refresh")
		_ = utils.WriteError(w, http.StatusUnauthorized, "SESSION_EXPIRED", "会话已过期", reqmw.GetRequestID(r.Context()), nil)
		h.logAuthError(r, "REFRESH", "SESSION_EXPIRED", "session expired", nil)
		return
	}

	// 若启用 OIDC，优先使用 refresh token 调用 IdP 轮换（若失败，按 419 处理）
	if h.isOIDCEnabled() && sess.RefreshTok != "" {
		doc, err := h.oidc.Discover()
		if err == nil {
			if tr, err2 := h.oidc.RefreshWithToken(doc.TokenEndpoint, sess.RefreshTok); err2 == nil {
				// 刷新成功：更新服务端 refresh token（rotation）
				if tr.RefreshToken != "" {
					logger.Info("OIDC refresh token rotated")
					sess.RefreshTok = tr.RefreshToken
				}
				sess.LastUsedAt = time.Now().UTC()
				h.store.Set(sess)
			} else {
				// 失败：判定会话失效（401路径）
				logger.WithFields(pkglogger.Fields{"error": err2}).Warn("OIDC refresh token flow failed")
				_ = utils.WriteError(w, http.StatusUnauthorized, "SESSION_EXPIRED", "会话已过期或刷新失败", reqmw.GetRequestID(r.Context()), map[string]string{"reason": err2.Error()})
				h.logAuthError(r, "REFRESH", "REFRESH_FAILED", err2.Error(), nil)
				return
			}
		} else {
			logger.WithFields(pkglogger.Fields{"error": err}).Warn("OIDC discovery failed during refresh")
		}
	}

	token, exp, err := MintAccessToken(h.jwtCfg, sess, h.accessTTL)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to mint access token during refresh")
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), err.Error())
		return
	}
	data := map[string]interface{}{
		"accessToken": token,
		"expiresIn":   exp - time.Now().UTC().Unix(),
	}
	_ = utils.WriteSuccess(w, data, "Access token refreshed", reqmw.GetRequestID(r.Context()))
	logger.WithFields(pkglogger.Fields{"userId": sess.UserID}).Info("session refreshed")
	h.logAuthSuccess(r, sess.TenantID, sess.UserID, "REFRESH", nil)
}

func (h *BFFHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	logger := h.requestLogger(r, "handleLogout", nil)
	if !h.checkCSRF(w, r) {
		logger.Warn("csrf validation failed during logout")
		return
	}
	sid, _ := r.Cookie("sid")
	var sess *Session
	if sid != nil && sid.Value != "" {
		if s, ok := h.store.Get(sid.Value); ok {
			sess = s
		}
		h.store.Delete(sid.Value)
	}
	// 清除Cookie
	h.clearCookie(w, "sid")
	h.clearCookie(w, "csrf")
	// POST 语义保持 204；若需要 IdP 退出，由前端再调用 GET /auth/logout 以跳转至 IdP
	h.logAuthSuccess(r, getSessionTenant(sess), getSessionUser(sess), "LOGOUT", nil)
	logger.WithFields(pkglogger.Fields{"userId": getSessionUser(sess)}).Info("session logout completed")
	w.WriteHeader(http.StatusNoContent)
}

// handleLogoutRedirect 触发RP-initiated IdP Logout
func (h *BFFHandler) handleLogoutRedirect(w http.ResponseWriter, r *http.Request) {
	logger := h.requestLogger(r, "handleLogoutRedirect", nil)
	// 读取会话，用于获取 id_token_hint
	sid, _ := r.Cookie("sid")
	var sess *Session
	if sid != nil && sid.Value != "" {
		if s, ok := h.store.Get(sid.Value); ok {
			sess = s
		}
	}
	// 清理本地会话与Cookie
	if sid != nil && sid.Value != "" {
		h.store.Delete(sid.Value)
	}
	h.clearCookie(w, "sid")
	h.clearCookie(w, "csrf")

	if !h.isOIDCEnabled() || sess == nil || sess.IDToken == "" {
		// 无法联动IdP，回到首页或redirect
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/"
		}
		http.Redirect(w, r, redirect, http.StatusFound)
		logger.WithFields(pkglogger.Fields{"redirect": redirect}).Info("fallback logout redirect without OIDC")
		return
	}
	// 发现文档
	doc, err := h.oidc.Discover()
	if err != nil || doc.EndSessionEndpoint == "" {
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/"
		}
		http.Redirect(w, r, redirect, http.StatusFound)
		logger.WithFields(pkglogger.Fields{"redirect": redirect}).Warn("OIDC end session endpoint unavailable")
		return
	}
	// 构建 IdP 退出URL：id_token_hint + post_logout_redirect_uri
	u, _ := url.Parse(doc.EndSessionEndpoint)
	q := u.Query()
	q.Set("id_token_hint", sess.IDToken)
	postLogout := h.oidc.cfg.PostLogoutURI
	if v := r.URL.Query().Get("redirect"); v != "" {
		postLogout = v
	}
	if postLogout != "" {
		q.Set("post_logout_redirect_uri", postLogout)
	}
	q.Set("state", uuid.NewString())
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
	h.logAuthSuccess(r, sess.TenantID, sess.UserID, "LOGOUT_RP", map[string]any{"end_session_endpoint": u.String()})
	logger.WithFields(pkglogger.Fields{"endSessionEndpoint": u.String()}).Info("redirected to OIDC end session endpoint")
}

// 审计工具函数
func (h *BFFHandler) logAuthSuccess(r *http.Request, tenantID string, actorID string, action string, extra map[string]any) {
	if h.auditor == nil {
		return
	}
	tid, _ := uuid.Parse(tenantID)
	_ = h.auditor.LogEvent(r.Context(), &audit.AuditEvent{
		TenantID:     tid,
		EventType:    audit.EventTypeAuth,
		ResourceType: audit.ResourceTypeUser,
		ResourceID:   actorID,
		ActorID:      actorID,
		ActorType:    audit.ActorTypeUser,
		ActionName:   action,
		RequestID:    reqmw.GetRequestID(r.Context()),
		Success:      true,
		AfterData:    extra,
	})
}

func (h *BFFHandler) logAuthError(r *http.Request, action string, code string, message string, reqData map[string]any) {
	if h.auditor == nil {
		return
	}
	tid := uuid.Nil
	if v := r.Header.Get("X-Tenant-ID"); v != "" {
		tid, _ = uuid.Parse(v)
	}
	_ = h.auditor.LogError(r.Context(), tid, audit.ResourceTypeSystem, "auth", action, "system", reqmw.GetRequestID(r.Context()), code, message, reqData)
}

func getSessionTenant(sess *Session) string {
	if sess != nil {
		return sess.TenantID
	}
	return ""
}
func getSessionUser(sess *Session) string {
	if sess != nil {
		return sess.UserID
	}
	return ""
}

func getCookieValue(cookie *http.Cookie) string {
	if cookie == nil {
		return ""
	}
	return cookie.Value
}

func (h *BFFHandler) handleWellKnown(w http.ResponseWriter, r *http.Request) {
	if !h.isOIDCEnabled() {
		_ = utils.WriteError(w, http.StatusNotImplemented, "OIDC_NOT_CONFIGURED", "OIDC未配置", reqmw.GetRequestID(r.Context()), nil)
		return
	}
	doc, err := h.oidc.Discover()
	if err != nil {
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), err.Error())
		return
	}
	// 契约要求：字段使用 camelCase，且直接返回对象（非统一信封）
	data := map[string]any{
		"issuer":                doc.Issuer,
		"authorizationEndpoint": doc.AuthorizationEndpoint,
		"tokenEndpoint":         doc.TokenEndpoint,
		"endSessionEndpoint":    doc.EndSessionEndpoint,
		"jwksUri":               doc.JWKSURI,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *BFFHandler) handleJWKS(w http.ResponseWriter, r *http.Request) {
	b, err := h.buildJWKS()
	if err != nil {
		_ = utils.WriteInternalError(w, reqmw.GetRequestID(r.Context()), err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (h *BFFHandler) isOIDCEnabled() bool {
	return h.oidc != nil && h.oidc.IsConfigured()
}

func getStringClaim(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func (h *BFFHandler) newSimulatedSession() *Session {
	// 默认租户
	tenant := os.Getenv("DEFAULT_TENANT_ID")
	if tenant == "" {
		tenant = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	}
	sid := uuid.NewString()
	now := time.Now().UTC()
	return &Session{
		ID:         sid,
		UserID:     "dev-user",
		UserName:   "开发用户",
		UserEmail:  "dev@example.com",
		TenantID:   tenant,
		Roles:      []string{"ADMIN", "HR_STAFF"},
		Scopes:     []string{"org:read", "org:update", "org:read:history"},
		RefreshTok: randomString(32),
		CreatedAt:  now,
		LastUsedAt: now,
		ExpiresAt:  now.Add(h.sessionTTL),
	}
}

func (h *BFFHandler) setSessionCookies(w http.ResponseWriter, sess *Session) {
	// sid HttpOnly
	sid := &http.Cookie{Name: "sid", Value: sess.ID, Path: "/", HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode, Expires: sess.ExpiresAt}
	if h.cookieDomain != "" {
		sid.Domain = h.cookieDomain
	}
	http.SetCookie(w, sid)
	// csrf 非HttpOnly
	csrf := &http.Cookie{Name: "csrf", Value: randomString(24), Path: "/", HttpOnly: false, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode, Expires: sess.ExpiresAt}
	if h.cookieDomain != "" {
		csrf.Domain = h.cookieDomain
	}
	http.SetCookie(w, csrf)
}

func (h *BFFHandler) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1, HttpOnly: name == "sid", Secure: h.secureCookie, SameSite: http.SameSiteLaxMode, Domain: h.cookieDomain})
}

func (h *BFFHandler) checkCSRF(w http.ResponseWriter, r *http.Request) bool {
	cookie, _ := r.Cookie("csrf")
	header := r.Header.Get("X-CSRF-Token")
	if cookie == nil || cookie.Value == "" || header == "" || cookie.Value != header {
		_ = utils.WriteError(w, http.StatusUnauthorized, "CSRF_CHECK_FAILED", "CSRF校验失败", reqmw.GetRequestID(r.Context()), map[string]string{"header": header})
		return false
	}
	return true
}

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func sanitizeAbsolutePath(raw string) (string, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return "", fmt.Errorf("路径不能为空")
	}
	clean = filepath.Clean(clean)
	if !filepath.IsAbs(clean) {
		return "", fmt.Errorf("路径必须为绝对路径: %s", raw)
	}
	return clean, nil
}

// RequestIDFromContext 兼容工具（避免直接引用私有中间件）
// 使用 utils 中的方法（若存在），否则回退
// 注：当前 utils 中没有导出获取requestID函数，这里做兼容处理
// TODO-TEMPORARY: 后续从统一中间件导出获取 requestID 的方法（2025-09-25 前）
var _ = strconv.Itoa // 保持导入
