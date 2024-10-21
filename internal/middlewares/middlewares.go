package middlewares

import (
	"context"
	"net/http"

	"github.com/madcarpet/gophermart/internal/authorization"
	"github.com/madcarpet/gophermart/internal/constants"
	"github.com/madcarpet/gophermart/internal/logger"
	"go.uber.org/zap"
)

type ctxKey string

const UID ctxKey = "uid"

func Authorize(a authorization.Authorizer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//jwt header checking
		jwtHeader := r.Header.Get(constants.CookieToken)
		if jwtHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Access denied"))
			return
		}
		//jwt header checking
		userID, err := a.VerifyToken(jwtHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Access denied"))
			return
		}
		ctx := context.WithValue(r.Context(), UID, userID)
		logger.Log.Debug("next user authorized successfully", zap.String("UserId", userID), zap.String("PATH", r.URL.Path))
		next(w, r.WithContext(ctx))
	}
}
