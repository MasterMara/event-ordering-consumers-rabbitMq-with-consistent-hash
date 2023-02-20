package main

import (
	app "event-ordering-consumers-rabbitmq-with-consistent-hash/internal/app/consumer"
	"github.com/spf13/viper"
)

func initConfigs() {

	//Set Configs.
	viper.SetConfigFile(`C:\Users\musta\OneDrive\Masaüstü\event-ordering-consumers-rabbitMq-with-consistent-hash\tools\k8s-files\qa\config.qa.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

}

func main() {

	//https://luturol.github.io/docker/mongodb/Using-MongoDB-with-Docker --> Mongo

	// Initialize Config File.
	initConfigs()

	// Run
	app.Run()
}
