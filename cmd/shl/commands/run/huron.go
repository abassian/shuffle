package run

import (
	"fmt"

	"github.com/abassian/shuffle/src/consensus/huron"
	"github.com/abassian/shuffle/src/engine"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//AddHuronFlags adds flags to the Huron command
func AddHuronFlags(cmd *cobra.Command) {
	cmd.Flags().String("huron.datadir", config.Huron.DataDir, "Directory contaning priv_key.pem and peers.json files")
	cmd.Flags().String("huron.listen", config.Huron.BindAddr, "IP:PORT of Huron node")
	cmd.Flags().String("huron.service-listen", config.Huron.ServiceAddr, "IP:PORT of Huron HTTP API service")
	cmd.Flags().Duration("huron.heartbeat", config.Huron.Heartbeat, "Heartbeat time milliseconds (time between gossips)")
	cmd.Flags().Duration("huron.timeout", config.Huron.TCPTimeout, "TCP timeout milliseconds")
	cmd.Flags().Int("huron.cache-size", config.Huron.CacheSize, "Number of items in LRU caches")
	cmd.Flags().Int("huron.sync-limit", config.Huron.SyncLimit, "Max number of Events per sync")
	cmd.Flags().Bool("huron.enable-fast-sync", config.Huron.EnableFastSync, "Enable FastSync")
	cmd.Flags().Int("huron.max-pool", config.Huron.MaxPool, "Max number of pool connections")
	cmd.Flags().Bool("huron.store", config.Huron.Store, "Use persistent store")
	cmd.Flags().Bool("huron.bootstrap", config.Huron.Bootstrap, "Bootstrap from Huron database")
	viper.BindPFlags(cmd.Flags())
}

//NewHuronCmd returns the command that starts Shuffle with Huron consensus
func NewHuronCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "huron",
		Short: "Run the shuffle node with Huron consensus",
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {

			config.SetDataDir(config.BaseConfig.DataDir)

			logger.WithFields(logrus.Fields{
				"Huron": config.Huron,
			}).Debug("Config")

			return nil
		},
		RunE: runHuron,
	}

	AddHuronFlags(cmd)

	return cmd
}

func runHuron(cmd *cobra.Command, args []string) error {

	huron := huron.NewInmemHuron(config.Huron, logger)
	engine, err := engine.NewEngine(*config, huron, logger)
	if err != nil {
		return fmt.Errorf("Error building Engine: %s", err)
	}

	engine.Run()

	return nil
}
