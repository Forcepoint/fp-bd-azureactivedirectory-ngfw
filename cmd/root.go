package cmd

import (
	"fmt"
	"github.cicd.cloud.fpdev.io/BD/bd-AzureActiveDirectory-ngfw/config"
	"github.cicd.cloud.fpdev.io/BD/bd-AzureActiveDirectory-ngfw/server/elements"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var (
	cfgFile string
	Configs config.Configurations
)

var rootCmd = &cobra.Command{
	Use:   "scim",
	Short: "scim service",
	Long:  `Provide support for SCIm v2 protocol`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "path to your config file.")
	//if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
	//	log.Fatal(err.Error())
	//}
}

func initConfig() {
	viper.SetDefault("ISSUER", "ForcePoint")
	viper.SetDefault("SCIM.HOSTNAME", "localhost")
	viper.SetDefault("SCIM.PORT", "8080")
	viper.SetDefault("CONNECTOR.NAME", "smc")
	viper.SetDefault("CONNECTOR.HOSTNAME", "localhost")
	viper.SetDefault("CONNECTOR.PORT", "8085")
	viper.SetDefault("LOGGER_JSON_FORMAT", false)

	if cfgFile != "" {
		//check if the file exists
		if !elements.FileExists(cfgFile) {
			log.Fatal("the given config file does not exist")
		}
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		//viper.SetConfigName("config")
		//viper.SetConfigType("yml")
	}

	// read in environment variables that match
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Unmarshal(&Configs); err != nil {
			log.Fatal(err.Error())
		}
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			if err := viper.Unmarshal(&Configs); err != nil {
				log.Fatal(err.Error())
			}
			if viper.GetBool("LOGGER_JSON_FORMAT") {
				logrus.SetFormatter(&logrus.JSONFormatter{})
			} else {
				logrus.SetFormatter(&logrus.TextFormatter{})
			}
			elements.ProductConnector.GetEntryPoints()

		})

	}
	if viper.GetBool("LOGGER_JSON_FORMAT") {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	logrus.SetOutput(os.Stdout)
}
