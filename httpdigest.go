package main

import (
	"log"
	"io/ioutil"
)

var (
	username = "guest"
	password = "passw0rd"
)

func main() {
	u := "http://localhost/private/"
	d := &digestHeaders{}
	var result bool
	var err error
	result, err, d = d.Auth(username, password, u)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(result)
	log.Println(d)
	resp, err := d.Get(u + "hoge.txt")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(body))

}
