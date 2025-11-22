package auth

import (
    "context"
    "errors"
    "net/http"
    "os"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

// AuthContext is attached to request contexts for handlers to access authenticated user info.
type AuthContext struct {
    UserID   string
    Username string
    Roles    []string
}

type authContextKey struct{}

// Claims mirrors the structure used in stakeholders-service (uid, username, roles)
type Claims struct {
    UserID   string   `json:"uid"`
    Username string   `json:"username"`
    Roles    []string `json:"roles"`
    jwt.RegisteredClaims
}

func jwtSecret() []byte {
    sec := os.Getenv("JWT_SECRET")
    if sec == "" {
        sec = "dev-secret-change-me"
    }
    return []byte(sec)
}

func ParseToken(tokenStr string) (*Claims, error) {
    parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
    token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
        return jwtSecret(), nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}

// JWTAuthMiddleware validates Authorization bearer tokens and injects AuthContext into request context.
func JWTAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        h := r.Header.Get("Authorization")
        if !strings.HasPrefix(h, "Bearer ") {
            http.Error(w, "missing bearer token", http.StatusUnauthorized)
            return
        }
        token := strings.TrimPrefix(h, "Bearer ")
        claims, err := ParseToken(token)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }
        // optional: check exp etc. jwt lib already validated registered claims
        ctx := context.WithValue(r.Context(), authContextKey{}, &AuthContext{
            UserID:   claims.UserID,
            Username: claims.Username,
            Roles:    claims.Roles,
        })
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetAuth retrieves AuthContext from request, or nil if unauthenticated
func GetAuth(r *http.Request) *AuthContext {
    v := r.Context().Value(authContextKey{})
    if v == nil {
        return nil
    }
    return v.(*AuthContext)
}
