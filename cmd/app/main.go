package main

import (
	app "event-ordering-consumers-rabbitmq-with-consistent-hash/internal/app/consumer"
	"github.com/spf13/viper"
	"log"
)

func initConfigs() {

	//Set Configs.
	viper.SetConfigFile(`config.qa.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}

}

func main() {

	// Initialize Config File.
	initConfigs()

	// Run
	app.Run()
}
