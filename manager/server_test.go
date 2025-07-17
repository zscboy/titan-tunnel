package main

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestToken(t *testing.T) {
	claims := jwt.MapClaims{
		"user": "abc",
		"exp":  time.Now().Add(time.Second * time.Duration(86400)).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("ef61299a-53e6-11f0-87d4-e72f7b1c2247"))
	if err != nil {
		t.Logf(err.Error())
		return
	}
	t.Log("token", tokenString)

}
