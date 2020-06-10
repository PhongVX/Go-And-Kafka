package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
)

var (
	KAFKA_CON   = os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")
	KAFKA_TOPIC = os.Getenv("KAFKA_TOPIC")
	PORT        = os.Getenv("PORT")
)

type Job struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Company     string `json:"company"`
	Salary      string `json:"salary"`
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/jobs", jobsPostHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(PORT, router))
}

func jobsPostHandler(w http.ResponseWriter, r *http.Request) {
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

	producer, err := initProducer()
	if err != nil {
		fmt.Println("Error producer: ", err.Error())
		panic(err)
	}

	fmt.Println("save to kafka")

	jsonString, err := json.Marshal(job)

	jobString := string(jsonString)
	fmt.Print(jobString)

	publishMessage(jobString, producer)
}

func initProducer() (sarama.SyncProducer, error) {
	// setup sarama log to stdout
	sarama.Logger = log.New(os.Stdout, "", log.Ltime)

	// producer config
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	// async producer
	//prd, err := sarama.NewAsyncProducer([]string{kafkaConn}, config)

	// sync producer
	prd, err := sarama.NewSyncProducer([]string{KAFKA_CON}, config)

	return prd, err
}

func publishMessage(message string, producer sarama.SyncProducer) {
	// publish sync
	msg := &sarama.ProducerMessage{
		Topic: KAFKA_TOPIC,
		Value: sarama.StringEncoder(message),
	}
	p, o, err := producer.SendMessage(msg)
	if err != nil {
		fmt.Println("Error publish: ", err.Error())
	}
	fmt.Println("Partition: ", p)
	fmt.Println("Offset: ", o)
}
