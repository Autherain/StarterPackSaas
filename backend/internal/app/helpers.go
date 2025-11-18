package app

import (
	"strconv"
	"time"

	"github.com/pascaldekloe/jwt"
)

func (app *Application) NewAuthenticationToken(userID int) (string, time.Time, error) {
	now := time.Now()

	var claims jwt.Claims
	claims.Subject = strconv.Itoa(userID)

	expiry := now.Add(24 * time.Hour)
	claims.Issued = jwt.NewNumericTime(now)
	claims.NotBefore = jwt.NewNumericTime(now)
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = app.Config.BaseURL
	claims.Audiences = []string{app.Config.BaseURL}

	jwt, err := claims.HMACSign(jwt.HS256, []byte(app.Config.JWT.SecretKey))
	return string(jwt), expiry, err
}
