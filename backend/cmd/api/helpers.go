package main

import (
	"strconv"
	"time"

	"github.com/pascaldekloe/jwt"
)

func (app *application) newAuthenticationToken(userID int) (string, time.Time, error) {
	now := time.Now()

	var claims jwt.Claims
	claims.Subject = strconv.Itoa(userID)

	expiry := now.Add(24 * time.Hour)
	claims.Issued = jwt.NewNumericTime(now)
	claims.NotBefore = jwt.NewNumericTime(now)
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = app.config.baseURL
	claims.Audiences = []string{app.config.baseURL}

	jwt, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
	return string(jwt), expiry, err
}
