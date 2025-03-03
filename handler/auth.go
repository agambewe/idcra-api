package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	gcontext "github.com/kerti/idcra-api/context"
	"github.com/kerti/idcra-api/model"
	"github.com/kerti/idcra-api/service"
	"golang.org/x/net/context"
)

func Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			isAuthorized = false
			userId       string
		)
		ctx := r.Context()
		token, err := validateBearerAuthHeader(ctx, r)
		if err == nil {
			isAuthorized = true
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userIdByte, _ := base64.StdEncoding.DecodeString(claims["id"].(string))
				userId = string(userIdByte[:])
			} else {
				log.Println(err)
			}
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Println(w, "Requester ip: %q is not IP:port", r.RemoteAddr)
		}

		ctx = context.WithValue(ctx, "user_id", &userId)
		ctx = context.WithValue(ctx, "requester_ip", &ip)
		ctx = context.WithValue(ctx, "is_authorized", isAuthorized)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		loginResponse := &model.LoginResponse{}

		if r.Method == http.MethodOptions {
			response := &model.Response{
				Code: http.StatusOK,
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}

		if r.Method != http.MethodPost {
			response := &model.Response{
				Code:  http.StatusMethodNotAllowed,
				Error: gcontext.PostMethodSupported,
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}
		userCredentials, err := validateBasicAuthHeader(r)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: err.Error(),
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}
		user, err := ctx.Value("userService").(*service.UserService).ComparePassword(userCredentials)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusUnauthorized,
				Error: err.Error(),
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}

		role, err := ctx.Value("userService").(*service.UserService).FindUserRole(&user.ID)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}

		tokenString, err := ctx.Value("authService").(*service.AuthService).SignJWT(user)
		if err != nil {
			response := &model.Response{
				Code:  http.StatusBadRequest,
				Error: gcontext.TokenError,
			}
			loginResponse.Response = response
			writeResponse(w, loginResponse, loginResponse.Code)
			return
		}

		response := &model.Response{
			Code: http.StatusOK,
		}
		loginResponse.Response = response
		loginResponse.AccessToken = *tokenString
		loginResponse.Role = *role

		writeResponse(w, loginResponse, loginResponse.Code)
	})
}

func writeResponse(w http.ResponseWriter, response interface{}, code int) {
	jsonResponse, _ := json.Marshal(response)
	w.WriteHeader(code)
	if _, err := w.Write(jsonResponse); err != nil {
		log.Println(err)
	}
}

func validateBasicAuthHeader(r *http.Request) (*model.UserCredentials, error) {
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		return nil, errors.New(gcontext.CredentialsError)
	}
	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return nil, errors.New(gcontext.CredentialsError)
	}
	userCredentials := model.UserCredentials{
		Email:    pair[0],
		Password: pair[1],
	}
	return &userCredentials, nil
}

func validateBearerAuthHeader(ctx context.Context, r *http.Request) (*jwt.Token, error) {
	var tokenString string
	keys, ok := r.URL.Query()["at"]
	if !ok || len(keys) < 1 {
		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 || auth[0] != "Bearer" {
			return nil, errors.New(gcontext.CredentialsError)
		}
		tokenString = auth[1]
	} else {
		tokenString = keys[0]
	}
	token, err := ctx.Value("authService").(*service.AuthService).ValidateJWT(&tokenString)
	return token, err
}
