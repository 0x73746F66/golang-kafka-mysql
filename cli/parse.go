package cli

import (
	"strings"

	"github.com/namsral/flag"
)

type Flags struct {
	brokers           []string
	topic             string
	clientId          string
	insertBatchSize   int
	flushIntervalSecs int
	mysqlHost         string
	mysqlPort         int
	mysqlUser         string
	mysqlPassword     string
	mysqlSchema       string
}

func Arguments() Flags {
	var brokerUrls string
	var flags Flags
	flag.StringVar(&brokerUrls, "brokers", "kafka:9092", "Kafka Broker Urls, comma separated")
	flag.StringVar(&flags.topic, "topic", "fiskil-logs", "Kafka topic")
	flag.StringVar(&flags.clientId, "client-id", "mysql-ingest", "client Id")
	flag.IntVar(&flags.insertBatchSize, "insert-batch-size", 5000, "how many INSERT statements to batch")
	flag.IntVar(&flags.flushIntervalSecs, "flush-interval-seconds", 60, "flush INSERT statements every n seconds")
	flag.StringVar(&flags.mysqlHost, "mysql-host", "mysql", "mysql main (write only) hostname")
	flag.IntVar(&flags.mysqlPort, "mysql-port", 3306, "mysql main (write only) port")
	flag.StringVar(&flags.mysqlUser, "mysql-user", "root", "mysql main (write only) user name")
	flag.StringVar(&flags.mysqlPassword, "mysql-password", "nil", "mysql main (write only) password")
	flag.StringVar(&flags.mysqlSchema, "mysql-schema", "fiskil", "mysql main (write only) schema")
	flag.Parse()
	flags.brokers = strings.Split(brokerUrls, ",")
	return flags
}
