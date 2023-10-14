package ksqlCmd

import (
	"fmt"

	"github.com/thmeitz/ksqldb-go"
	"github.com/thmeitz/ksqldb-go/net"
)

func init() {

}

func UpKsql(userName string, pass string, host string) error {
	options := net.Options{
		Credentials: net.Credentials{Username: userName, Password: pass},
		BaseUrl:     host,
		AllowHTTP:   true,
	}

	kcl, err := ksqldb.NewClientWithOptions(options)
	if err != nil {
		fmt.Errorf("error connect client ", err)
		return err
	}
	defer kcl.Close()
	_, err = kcl.Execute(ksqldb.ExecOptions{KSql: `
			CREATE STREAM IF NOT EXISTS MIGRATION_EVENTS (
			  version_key STRING KEY, version STRING,
			  name STRING, state STRING, checksum STRING,
			  started_on STRING, completed_on STRING,
			  previous STRING, error_reason STRING
			) WITH (
			  KAFKA_TOPIC = 'MIGRATION_EVENTS',
			  VALUE_FORMAT = 'JSON', PARTITIONS = 1,
			  REPLICAS = 1
			);
		`})
	if err != nil {
		return fmt.Errorf("create stream migrate fail: %w", err)
	}
	fmt.Println("create stream MIGRATION_EVENTS success ")

	_, err = kcl.Execute(ksqldb.ExecOptions{KSql: `
CREATE TABLE IF NOT EXISTS MIGRATION_SCHEMA_VERSIONS WITH (
	  KAFKA_TOPIC = 'MIGRATION_SCHEMA_VERSIONS'
	) AS 
	SELECT 
	  version_key, 
	  latest_by_offset(version) AS version, 
	  latest_by_offset(name) AS name, 
	  latest_by_offset(state) AS state, 
	  latest_by_offset(checksum) AS checksum, 
	  latest_by_offset(started_on) AS started_on, 
	  latest_by_offset(completed_on) AS completed_on, 
	  latest_by_offset(previous) AS previous, 
	  latest_by_offset(error_reason) AS error_reason 
	FROM 
	  MIGRATION_EVENTS 
	GROUP BY 
	  version_key;
		`})
	if err != nil {
		return fmt.Errorf("create stream migrate fail  %w", err)
	}
	fmt.Println("create table MIGRATION_SCHEMA_VERSIONS success")

	_, err = kcl.Execute(ksqldb.ExecOptions{KSql: `
			CREATE STREAM IF NOT EXISTS KEC_MIGRATION_EVENTS (
			  version_key STRING KEY, version STRING,
			  name STRING, state STRING, checksum STRING,
			  started_on STRING, completed_on STRING,
			  previous STRING, error_reason STRING
			) WITH (
			  KAFKA_TOPIC = 'KEC_MIGRATION_EVENTS',
			  VALUE_FORMAT = 'JSON', PARTITIONS = 1,
			  REPLICAS = 1
			);
		`})
	if err != nil {
		return fmt.Errorf("create stream migrate fail: %w", err)
	}
	fmt.Println("create stream KEC_MIGRATION_EVENTS success ")

	_, err = kcl.Execute(ksqldb.ExecOptions{KSql: `
CREATE TABLE IF NOT EXISTS KEC_MIGRATION_SCHEMA_VERSIONS WITH (
	  KAFKA_TOPIC = 'KEC_MIGRATION_SCHEMA_VERSIONS'
	) AS 
	SELECT 
	  version_key, 
	  latest_by_offset(version) AS version, 
	  latest_by_offset(name) AS name, 
	  latest_by_offset(state) AS state, 
	  latest_by_offset(checksum) AS checksum, 
	  latest_by_offset(started_on) AS started_on, 
	  latest_by_offset(completed_on) AS completed_on, 
	  latest_by_offset(previous) AS previous, 
	  latest_by_offset(error_reason) AS error_reason 
	FROM 
	  KEC_MIGRATION_EVENTS 
	GROUP BY 
	  version_key;
		`})
	if err != nil {
		return fmt.Errorf("create stream migrate fail  %w", err)
	}

	fmt.Println("create table KEC_MIGRATION_SCHEMA_VERSIONS success")
	return nil
}
