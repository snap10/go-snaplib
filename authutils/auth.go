package authutils

import (
	"context"
	"crypto/rsa"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/snap10/go-snaplib/httputils"
)

// Helper is the type for the object we need in that library
type Helper struct {
	PrivKeyPathFile string
	PubKeyPathFile  string
	TokenAudience   string
	TokenIssuer     string
	verifyKey       *rsa.PublicKey
	signKey         *rsa.PrivateKey
}

var helper Helper

// Private key for signing and public key for verification
var (
//verifyKey, signKey []byte

)

// InitKeys Read the key files before starting http handlers
func (h *Helper) InitKeys() {
	if h.PrivKeyPathFile != "" {
		signBytes, err := ioutil.ReadFile(h.PrivKeyPathFile)
		if err != nil {
			log.Printf("[initKeys]: %s\n", err)
		}
		h.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
		if err != nil {
			log.Printf("[initKeys]: %s\n", err)
		}
	}
	if h.PubKeyPathFile != "" {
		verifyBytes, err := ioutil.ReadFile(h.PubKeyPathFile)
		if err != nil {
			log.Fatalf("[initKeys]: %s\n", err)
		}

		h.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
		if err != nil {
			log.Fatalf("[initKeys]: %s\n", err)
		}
	}
}

const (
	// ID_TOKEN
	ID_TOKEN = iota // 0
	// REFRESH_TOKEN
	REFRESH_TOKEN // 1
	// ACCESS_TOKEN
	ACCESS_TOKEN // 2
)

// AppClaims provides custom claim for JWT
type AppClaims struct {
	jwt.StandardClaims
	Role string     `json:"role"`
	Type string     `json:"typ"`
	User *TokenUser `json:"user,omitempty"`
}

// TokenUser is the type used for the "user" claim in the token
type TokenUser struct {
	UserName           string `json:"username,omitempty"`
	Email              string `json:"email,omitempty"`
	ProfilePicturePath string `json:"profile_picture_path,omitempty"`
	CreatedAt          int64  `json:"created_at,omitempty"`
	ModifiedAt         int64  `json:"modified_at,omitempty"`
}

// GenerateIDToken returns an id_token with a claim "user" containing the given parameters
func (h *Helper) GenerateIDToken(id, role, username, email, picturepath string, createdAt, modifiedAt int64) (string, error) {
	return h.generateToken(ID_TOKEN, time.Hour*24*30, id, role, username, email, picturepath, createdAt, modifiedAt)
}

// GenerateRefreshToken returns an refresh_token
func (h *Helper) GenerateRefreshToken(id, role string) (string, error) {
	return h.generateToken(REFRESH_TOKEN, time.Hour*24*30, id, role, "", "", "", 0, 0)
}

// GenerateAccessToken returns an access_token
func (h *Helper) GenerateAccessToken(id, role string) (string, int64, error) {
	exp := time.Second * 3600
	token, err := h.generateToken(ACCESS_TOKEN, exp, id, role, "", "", "", 0, 0)
	return token, int64(exp.Seconds()), err
}

// GenerateBearer generates a new JWT token
func (h *Helper) generateToken(tokentype int, expiresInSec time.Duration, id, role, username, email, picturepath string, createdAt, modifiedAt int64) (string, error) {
	// Create the Claims
	claims := AppClaims{
		StandardClaims: jwt.StandardClaims{
			Audience: helper.TokenAudience,
			Subject:  id,
			IssuedAt: time.Now().Unix(),
			//1Day
			ExpiresAt: time.Now().Add(expiresInSec).Unix(),
			Issuer:    helper.TokenIssuer,
		},
		Role: role,
	}
	switch tokentype {
	case ID_TOKEN:
		claims.Type = "id_token"
		claims.User = &TokenUser{username, email, picturepath, createdAt, modifiedAt}
	case REFRESH_TOKEN:
		claims.Type = "refresh"
	case ACCESS_TOKEN:
		claims.Type = "bearer"
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ss, err := token.SignedString(h.signKey)
	if err != nil {
		return "", err
	}
	return ss, nil
}

func HandleHttpError(w http.ResponseWriter, message string, code int) {
	httputils.DisplayAppError(w, nil, message, code)
}

// Authenticate Middleware for validating JWT tokens
func (h *Helper) Authenticate(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	h.authenticateWithErrHandler(w, r, next, HandleHttpError)
}

func (h *Helper) authenticateWithErrHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc, errHandler func(http.ResponseWriter, string, int)) {

	// Get token from request
	token, err := request.ParseFromRequestWithClaims(r, request.OAuth2Extractor, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return h.verifyKey, nil
	})
	if err != nil {
		switch err.(type) {
		case *jwt.ValidationError: // JWT validation error
			vErr := err.(*jwt.ValidationError)

			switch vErr.Errors {
			case jwt.ValidationErrorExpired: //JWT expired
				errHandler(w, "Access Token is expired, get a new Token", 401)
				return
			default:
				errHandler(w, "Error while parsing the Access Token!", 401)
				return
			}
		default:
			errHandler(w, "Error while parsing the Access Token!", 401)
			return
		}
	}
	if token.Valid {
		// Set user name to HTTP context
		ctx := context.WithValue(r.Context(), "userid", token.Claims.(*AppClaims).Subject)
		ctx = context.WithValue(ctx, "role", token.Claims.(*AppClaims).Role)
		next(w, r.WithContext(ctx))
	} else {
		errHandler(w, "Invalid Access Token", 401)
		return
	}
}

// TokenValidWithToken returns if the given tokenstring is valid and the decoded tokenobject
func (h *Helper) TokenValidWithToken(tokenString string) (bool, *jwt.Token) {
	token, err := jwt.ParseWithClaims(tokenString, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return h.verifyKey, nil
	})
	if err != nil {
		return false, nil
	}
	return token.Valid, token
}

// TokenValid returns if the given tokenstring is valid or not
func (h *Helper) TokenValid(tokenString string) bool {
	valid, _ := h.TokenValidWithToken(tokenString)
	return valid
}
