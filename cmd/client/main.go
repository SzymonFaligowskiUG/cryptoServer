package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Argon2id
var action = flag.String("a", "", "chose action: login, register")
var email = flag.String("e", "", "eneter your email")
var password = flag.String("p", "", "eneter your password")

func main() {
	flag.Parse()
	if *action == "" || *email == "" || *password == "" {
		log.Fatal("you have to specify all parameters")
	}
	if *action == "login" {
		login(*email, *password)
	}
	if *action == "register" {
		register(*email, *password)
	}

}

func login(email, password string) {
	h := sha256.New()
	h.Write([]byte(password))

	b, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": hex.EncodeToString(h.Sum(nil)),
	})
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/login", bytes.NewBuffer(b))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	rawBody, _ := io.ReadAll(res.Body)
	fmt.Printf("%s", rawBody)
}

func register(email, password string) {
	h := sha256.New()
	h.Write([]byte(password))

	b, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": hex.EncodeToString(h.Sum(nil)),
	})
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/register", bytes.NewBuffer(b))
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	rawBody, _ := io.ReadAll(res.Body)
	fmt.Printf("%s", rawBody)
}
