package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fzzy/radix/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	//"github.com/sendgrid/sendgrid-go"
	"log"
	"net/http"
)

//todo
// - make a alerting que
// - add twillio support

//var RedisClient redis.Client
var BoltClient *bolt.DB

type alertSet struct {
	UUID  string
	score int
}

type message struct {
	UUID           string
	score          int
	recipientEmail string
	recipientName  string
	subject        string
	text           string
}

var RedisChannel = make(chan message)
var SendgridChannel = make(chan message)

func main() {
	initializeClients()

	go timeChecker()
	go redisClient()
	go sendgridClient()

	router := httprouter.New()
	router.POST("/message", messageHandler)
	router.POST("/alert/:time", alertHandler)
	router.DELETE("/alert/:time", alertDeleteHandler)

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

//******************************Handlers*********************************************************

func messageHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//unmarshall the JSON into message

	m := message{
		UUID:           "",
		score:          0,
		recipientEmail: "charlie@test.com",
		recipientName:  "charlie",
		subject:        "test email",
		text:           "this is a test",
	}

	SendgridChannel <- m
}

func alertHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	uID := uuid.NewV4().String()
	fmt.Println(uID)

	m := message{UUID: uID, score: 1}

	//add the UUID and score to the redis sorted set
	RedisChannel <- m

	//boltDB(key[UUID] value[r.body])
	// - maybe write to bolt to store the set incase of redis meltdown or just get redis to backup
	//w.Write( UUID here )
	//return 200
}

func alertDeleteHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//redis(Z delete UUID)
	//boltDB(delete key[UUID])
	//return 200
}

func timeChecker() {
	//for;;
	//	redis(get top of set alerts score)
	//	wait for score <= current time in millis
	//		boltDB get UUID
	//		parse message
	//		email(message)

	fmt.Println("time checker is a-go-go")
}

//initializeClients starts the clients for Redis, BoltDB and SendGrid
func initializeClients() {
	//initialize BoltDB client
	//this will create messages.db file in the program directory if it
	//doesn't already exist
	BoltClient, err := bolt.Open("messages.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer BoltClient.Close()

}

func redisClient() {
	//initialize Redis client
	redisClient, err := redis.Dial("tcp", "pub-redis-10214.us-east-1-3.2.ec2.garantiadata.com:10214")
	errLog(err)
	defer redisClient.Close()

	foo, err := redisClient.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")

	_, err = redisClient.Cmd("FLUSHALL").Str() //test code
	errLog(err)                                //test code

	for {
		m := <-RedisChannel
		result := redisClient.Cmd("ZADD", "alerts", m.score, m.UUID)
		errLog(result.Err)
		fmt.Println("job Done")
	}
}

func sendgridClient() {

	//initialize Sendgrid client
	//sendgridClient := sendgrid.NewSendGridClient("*********", "********")

	for {
		m := <-SendgridChannel
		fmt.Println(m.recipientEmail, m.recipientName, m.subject, m.text)
		//	message := sendgrid.NewMail()
		//	message.AddTo(recipientEmail)
		//	message.AddToName(recipientName)
		//	message.SetSubject(subject)
		//	message.SetText(text)
		//	message.SetFrom("charlie@sendgridtesting.com")
		//	if r := sendgridClient.Send(message); r == nil {
		//		fmt.Println("Email sent!")
		//	} else {
		//		fmt.Println(r)
		//	}
	}
}

func errFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func errLog(err error) {
	if err != nil {
		log.Print(err)
	}
}
