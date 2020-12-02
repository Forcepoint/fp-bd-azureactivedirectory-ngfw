package cmd

import (
	"github.cicd.cloud.fpdev.io/BD/bd-AzureActiveDirectory-ngfw/server/elements"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run SCIM service",
	Long: `the run command reads the configs and runs the SCIM service. For example:
scim run --config <path_to_config_file> --key <access_key_of_forcepoint_product>`,
	Run: func(cmd *cobra.Command, args []string) {
		time.Sleep(5 * time.Second)
		elements.ProductConnector.GetEntryPoints()
		muxRouter := mux.NewRouter().StrictSlash(true)
		router := elements.AddRoutes(muxRouter)
		err := http.ListenAndServe(viper.GetString("SCIM.HOSTNAME")+":"+viper.GetString("SCIM.PORT"), router)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
