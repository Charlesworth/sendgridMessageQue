package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fzzy/radix/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/sendgrid/sendgrid-go"
	"log"
	"net/http"
)

//todo
// - make a alerting que
// - add twillio support

var SendgridClient *sendgrid.SGClient
var RedisClient redis.Client
var BoltClient *bolt.DB

func messageHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	email("cochrane26@gmail.com", "charlie", "test email", "did it work?")
}

func alertHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//give request a UUID
	//redis("ZADD", "alerts", params.ByName("time"), UUID)
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

func main() {
	initializeClients()

	go timeChecker()

	router := httprouter.New()
	router.POST("/message", messageHandler)
	router.POST("/alert/:time", alertHandler)
	router.DELETE("/alert/:time", alertDeleteHandler)

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func email(recipientEmail string, recipientName string, subject string, text string) {

	message := sendgrid.NewMail()
	message.AddTo(recipientEmail)
	message.AddToName(recipientName)
	message.SetSubject(subject)
	message.SetText(text)
	message.SetFrom("charlie@sendgridtesting.com")
	if r := SendgridClient.Send(message); r == nil {
		fmt.Println("Email sent!")
	} else {
		fmt.Println(r)
	}
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
	//initialize Redis client
	RedisClient, err := redis.Dial("tcp", "178.62.74.225:6379")
	errLog(err)
	defer RedisClient.Close()

	foo, err := RedisClient.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")

	_, err = RedisClient.Cmd("FLUSHALL").Str() //test code
	errLog(err)                                //test code

	//initialize BoltDB client
	//this will create messages.db file in the program directory if it
	//doesn't already exist
	BoltClient, err = bolt.Open("messages.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer BoltClient.Close()

	//initialize Sendgrid client
	SendgridClient = sendgrid.NewSendGridClient("charlesworth", "c30120509")
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
