package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/argon2"
)

func main() {
	email := "makapaka@example.com"
	passwor := "cryptIsAwsome123#"
	// Register
	regTime := time.Now()
	register(email, passwor)
	fmt.Printf("register took:%s\n", time.Since(regTime))

	//Good login
	goodLogTime := time.Now()
	login(email, passwor)
	fmt.Printf("good login took:%s\n", time.Since(goodLogTime))

	//bad login
	badLogTime := time.Now()
	login(email, "heheWrongPassword")
	fmt.Printf("bas login took:%s\n", time.Since(badLogTime))

	//server metodology stest
	serTime := time.Now()
	salt := make([]byte, 10)
	rand.Read(salt)

	hashServer := argon2.IDKey([]byte(passwor), salt, 1, 64*1024, 4, 32)
	h := sha256.New()
	h.Write([]byte(passwor))
	hashUser := argon2.IDKey([]byte(hex.EncodeToString(h.Sum(nil))), salt, 1, 64*1024, 4, 32)
	if hex.EncodeToString(hashServer) != hex.EncodeToString(hashUser) {
		fmt.Println("server cant decode user password")
	}

	fmt.Printf("server took:%s\n", time.Since(serTime))
}
func login(email, password string) string {
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
	return string(rawBody)
}

func register(email, password string) string {
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
	return string(rawBody)
}
