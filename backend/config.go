package ganalytics

import "os"

type Configuration struct {
	EchoIPHost         string
	ClickHouseHost     string
	ClickHouseDB       string
	ClickHouseUser     string
	ClickHousePassword string

	GoTrackerHost string
}

var config Configuration

func LoadConfig() {
	config = Configuration{
		EchoIPHost: os.Getenv("ECHOIP_HOST"),
		ClickHouseHost: os.Getenv("CLICKHOUSE_HOST"),
		ClickHouseDB: os.Getenv("CLICKHOUSE_DB"),
		ClickHouseUser: os.Getenv("CLICKHOUSE_USER"),
		ClickHousePassword: os.Getenv("CLICKHOUSE_PASSWORD"),
		GoTrackerHost: os.Getenv("GOTRACKER_HOST"),
	}
}

func GetConfig() Configuration {
	return config;
}
