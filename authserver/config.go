package main

var ORIGINS []string = []string{}

var METHODS []string = []string{
	"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
}

var MediaTypes = map[string]string{
	"JSON": "application/json",
	"text": "text/html",
	"form": "multipart/form-data",
}
