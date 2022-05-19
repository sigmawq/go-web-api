package main

import (
	"github.com/golang-jwt/jwt"
	"fmt"
	"os"
)

// var signingKey = []byte(os.Getenv("JWT_SECRET_KEY")) // <- Better option 
var signingKey = []byte("very-very-secret-key")

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func isAuthorized(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signing method")
		}

		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})

	if err != nil {
		fmt.Errorf("An error occured %v", err)
		return false
	}

	if !token.Valid {
		return false
	}

	return true
}
