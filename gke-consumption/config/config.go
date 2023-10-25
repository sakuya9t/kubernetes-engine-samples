package config

import (
	"fmt"
	"time"

	"github.com/apsdehal/go-logger"
	"github.com/spf13/viper"
)

const (
	/* Logger Setting */

	// LogLevel is the default level of logging.
	// LogLevel = logger.InfoLevel
	LogLevel = logger.DebugLevel

	/* Metrics tracker parameters */

	// ExportInterval defines the interval we want to export the consumption data.
	ExportInterval = time.Hour

	/* Cloud monitoring query parameters */

	// DefaultResolution defines to get a datapoint every 1min
	DefaultResolution = "1m"
	// DefaultPeriod defines to get data for 1h our start time
	DefaultPeriod = "1h"
	// DefaultStart defines to get data starting from 1h15m ago.
	DefaultStart = "-75m"
)

// Configurations exported
type Configurations struct {
	ProjectID string
	BigQuery  BigQueryConfig
}

// BigQueryConfig is the config for bigquery
type BigQueryConfig struct {
	BillingProject          string
	DataSetName             string
	ConsumptionTableName    string
}

func (conf *Configurations) init() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&conf)
	if err != nil {
		return fmt.Errorf("Unable to decode into struct, %v", err)
	}
	return nil
}

// NewConfig loads config file and return the content struct.
func NewConfig() (*Configurations, error) {
	var conf Configurations
	if err := conf.init(); err != nil {
		return nil, err
	}
	return &conf, nil
}
