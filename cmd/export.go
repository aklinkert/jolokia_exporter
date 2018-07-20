// Copyright © 2017 Alexander Pinnecke <alexander.pinnecke@googlemail.com>
//

package cmd

import (
	"net/http"

	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/scalify/jolokia_exporter/jolokia"
	"github.com/spf13/cobra"
)

var (
	verbose           bool
	insecure          bool
	basicAuthUser     string
	basicAuthPassword string
	scrapeListen      string
	scrapeEndpoint    string
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export <metrics-config-file> <endpoint>",
	Short: "Exports jolokia metrics from given endpoint, using given metrics mapping config",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cmd.Usage()
			os.Exit(1)
		}

		configFile := args[0]
		config, err := jolokia.LoadConfig(configFile)
		if err != nil {
			panic(err)
		}

		endpoint := args[1]

		logger := log.Base()
		if verbose {
			logger.SetLevel("debug")
			logger.Debug("Starting in debug level")
		} else {
			logger.SetLevel("info")
		}
		exp, err := jolokia.NewExporter(logger, config, jolokia.Namespace, insecure, endpoint, basicAuthUser, basicAuthPassword)
		if err != nil {
			panic(err)
		}

		prometheus.MustRegister(exp)
		prometheus.MustRegister(version.NewCollector("jolokia_exporter"))

		log.Infof("Exporting jolokia endpoint: %v", endpoint)
		log.Info("Starting jolokia_exporter", version.Info())
		log.Info("Build context", version.BuildContext())
		log.Infof("Starting Server: %s", scrapeListen)

		http.Handle(scrapeEndpoint, promhttp.Handler())
		log.Fatal(http.ListenAndServe(scrapeListen, nil))
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)

	exportCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Whether to use verbose https mode")
	exportCmd.Flags().BoolVarP(&insecure, "insecure", "i", false, "Whether to use insecure https mode, i.e. skip ssl cert validation (only useful with https endpoint)")
	exportCmd.Flags().StringVar(&basicAuthUser, "basic-auth-user", "", "HTTP Basic auth user for authentication on the jolokia endpoint")
	exportCmd.Flags().StringVar(&basicAuthPassword, "basic-auth-password", "", "HTTP Basic auth password for authentication on the jolokia endpoint")
	exportCmd.Flags().StringVarP(&scrapeListen, "listen", "l", ":9422", "Host/Port the exporter should listen listen on")
	exportCmd.Flags().StringVarP(&scrapeEndpoint, "endpoint", "e", "/metrics", "Path the exporter should listen listen on")
}
