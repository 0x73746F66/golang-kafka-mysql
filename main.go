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

	_ "github.com/go-sql-driver/mysql"
	kafka "github.com/segmentio/kafka-go"
	"gitlab.com/chrislangton/fiskil/cli"
	"gitlab.com/chrislangton/fiskil/generator"
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

func gen_data(args cli.Flags) {
	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: args.ClientId,
	}

	config := kafka.WriterConfig{
		Brokers:      args.Brokers,
		Topic:        args.Topic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	w := kafka.NewWriter(config)
	logs := generator.Generate(50000)
	services := []string{"api", "web", "cache", "authz", "authn", "idp", "dashboard", "backend"}
	sev := []string{"debug", "info", "warn", "error", "fatal"}
	ctx := context.Background()
	i := 0
	for _, payload := range logs {
		pickService := services[rand.Intn(len(services))]
		pickSeverity := sev[rand.Intn(len(sev))]
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

func process(args cli.Flags) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", args.MysqlUser, args.MysqlPassword, args.MysqlHost, args.MysqlPort, args.MysqlSchema))
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	queue := make([]Message, 0, args.InsertBatchSize)
	config := kafka.ReaderConfig{
		Brokers:         args.Brokers,
		GroupID:         args.ClientId,
		Topic:           args.Topic,
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
		if int(duration.Seconds()) >= args.FlushIntervalSecs {
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
		if len(queue) == args.InsertBatchSize {
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
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	args := cli.Arguments()
	for i := 0; i < 100; i++ {
		go gen_data(args)
	}
	process(args)
}
