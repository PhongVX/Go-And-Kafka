package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Shopify/sarama"
	"gopkg.in/mgo.v2"
)

var (
	MONGODB_HOST_PORT = os.Getenv("MONGODB_HOST_PORT") //"172.88.88.2:27017"
	DATABASE          = os.Getenv("DATABASE")          //"db"
	DB_USER_NAME      = os.Getenv("DB_USER_NAME")      //""
	DB_PASSWORD       = os.Getenv("DB_PASSWORD")       //""
	COLLECTION        = os.Getenv("COLLECTION")        //"jobs"
	KAFKA_HOST_PORT   = os.Getenv("KAFKA_HOST_PORT")   //"172.88.88.5:9092"
	KAFKA_TOPIC       = os.Getenv("KAFKA_TOPIC")       //"demo"
)

type MongoStore struct {
	session *mgo.Session
}

type Job struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Company     string `json:"company"`
	Salary      string `json:"salary"`
}

var mongoStore = MongoStore{}

func main() {

	//Create MongoDB session
	session := initialiseMongo()
	mongoStore.session = session

	receiveFromKafka()

}

func initialiseMongo() (session *mgo.Session) {
	info := &mgo.DialInfo{
		Addrs:    []string{MONGODB_HOST_PORT},
		Timeout:  60 * time.Second,
		Database: DATABASE,
		Username: DB_USER_NAME,
		Password: DB_PASSWORD,
	}

	session, err := mgo.DialWithInfo(info)
	if err != nil {
		panic(err)
	}

	return
}

func receiveFromKafka() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	brokers := []string{KAFKA_HOST_PORT}
	// Create new consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()
	topic := KAFKA_TOPIC
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				fmt.Println("Received messages", string(msg.Value))
				saveJobToMongo(string(msg.Value))
			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh

}

func saveJobToMongo(jobString string) {

	fmt.Println("Save to MongoDB")
	col := mongoStore.session.DB(DATABASE).C(COLLECTION)
	fmt.Println("1")
	//Save data into Job struct
	var _job Job
	b := []byte(jobString)
	err := json.Unmarshal(b, &_job)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	//Insert job into MongoDB
	errMongo := col.Insert(_job)
	if errMongo != nil {
		fmt.Printf("%v", errMongo)
		return
		//panic(errMongo)
	}

	fmt.Printf("Saved to MongoDB : %s", jobString)

}
