package clerk

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/Lev1ty/lmsysmd/lib/context/value"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
)

func WithHeaderAuthorization(pattern string, next http.Handler) (string, http.Handler) {
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))
	return pattern, clerkhttp.WithHeaderAuthorization()(next)
}

type Middleware struct {
	Configs []Config
}

type Config struct {
	Includes  []string
	Excludes  []string
	RootKey   string
	Allowlist []string
	Denylist  []string
}

func (mi Middleware) Handler(next http.Handler) http.Handler {
	si, err := url.Parse(os.Getenv("CLERK_SIGN_IN_URL"))
	if err != nil {
		log.Fatal(err)
	}
	ru, err := url.Parse(os.Getenv("CLERK_SIGN_IN_REDIRECT_URL"))
	if err != nil {
		log.Fatal(err)
	}
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))
	redirectToSignIn := func(w http.ResponseWriter, r *http.Request) {
		ru2 := *ru
		rq := ru2.Query()
		rq.Add("redirect_url", r.URL.Path)
		ru2.RawQuery = rq.Encode()
		si2 := *si
		sq := si2.Query()
		sq.Add("redirect_url", ru2.String())
		si2.RawQuery = sq.Encode()
		http.Redirect(w, r, si2.String(), http.StatusTemporaryRedirect)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cfg *Config
		for _, c := range mi.Configs {
			if slices.ContainsFunc(c.Includes, func(p string) bool {
				if strings.HasSuffix(p, "/{$}") {
					return strings.TrimSuffix(r.URL.Path, "/") == strings.TrimSuffix(p, "/{$}")
				}
				return strings.HasPrefix(r.URL.Path, p)
			}) && !slices.ContainsFunc(c.Excludes, func(p string) bool {
				if strings.HasSuffix(p, "/{$}") {
					return strings.TrimSuffix(r.URL.Path, "/") == strings.TrimSuffix(p, "/{$}")
				}
				return strings.HasPrefix(r.URL.Path, p)
			}) {
				cfg = &c
				break
			}
		}
		if cfg == nil {
			next.ServeHTTP(w, r)
			return
		}
		if cfg.RootKey != "" {
			if k := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer "); k == cfg.RootKey {
				next.ServeHTTP(w, r)
				return
			}
		}
		s, err := r.Cookie("__session")
		if err != nil {
			slog.InfoContext(r.Context(), "session claims not found", "r", r, "e", err)
			redirectToSignIn(w, r)
			return
		}
		sc, err := jwt.Verify(r.Context(), &jwt.VerifyParams{Leeway: time.Second, Token: s.Value})
		if err != nil {
			slog.InfoContext(r.Context(), "invalid session claims", "r", r, "e", err)
			redirectToSignIn(w, r)
			return
		}
		if len(cfg.Allowlist) > 0 {
			if !slices.Contains(cfg.Allowlist, sc.Subject) {
				slog.InfoContext(r.Context(), "user not allowed", "r", r, "sc", sc)
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}
		if slices.Contains(cfg.Denylist, sc.Subject) || slices.Contains(cfg.Denylist, "*") {
			slog.InfoContext(r.Context(), "user denied", "r", r, "sc", sc)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), value.ClerkSessionClaims, sc)))
	})
}

func (mi Middleware) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		var cfg *Config
		for _, c := range mi.Configs {
			if slices.ContainsFunc(c.Includes, func(p string) bool {
				return strings.HasPrefix(req.Spec().Procedure, p)
			}) && !slices.ContainsFunc(c.Excludes, func(p string) bool {
				return strings.HasPrefix(req.Spec().Procedure, p)
			}) {
				cfg = &c
				break
			}
		}
		if cfg == nil {
			return next(ctx, req)
		}
		if cfg.RootKey != "" {
			if k := strings.TrimPrefix(req.Header().Get("authorization"), "Bearer "); k == cfg.RootKey {
				return next(ctx, req)
			}
		}
		sc, ok := clerk.SessionClaimsFromContext(ctx)
		if !ok {
			slog.InfoContext(ctx, "invalid session claims", "r", req)
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("clerk unauthenticated"))
		}
		if len(cfg.Allowlist) > 0 {
			if !slices.Contains(cfg.Allowlist, sc.Subject) {
				slog.InfoContext(ctx, "user not allowed", "r", req, "sc", sc)
				return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("clerk permission denied"))
			}
		}
		if slices.Contains(cfg.Denylist, sc.Subject) || slices.Contains(cfg.Denylist, "*") {
			slog.InfoContext(ctx, "user denied", "r", req, "sc", sc)
			return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("clerk permission denied"))
		}
		return next(context.WithValue(ctx, value.ClerkSessionClaims, sc), req)
	})
}

func (mi Middleware) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (mi Middleware) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		var cfg *Config
		for _, c := range mi.Configs {
			if slices.ContainsFunc(c.Includes, func(p string) bool {
				return strings.HasPrefix(conn.Spec().Procedure, p)
			}) && !slices.ContainsFunc(c.Excludes, func(p string) bool {
				return strings.HasPrefix(conn.Spec().Procedure, p)
			}) {
				cfg = &c
				break
			}
		}
		if cfg == nil {
			return next(ctx, conn)
		}
		if cfg.RootKey != "" {
			if k := strings.TrimPrefix(conn.RequestHeader().Get("authorization"), "Bearer "); k == cfg.RootKey {
				return next(ctx, conn)
			}
		}
		sc, ok := clerk.SessionClaimsFromContext(ctx)
		if !ok {
			slog.InfoContext(ctx, "invalid session claims", "r", conn.RequestHeader())
			return connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("clerk unauthenticated"))
		}
		if len(cfg.Allowlist) > 0 {
			if !slices.Contains(cfg.Allowlist, sc.Subject) {
				slog.InfoContext(ctx, "user not allowed", "r", conn.RequestHeader(), "sc", sc)
				return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("clerk permission denied"))
			}
		}
		if slices.Contains(cfg.Denylist, sc.Subject) || slices.Contains(cfg.Denylist, "*") {
			slog.InfoContext(ctx, "user denied", "r", conn.RequestHeader(), "sc", sc)
			return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("clerk permission denied"))
		}
		return next(context.WithValue(ctx, value.ClerkSessionClaims, sc), conn)
	})
}
