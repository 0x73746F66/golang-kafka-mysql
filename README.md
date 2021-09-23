# Awesome code test for Fiskil

# Windows Prerequisites

See: [3musketeers.io](https://3musketeers.io/docs/#prerequisites)

# Mac Prerequisites

See: [3musketeers.io](https://3musketeers.io/docs/#prerequisites)

# Linux Prerequisites

See: [3musketeers.io](https://3musketeers.io/docs/#prerequisites)

# Getting Started

1. Everything happens inside Docker. Complete the [3musketeers.io hello world](https://3musketeers.io/docs/#hello-world) and you're all set.

2. Checkout the make targets `make help`:

```
help                           This help.
setup                          Creates docker networks and volumes
mysql-init                     applies mysql schema and initial test data
update                         Pulls the latest go mysql zookeeper and kafka images
up                             Starts the publisher, kafka, mysql, and zookeeper containers
run                            Build and run publisher
services                       Starts all containers
ps                             Shows this projects running docker containers
down                           Bring down containers and removes anything else orphaned
test                           runs go tests inside Docker
```

3. Logical order of commands to run this project

- prepare teh `.env` file either my writing your own or `cp .env-example .env` and edit the values
- `make update`
- `make services`
- `make mysql-init` or use a GUI to execute `.development/mysql/schema.sql`
- Create the Topic manually may be needed on Mac (I think, because auto create indexes setting worked fine on Linux): `docker exec -ti kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --zookeeper zookeeper:2181 --topic fiskil-logs --partitions 1 --replication-factor 1`
- `make run` to consume the logs from kafka and ingest into mysql
- `make test` to run some unit tests

# Rationale

## Using MyISAM engine

There is almost only writes being done, InnoDB is significantly slower than MyISAM. Specifically we can utilise the `ROW_FORMAT=FIXED` because InnoDB will ignore this.
In this setup there is no row fragmentation which means quicker operations, and if we used variable width the data file pointer is based on the byte offset which for large tables of log data there is unnecessary overhead to the table when we can simple have fixed width and pointer using the row offset.

Another optimisation is to spread log data over multiple super-fast nvme drives using `partition by key(<special_key>) PARTITIONS <n> (PARTITION p1 DATA DIRECTORY='/nvme/p1/service_logs', PARTITION p2 DATA DIRECTORY='/nvme/p2/service_logs')` etc. where `n` is how many partitions and `special_key` is a way to ensure that concurrent data can land on separate disks, e.g. current nanosecond divided by how many disks to equally spread the load so each disk controller can get a set of data to write with a consistent break but there is a constant flow.

Preferably I would just ingest log data into something like Elasticsearch which manages this well enough for you, but if we are using MySQL this setup can be like rocket fuel compared to normal MySQL deployments.

## Using a scheduled stored procedure

In the spec the column `created_at` used datatype TIMESTAMP which has precision to the second, for high precision you would use an unsigned integer representing unix epoch.

MySQL table `service_severity` is kept up-to-date to the second, the highest precision capable of being recorded in the `created_at`.
The MyISAM engine is extremely high performing for writes, so while writing to the table more than 1 time per second is possible, it is wasteful.
Furthermore this approach minimises code complexity, given there is zero code used to read the database and calculate the aggregate count column at all.

Using `INSERT .. ON DUPLICATE KEY UPDATE` is a single operation, significantly reduces IOPS requirements versus `SELECT` if exists `INSERT` else compute update values and `UPDATE` which is at least 2 separate operations. If done in code there is additional I/O latency and network jitter so the scheduled stored procedure is optimal in all regards.

## Not mocking things

I like to mock for QA, but this is an engineering role and a coding challenge. So standing up a couple of services and seeding dummy data is pretty straight forward (and fun).
I ran out of time for writing full test coverage, so the functions with side effects (mysql/kafka) don't have unit test and typically have good QA tests done before a production release in a typical workplace.

## Writing less go than expected

Why write code if there is a better way? Let the database do database things.
