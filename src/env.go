// Copyright Epic Games, Inc. All Rights Reserved.

package main

import (
	"log"

	"github.com/spf13/viper"
	"github.com/winterspite/ssrf-sheriff/src/constants"
)

func initEnv() {
	viper.AutomaticEnv()

	viper.SetDefault(constants.EnvAddr, ":8000")
	viper.SetDefault(constants.EnvSSRFToken, "insert-your-ssrf-token-here")
	viper.SetDefault(constants.EnvWebhookURL, "insert-slack-webhook-here")
	viper.SetDefault(constants.EnvHealthcheckURL, "insert-a-healthcheck-url-here")
	viper.SetDefault(constants.EnvLoggingFormat, "console")

	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}
}
