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
	// we want to checkpoint this client to preserve our mysql table consistency
	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: args.ClientId,
	}
	// sane defaults for localhost and the parameters of this test
	config := kafka.WriterConfig{
		Brokers:      args.Brokers,
		Topic:        args.Topic,
		Balancer:     &kafka.LeastBytes{},
		Dialer:       dialer,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	w := kafka.NewWriter(config)
	// release the dummies!
	logs := generator.Generate(50000)
	services := []string{"api", "web", "cache", "authz", "authn", "idp", "dashboard", "backend"}
	sev := []string{"debug", "info", "warn", "error", "fatal"}
	ctx := context.Background()
	// outside we'll call this 100 times concurrently for a total of 5mil records to insert
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
	// in prod this will be a connection to main as we will be writing
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", args.MysqlUser, args.MysqlPassword, args.MysqlHost, args.MysqlPort, args.MysqlSchema))
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(10) // pooling is necessary as we rely on pointers further on. good habbit too
	db.SetMaxIdleConns(5)  // another good habit for long running threads, we don't want GC to cycle connections and force tcp too often
	queue := make([]Message, 0, args.InsertBatchSize)
	// kafka has a checkpoint implementation via group id (the messages checkpointed) for each client id (how a process self identifies)
	// personally I prefer the simpler data science checkpointing used by Spark, Kafka is overly complex for no additional function
	config := kafka.ReaderConfig{
		Brokers:         args.Brokers,
		GroupID:         args.ClientId,
		Topic:           args.Topic,
		MinBytes:        10e3,            // 10KB
		MaxBytes:        10e6,            // 10MB
		MaxWait:         1 * time.Second, // wait for new data when fetching batches
		ReadLagInterval: -1,
	}
	r := kafka.NewReader(config)
	// defer is safe here: not in a loop, block, or wrapper and no reliance on pointers
	defer r.Close() // reconsider if any of these change
	// prep FlushIntervalSecs
	timer := time.Now()
	for {
		// Huzzah!
		msg, err := r.ReadMessage(context.Background())
		if err != nil {
			panic(err.Error())
		}
		// Just a charp..
		var message Message
		err = json.Unmarshal([]byte(msg.Value), &message)
		if err != nil {
			panic(err)
		}
		// json was good
		queue = append(queue, message)
		// queue for bulk INSERT
		var currentTime = time.Now()
		// handle FlushIntervalSecs
		var duration = currentTime.Sub(timer)
		if int(duration.Seconds()) >= args.FlushIntervalSecs {
			timer = time.Now()
			con := MysqlQueue{
				mysql:  *db,   // wrapping makes the connection easier to reuse
				values: queue, // this could be passed as an argument, but this reads clearer if we are wrapping anyway
			}
			// it's 'time' to persist any rows we have gathered (pun intended)
			_, dbErr := con.Persist()
			if dbErr != nil {
				panic(dbErr.Error()) // it's just a code test so let's panic any errors as they're likely from services on localhost
			}
		}
		// a lot of messages came quickly and we can't wait for the interval
		if len(queue) == args.InsertBatchSize {
			con := MysqlQueue{ // as above, see cleaner reuse!
				mysql:  *db,
				values: queue,
			}
			_, dbErr := con.Persist()
			if dbErr != nil {
				panic(dbErr.Error()) // same as above
			}
		}
	}
}

func (con MysqlQueue) Persist() (sql.Result, error) {
	valueStrings := make([]string, 0, len(con.values))
	// prepared statements are limited to one insert only, we need to optimise this
	//@TODO fix upstream database/sql
	for _, message := range con.values {
		valueStrings = append(valueStrings, fmt.Sprintf("('%s', '%s', '%s', '%s')", message.ServiceName, message.Payload, message.Severity, message.Timestamp.Format("2006-01-02 15:04:05")))
	}
	// at the risk of SQLi - ensure inputs are trusted and values are not not an end user input
	stmt := fmt.Sprintf("INSERT INTO `service_logs` (`service_name`, `payload`, `severity`, `timestamp`) VALUES %s", strings.Join(valueStrings, ","))
	con.values = con.values[:0]
	return con.mysql.Exec(stmt)
}

func main() {
	args := cli.Arguments()
	sigchan := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	// As this is just a code test, let's play with signals and seeding dummy data!
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigchan
		// interupted ctrl+c or docker stop
		fmt.Println(sig)
		// otherwise process() runs forever and maybe 5mil mysql rows is not ideal for your laptop
		done <- true
	}()
	// seed 5mil records to kafka for demo
	for i := 0; i < 100; i++ {
		go gen_data(args)
	}
	// continue polling kafka for messages, fluching periodically and when the threshold is met
	process(args)
	<-done
}
