/*
Package cmd implements paygated commands and flags
*/
package cmd

import (
	"github.com/aavzz/daemon"
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/daemon/pid"
	"github.com/aavzz/daemon/signal"
	"github.com/aavzz/dqb-paygate/paygated/rest"
	"github.com/aavzz/dqb-paygate/paygated/storage"
	"github.com/aavzz/dqb-paygate/paygated/billing"
	"github.com/aavzz/dqb-paygate/paygated/ofd"
	"github.com/aavzz/dqb-paygate/paygated/notification"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var paygated = &cobra.Command{
	Use:   "paygated",
	Short: "paygated processes payments",
	Long:  `paygated receives payments info, does simple checks on it, registers it with fiscal authorities and stores it locally for further processing`,
	Run:   paygatedCommand,
}

func paygatedCommand(cmd *cobra.Command, args []string) {

	if viper.GetBool("daemonize") == true {
		log.InitSyslog("paygated")
		daemon.Daemonize()
	}

	//After daemon.Daemonize() this part runs in child only

	viper.SetConfigType("toml")
	viper.SetConfigFile(viper.GetString("config"))
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

        //initialization happens here
	notification.InitNotification()
        billing.InitBilling()
        ofd.InitOfd()
        storage.InitStorage()

	if viper.GetBool("daemonize") == true {
		pid.Write(viper.GetString("pidfile"))
		signal.Ignore()
		signal.Hup(func() {
			log.Info("SIGHUP received, re-reading configuration file")
			if err := viper.ReadInConfig(); err != nil {
				pid.Remove()
				log.Fatal(err.Error())
			}
		})
		signal.Term(func() {
			log.Info("SIGTERM received, exitting")
                        billing.Billing.Shutdown()
                        storage.Storage.Shutdown()
			pid.Remove()
			os.Exit(0)
		})
	}

        //InitHttp never returns
	rest.InitHttp()
}

// Execute starts paygated execution
func Execute() {
	paygated.Flags().StringP("config", "c", "/etc/paygate/paygated.conf", "configuration file")
	paygated.Flags().StringP("pidfile", "p", "/var/run/paygated.pid", "process ID file")
	paygated.Flags().StringP("address", "a", "127.0.0.1:8084", "address and port to bind to")
	paygated.Flags().BoolP("daemonize", "d", false, "run as a daemon (default false)")
	viper.BindPFlag("config", paygated.Flags().Lookup("config"))
	viper.BindPFlag("pidfile", paygated.Flags().Lookup("pidfile"))
	viper.BindPFlag("address", paygated.Flags().Lookup("address"))
	viper.BindPFlag("daemonize", paygated.Flags().Lookup("daemonize"))

	if err := paygated.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
