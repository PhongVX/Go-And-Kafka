package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Job struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Company     string `json:"company"`
	Salary      string `json:"salary"`
}

type MongoStore struct {
	session *mgo.Session
}

var mongoStore = MongoStore{}

func main() {

	//Create MongoDB session
	session := initialiseMongo()
	mongoStore.session = session

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/jobs", jobsGetHandler).Methods("GET")
	router.HandleFunc("/jobs", jobsPostHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":9090", router))

}

func jobsGetHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	col := mongoStore.session.DB(database).C(collection)

	results := []Job{}
	col.Find(bson.M{"title": bson.RegEx{"", ""}}).All(&results)
	jsonString, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(jsonString))

}

func jobsPostHandler(w http.ResponseWriter, r *http.Request) {

	col := mongoStore.session.DB(database).C(collection)

	//Retrieve body from http request
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	//Save data into Job struct
	var _job Job
	err = json.Unmarshal(b, &_job)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//Insert job into MongoDB
	// err = col.Insert(_job)
	// if err != nil {
	// 	panic(err)
	// }

	saveJobToKafka(_job)

	//Convert job struct into json
	jsonString, err := json.Marshal(_job)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//Set content-type http header
	w.Header().Set("content-type", "application/json")

	//Send back data as response
	w.Write(jsonString)

}

func saveJobToKafka(job Job) {

	fmt.Println("save to kafka")

	jsonString, err := json.Marshal(job)

	jobString := string(jsonString)
	fmt.Print(jobString)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})
	if err != nil {
		panic(err)
	}

	// Produce messages to topic (asynchronously)
	topic := "jobs-topic1"
	for _, word := range []string{string(jobString)} {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
	}
}
