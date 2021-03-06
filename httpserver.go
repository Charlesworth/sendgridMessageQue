package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fzzy/radix/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	//"github.com/sendgrid/sendgrid-go"
	"encoding/json"
	"log"
	"net/http"
	//"time"
)

//todo
// - make a alerting que
// - add twillio support

type Message struct {
	UUID           string
	Score          int
	RecipientEmail string
	RecipientName  string
	Subject        string
	Text           string
}

var RedisChannel = make(chan Message)
var SendgridChannel = make(chan Message)
var BoltReadChannel = make(chan string)
var BoltWriteChannel = make(chan Message)
var BoltDeleteChannel = make(chan string)

func main() {
	go alertManager()
	go redisClient()
	go sendgridClient()
	go boltWriteClient()
	go boltReadClient()

	router := httprouter.New()
	router.POST("/message", messageHandler)
	router.POST("/alert/:time", alertPostHandler)
	router.DELETE("/alert/:time", alertDeleteHandler)

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//******************************Handlers*********************************************************

func messageHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//unmarshall the JSON into message

	m := Message{
		UUID:           "123",
		Score:          0,
		RecipientEmail: "charlie@test.com",
		RecipientName:  "charlie",
		Subject:        "test email",
		Text:           "this is a test",
	}

	//	mjson, err := json.Marshal(m)
	//	errLog(err)

	//	var poo Message
	//	err = json.Unmarshal(mjson, &poo)
	//	fmt.Println("!")
	//	fmt.Println(poo.RecipientName)

	//SendgridChannel <- m

	BoltWriteChannel <- m
	//time.Sleep(time.Second)
	BoltReadChannel <- "123"
	//BoltDeleteChannel <- "123"
}

func alertPostHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	uID := uuid.NewV4().String()
	fmt.Println(uID)

	m := Message{UUID: uID, Score: 1}

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

//******************************Alert Manager*********************************************************

func alertManager() {
	//for;;
	//	redis(get top of set alerts score)
	//	wait for score <= current time in millis
	//		boltDB get UUID
	//		parse message
	//		email(message)

	fmt.Println("time checker is a-go-go")
}

//******************************Clients*********************************************************

func boltWriteClient() {
	boltClient, err := bolt.Open("messages.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer boltClient.Close()

	//do we need to check a bucket exists or make one
	boltClient.Update(func(tx *bolt.Tx) error {
		// Create a bucket.
		tx.CreateBucketIfNotExists([]byte("m"))
		return nil
	})

	fmt.Println("bolt writer ready")

	for {
		select {
		case m := <-BoltWriteChannel:
			mjson, err := json.Marshal(m)
			errLog(err)
			boltClient.Update(func(tx *bolt.Tx) error {
				// Set the value "bar" for the key "foo".
				err = tx.Bucket([]byte("m")).Put([]byte(m.UUID), []byte(mjson))
				errLog(err)
				return nil
			})

		case m := <-BoltDeleteChannel:
			boltClient.Update(func(tx *bolt.Tx) error {
				// Set the value "bar" for the key "foo".
				err = tx.Bucket([]byte("m")).Delete([]byte(m))
				errLog(err)
				return nil
			})
		}
	}
}

func boltReadClient() {
	boltClient, err := bolt.Open("messages.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer boltClient.Close()

	fmt.Println("bolt reader ready")

	for {
		uuID := <-BoltReadChannel

		var b []byte
		boltClient.View(func(tx *bolt.Tx) error {
			// Set the value "bar" for the key "foo".
			b = tx.Bucket([]byte("m")).Get([]byte(uuID))
			errLog(err)

			return nil
		})

		var mjson Message
		err := json.Unmarshal(b, &mjson)
		errLog(err)
		fmt.Println("*************")
		fmt.Println(mjson.Text)
		fmt.Println("*************")
	}
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
		result := redisClient.Cmd("ZADD", "alerts", m.Score, m.UUID)
		errLog(result.Err)
		fmt.Println("job Done")
	}
}

func sendgridClient() {

	//initialize Sendgrid client
	//sendgridClient := sendgrid.NewSendGridClient("*********", "********")

	for {
		m := <-SendgridChannel
		fmt.Println(m.RecipientEmail, m.RecipientName, m.Subject, m.Text)
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

//******************************Errors*********************************************************

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
