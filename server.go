package crypto6temp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/argon2"
)

type Server struct {
	Sessions     map[string]string
	dbConnection *dynamodb.Client
}

type User struct {
	Email    string `dynamodbav:"email"`
	Password string `dynamodbav:"password"`
	Salt     string `dynamodbav:"salt"`
}

func (u User) GetKey() map[string]types.AttributeValue {
	email, err := attributevalue.Marshal(u.Email)
	if err != nil {
		panic(err)
	}
	return map[string]types.AttributeValue{"email": email}
}
func (s Server) loginHandle(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	rawBody, _ := io.ReadAll(r.Body)
	json.Unmarshal(rawBody, &body)

	user := User{
		Email: body["email"],
	}
	response, _ := s.dbConnection.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key: user.GetKey(), TableName: aws.String("users"),
	})

	attributevalue.UnmarshalMap(response.Item, &user)
	salt, _ := hex.DecodeString(user.Salt)
	hash := argon2.IDKey([]byte(body["password"]), salt, 1, 64*1024, 4, 32)
	if user.Password == hex.EncodeToString(hash) {
		fmt.Fprintf(w, "Welcom %s", user.Email)
		return
	}
	fmt.Fprint(w, "Wrong email or password")

}
func (s Server) registerHandle(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	rawBody, _ := io.ReadAll(r.Body)
	json.Unmarshal(rawBody, &body)
	salt := make([]byte, 10)
	rand.Read(salt)

	hash := argon2.IDKey([]byte(body["password"]), salt, 1, 64*1024, 4, 32)

	newUser := User{
		Email:    body["email"],
		Password: hex.EncodeToString(hash),
		Salt:     hex.EncodeToString(salt),
	}

	item, err := attributevalue.MarshalMap(newUser)
	if err != nil {
		panic(err)
	}

	_, err = s.dbConnection.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("users"), Item: item,
	})
	if err != nil {
		log.Printf("Couldn't add item to table. Here's why: %v\n", err)
		fmt.Fprint(w, "Someting broke")
		return
	}
	fmt.Fprint(w, "OK")

}
func (s Server) protectedHandle(w http.ResponseWriter, r *http.Request) {}

func (s Server) ServerInit() {
	r := mux.NewRouter()
	s.dbConnection = s.createLocalClient()

	r.HandleFunc("/login", s.loginHandle).Methods("POST")
	r.HandleFunc("/register", s.registerHandle).Methods("POST")
	r.HandleFunc("/protected", s.protectedHandle).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
	}

	log.Fatal(srv.ListenAndServe())
}

func (s Server) createLocalClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:8000")
	})
}

func (s Server) CreateUserTable() {
	s.dbConnection = s.createLocalClient()
	_, err := s.dbConnection.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{{
			AttributeName: aws.String("email"),
			AttributeType: types.ScalarAttributeTypeS,
		}},
		KeySchema: []types.KeySchemaElement{{
			AttributeName: aws.String("email"),
			KeyType:       types.KeyTypeHash,
		}},
		TableName: aws.String("users"),
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})
	if err != nil {
		log.Printf("Couldn't create table %v. Here's why: %v\n", "users", err)
	}
}
