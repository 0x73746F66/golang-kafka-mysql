package cli

import (
	"strings"

	"github.com/namsral/flag"
)

type Flags struct {
	Brokers           []string
	Topic             string
	ClientId          string
	InsertBatchSize   int
	FlushIntervalSecs int
	MysqlHost         string
	MysqlPort         int
	MysqlUser         string
	MysqlPassword     string
	MysqlSchema       string
}

func Arguments() Flags {
	var brokerUrls string
	var flags Flags
	flag.StringVar(&brokerUrls, "brokers", "kafka:9092", "Kafka Broker Urls, comma separated")
	flag.StringVar(&flags.Topic, "topic", "fiskil-logs", "Kafka topic")
	flag.StringVar(&flags.ClientId, "client-id", "mysql-ingest", "client Id")
	flag.IntVar(&flags.InsertBatchSize, "insert-batch-size", 5000, "how many INSERT statements to batch")
	flag.IntVar(&flags.FlushIntervalSecs, "flush-interval-seconds", 60, "flush INSERT statements every n seconds")
	flag.StringVar(&flags.MysqlHost, "mysql-host", "mysql", "mysql main (write only) hostname")
	flag.IntVar(&flags.MysqlPort, "mysql-port", 3306, "mysql main (write only) port")
	flag.StringVar(&flags.MysqlUser, "mysql-user", "root", "mysql main (write only) user name")
	flag.StringVar(&flags.MysqlPassword, "mysql-password", "nil", "mysql main (write only) password")
	flag.StringVar(&flags.MysqlSchema, "mysql-schema", "fiskil", "mysql main (write only) schema")
	flag.Parse()
	flags.Brokers = strings.Split(brokerUrls, ",")
	return flags
}
