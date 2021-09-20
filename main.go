package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	gofakeit "github.com/brianvoe/gofakeit/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/namsral/flag"
	kafka "github.com/segmentio/kafka-go"
)

type Message struct {
	ServiceName string    `json:"service_name"`
	Payload     string    `json:"payload"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}
type MysqlQueue struct {
	mysql  sql.DB
	values []Message
}

var (
	// cli arguments or env vars
	brokerUrls        string
	topic             string
	clientId          string
	insertBatchSize   int
	flushIntervalSecs int
	mysqlHost         string
	mysqlPort         int
	mysqlUser         string
	mysqlPassword     string
	mysqlSchema       string
)

func gen_data(brokers []string) {
	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: clientId,
	}

	config := kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	w := kafka.NewWriter(config)
	services := []string{"api", "web", "cache", "authz", "authn", "idp", "dashboard", "backend"}
	codes := []int{200, 301, 403, 404, 500}
	in := []string{"debug", "info", "warn", "error", "fatal"}
	ctx := context.Background()
	i := 0
	for {
		pickService := services[rand.Intn(len(services))]
		pickSeverity := in[rand.Intn(len(in))]
		pickCode := codes[rand.Intn(len(codes))]
		payload := fmt.Sprintf("%s %d /%s/%s", gofakeit.HTTPMethod(), pickCode, gofakeit.BuzzWord(), gofakeit.BuzzWord())
		message := Message{
			ServiceName: pickService,
			Payload:     payload,
			Severity:    pickSeverity,
			Timestamp:   time.Now(),
		}
		b, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}
		log.Println(string(b))
		err = w.WriteMessages(ctx, kafka.Message{
			Key:   []byte(strconv.Itoa(i)),
			Value: b,
		})
		if err != nil {
			panic(err.Error())
		}
		i++
	}
}

func process(brokers []string) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlSchema))
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	queue := make([]Message, 0, insertBatchSize)
	config := kafka.ReaderConfig{
		Brokers:         brokers,
		GroupID:         clientId,
		Topic:           topic,
		MinBytes:        10e3,            // 10KB
		MaxBytes:        10e6,            // 10MB
		MaxWait:         1 * time.Second, // Maximum amount of time to wait for new data to come when fetching batches of messages from kafka.
		ReadLagInterval: -1,
	}
	r := kafka.NewReader(config)
	defer r.Close()
	timer := time.Now()
	for {
		msg, err := r.ReadMessage(context.Background())
		if err != nil {
			panic(err.Error())
		}
		var message Message
		err = json.Unmarshal([]byte(msg.Value), &message)
		if err != nil {
			panic(err)
		}
		queue = append(queue, message)
		var currentTime = time.Now()
		var duration = currentTime.Sub(timer)
		if int(duration.Seconds()) >= flushIntervalSecs {
			timer = time.Now()
			con := MysqlQueue{
				mysql:  *db,
				values: queue,
			}
			_, dbErr := con.Persist()
			if dbErr != nil {
				panic(dbErr.Error())
			}
		}
		if len(queue) == insertBatchSize {
			con := MysqlQueue{
				mysql:  *db,
				values: queue,
			}
			_, dbErr := con.Persist()
			if dbErr != nil {
				panic(dbErr.Error())
			}
		}
	}
}

func (con MysqlQueue) Persist() (sql.Result, error) {
	valueStrings := make([]string, 0, len(con.values))
	for _, message := range con.values {
		valueStrings = append(valueStrings, fmt.Sprintf("('%s', '%s', '%s', '%s')", message.ServiceName, message.Payload, message.Severity, message.Timestamp.Format("2006-01-02 15:04:05")))
	}
	stmt := fmt.Sprintf("INSERT INTO `service_logs` (`service_name`, `payload`, `severity`, `timestamp`) VALUES %s", strings.Join(valueStrings, ","))
	con.values = con.values[:0]
	return con.mysql.Exec(stmt)
}

func main() {
	flag.StringVar(&brokerUrls, "brokers", "kafka:9092", "Kafka Broker Urls, comma separated")
	flag.StringVar(&topic, "topic", "fiskil-logs", "Kafka topic")
	flag.StringVar(&clientId, "client-id", "mysql-ingest", "client Id")
	flag.IntVar(&insertBatchSize, "insert-batch-size", 5000, "how many INSERT statements to batch")
	flag.IntVar(&flushIntervalSecs, "flush-interval-seconds", 60, "flush INSERT statements every n seconds")
	flag.StringVar(&mysqlHost, "mysql-host", "mysql", "mysql main (write only) hostname")
	flag.IntVar(&mysqlPort, "mysql-port", 3306, "mysql main (write only) port")
	flag.StringVar(&mysqlUser, "mysql-user", "root", "mysql main (write only) user name")
	flag.StringVar(&mysqlPassword, "mysql-password", "nil", "mysql main (write only) password")
	flag.StringVar(&mysqlSchema, "mysql-schema", "fiskil", "mysql main (write only) schema")
	flag.Parse()
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	brokers := strings.Split(brokerUrls, ",")
	for i := 0; i < 10; i++ {
		go gen_data(brokers)
	}
	process(brokers)
}
