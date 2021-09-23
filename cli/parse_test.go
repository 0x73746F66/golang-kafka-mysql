package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/namsral/flag"
	"github.com/stretchr/testify/assert"
)

func TestArgumentsDefaults(t *testing.T) {
	defaultBrokers := []string{"kafka:9092"}
	defaultTopic := "fiskil-logs"
	defaultClientId := "mysql-ingest"
	defaultInsertBatchSize := 5000
	defaultflushInterval := 60
	defaultMysqlHost := "mysql"
	defaultMysqlPort := 3306
	defaultMysqlUser := "root"
	defaultMysqlPassword := "nil"
	defaultMysqlSchema := "fiskil"
	os.Args = []string{"main.go"}
	args := Arguments()
	assert.Equal(t, args.brokers, defaultBrokers)
	assert.Equal(t, args.topic, defaultTopic)
	assert.Equal(t, args.clientId, defaultClientId)
	assert.Equal(t, args.insertBatchSize, defaultInsertBatchSize)
	assert.Equal(t, args.flushIntervalSecs, defaultflushInterval)
	assert.Equal(t, args.mysqlHost, defaultMysqlHost)
	assert.Equal(t, args.mysqlPort, defaultMysqlPort)
	assert.Equal(t, args.mysqlUser, defaultMysqlUser)
	assert.Equal(t, args.mysqlPassword, defaultMysqlPassword)
	assert.Equal(t, args.mysqlSchema, defaultMysqlSchema)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsTopic(t *testing.T) {
	want := "test"
	os.Args = []string{"main.go", "-topic", want}
	args := Arguments()
	assert.Equal(t, args.topic, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsBrokers(t *testing.T) {
	want := []string{"kafka:9093"}
	os.Args = []string{"main.go", "-brokers", "kafka:9093"}
	args := Arguments()
	assert.Equal(t, args.brokers, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsClientId(t *testing.T) {
	want := "foobar"
	os.Args = []string{"main.go", "-client-id", want}
	args := Arguments()
	assert.Equal(t, args.clientId, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsInsertBatchSize(t *testing.T) {
	want := 1000
	os.Args = []string{"main.go", "-insert-batch-size", fmt.Sprintf("%d", want)}
	args := Arguments()
	assert.Equal(t, args.insertBatchSize, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsFlushIntervalSecs(t *testing.T) {
	want := 30
	os.Args = []string{"main.go", "-flush-interval-seconds", fmt.Sprintf("%d", want)}
	args := Arguments()
	assert.Equal(t, args.flushIntervalSecs, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsMysqlHost(t *testing.T) {
	want := "foobar"
	os.Args = []string{"main.go", "-mysql-host", want}
	args := Arguments()
	assert.Equal(t, args.mysqlHost, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsMysqlPort(t *testing.T) {
	want := 33306
	os.Args = []string{"main.go", "-mysql-port", fmt.Sprintf("%d", want)}
	args := Arguments()
	assert.Equal(t, args.mysqlPort, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsMysqlUser(t *testing.T) {
	want := "foobar"
	os.Args = []string{"main.go", "-mysql-user", want}
	args := Arguments()
	assert.Equal(t, args.mysqlUser, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsMysqlPassword(t *testing.T) {
	want := "foobar"
	os.Args = []string{"main.go", "-mysql-password", want}
	args := Arguments()
	assert.Equal(t, args.mysqlPassword, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestArgumentsMysqlSchema(t *testing.T) {
	want := "foobar"
	os.Args = []string{"main.go", "-mysql-schema", want}
	args := Arguments()
	assert.Equal(t, args.mysqlSchema, want)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}
