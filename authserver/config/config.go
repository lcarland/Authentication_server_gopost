package config

import "os"

// GLOBAL Config settings

var ORIGINS = []string{
	"http://localhost:3000",
	"http://www.arricanyo.net",
	"https://www.arricanyo.net",
	"http://arricanyo.net",
	"https://arricanyo.net",
}
var METHODS = []string{
	"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
}

var SECRET string = os.Getenv("SECRET_KEY")
