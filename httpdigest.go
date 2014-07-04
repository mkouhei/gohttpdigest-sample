package main

import (
	"log"
)

var (
	username = "guest"
	password = "passw0rd"
)

func main() {
	u := "http://localhost/private/"
	d := &digestHeaders{}
	result, err := d.Auth(username, password, u)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)

}
