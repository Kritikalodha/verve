package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"bytes"
	"strconv"
	"github.com/go-redis/redis/v8"
	"context"
	//"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	uniqueRequests sync.Map
	countMutex     sync.Mutex
	uniqueCount    int
	ctx            = context.Background()
	redisClient    *redis.Client
	logger         *log.Logger
)

func main() {
	// setup Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr, // Use REDIS_ADDR environment variable
	})

	// Can be added to use kafka
	// kafkaBroker := os.Getenv("KAFKA_BROKER")
	// kafkaProducer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBroker})
	// if err != nil {
	// 	log.Fatalf("Failed to create Kafka producer: %v", err)
	// }
	// defer kafkaProducer.Close()


	// setup log file
	logFile, err := os.OpenFile("requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// create a logger that writes to the log file
	logger = log.New(logFile, "APP_LOG: ", log.LstdFlags)

	// set up HTTP server
	http.HandleFunc("/api/verve/accept", acceptHandler)
	go logUniqueRequestCount()

	logger.Println("Server started at :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}

func acceptHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	endpoint := r.URL.Query().Get("endpoint")

	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		logger.Println("Request failed - Missing id parameter")
		return
	}

	// check uniqueness in Redis for Load-Balancer Deduplication (Extension 2)
	unique := checkAndStoreID(id)

	if unique {
		countMutex.Lock()
		uniqueCount++
		countMutex.Unlock()
	}

	if endpoint != "" {
		go sendHTTPPost(endpoint) // Fire POST request (Extension 1)
	}

	_, err := w.Write([]byte("ok"))
	if err != nil {
		logger.Println("Failed to write response:", err)
	} else {
		logger.Printf("Request processed for ID: %s", id)
	}
}


func checkAndStoreID(id string) bool {
	// store in Redis with expiration to ensure deduplication across instances
	isNew, err := redisClient.SetNX(ctx, id, true, time.Minute).Result()
	if err != nil {
		logger.Println("Failed to connect to Redis:", err)
		return false
	}
	// if isNew is true, the ID was unique and successfully stored
	if isNew {
		logger.Printf("ID %s is unique", id)
		return true
	} else {
		logger.Printf("ID %s is a duplicate", id)
		return false
	}
}

func logUniqueRequestCount() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		countMutex.Lock()
		logger.Printf("Unique requests in the last minute: %d", uniqueCount)
		countMutex.Unlock()

		// send the unique count to Kafka
		//sendUniqueCountToKafka(count)
	}
}

func sendHTTPPost(endpoint string) {
	count := strconv.Itoa(uniqueCount)
	body := []byte(fmt.Sprintf(`{"unique_requests": %s}`, count))

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logger.Println("Failed to send HTTP POST request:", err)
		return
	}
	defer resp.Body.Close()

	logger.Printf("POST to %s returned status: %d", endpoint, resp.StatusCode)
}

// Can be added to enable kafka
// func sendUniqueCountToKafka(count int) {
// 	message := fmt.Sprintf(`{"unique_requests": %d}`, count)

// 	// Produce the message to Kafka
// 	err := kafkaProducer.Produce(&kafka.Message{
// 		Topic: "unique_requests_topic", // Replace with your topic name
// 		Value: []byte(message),
// 	}, nil)

// 	if err != nil {
// 		logger.Println("Failed to send message to Kafka:", err)
// 	} else {
// 		logger.Printf("Sent unique count %d to Kafka", count)
// 	}
// }
