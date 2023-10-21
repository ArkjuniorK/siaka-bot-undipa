package main

import (
	"net/http"
)

func GetSTB(cs []*http.Cookie) string {
	var INDEX = 2
	return cs[INDEX].Value
}
