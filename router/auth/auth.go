package auth

import (
  "context"
  "fmt"
  "log"
  "net/http"
  "time"

  "github.com/go-chi/jwtauth"
)

// JWT code

var tokenAuth *jwtauth.JWTAuth

func GetTokenAuth() *jwtauth.JWTAuth {
  return tokenAuth
}

// InitializeJWT creates the internal state for JWT authentification
func InitializeJWT(private_signing_key string) {
  tokenAuth = jwtauth.New("HS256", []byte(private_signing_key), nil)

  // For debugging/example purposes, we generate and print
  // a sample jwt token with claims `user_id:123` here:
  claims := CreateClaimsForUserID(1)
  tokenString, _ := EncodeClaims(claims)
  fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
  fmt.Printf("curl -H\"Authorization: BEARER \"  -i -X  %s\n\n", tokenString)
}

func CreateClaimsForUserID(userID int) jwtauth.Claims {
  // TODO: for now we kick out the user every 15min ;-)
  // https://security.stackexchange.com/q/119371/167975
  claims := jwtauth.Claims{
    "login_id": userID,                             // the user-id from login
    "exp":      jwtauth.ExpireIn(15 * time.Minute), // expiry
    "iat":      time.Now().UTC().Unix(),            // issued at
    // "exp": jwtauth.EpochNow() - 1000,
  }
  return claims
}

// EncodeClaims wraps the function to encode claims.
// We mainly use `login_id` here only to store the user-id from the
// authentificated user to keep the request small.
func EncodeClaims(claims jwtauth.Claims) (tokenString string, err error) {
  _, tokenString, err = tokenAuth.Encode(claims)
  if err != nil {
    // TODO: should not happen (What is a proper error handling here?)
    panic(err)
  }
  return tokenString, err
}

// AuthenticatorCtx is a custom middleware which handles the entire authentification.
// It tests whether the JW-token is valid and checks whether the claims contain
// an `login_id` identifying the user doing the api-request.
// NOTE: We do not check whether the `login_id` leads to a valid user in the Table.
func AuthenticatorCtx(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    token, claims, err := jwtauth.FromContext(r.Context())

    if err != nil {
      http.Error(w, http.StatusText(401), 401)
      return
    }

    if token == nil || !token.Valid {
      http.Error(w, http.StatusText(401), 401)
      return
    }
    // Token is authenticated
    login_id, ok := claims["login_id"].(float64)

    if !ok {
      http.Error(w, http.StatusText(401), 401)
      return
    }

    // claim is also ok
    log.Println("api-call with login_id:", login_id)

    ctx := context.WithValue(r.Context(), "login_id", int(login_id))
    next.ServeHTTP(w, r.WithContext(ctx))

  })
}
