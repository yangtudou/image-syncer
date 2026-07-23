package cmd

import (
	"fmt"
	"os"

	"github.com/AliyunContainerService/image-syncer/pkg/client"
	"github.com/AliyunContainerService/image-syncer/pkg/config"
	"github.com/AliyunContainerService/image-syncer/pkg/utils"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logPath, syncFile, successImagesFile, outputImagesFormat string

	procNum, retries int

	osFilterList, archFilterList []string

	forceUpdate bool

	debug bool
)

var RootCmd = &cobra.Command{
	Use:     "image-syncer",
	Aliases: []string{"image-syncer"},
	Short:   "A docker registry image synchronization tool",
	Long:    "A Fast and Flexible docker registry image synchronization tool implement by Go.",

	RunE: func(cmd *cobra.Command, args []string) error {

		cmd.SilenceErrors = true

		logger := client.NewFileLogger(
			logPath,
		)

		if debug {
			logger.SetLevel(
				logrus.DebugLevel,
			)
		} else {
			logger.SetLevel(
				logrus.InfoLevel,
			)
		}

		logger.Info(
			"image-syncer starting",
		)

		logger.WithFields(logrus.Fields{
			"sync":    syncFile,
			"workers": procNum,
			"retry":   retries,
			"force":   forceUpdate,
		}).Info(
			"runtime options",
		)

		if syncFile == "" {
			return fmt.Errorf(
				"sync config is required",
			)
		}

		cfg, err := config.Load(
			syncFile,
		)

		if err != nil {

			logger.WithError(err).
				Error(
					"load sync config failed",
				)

			return fmt.Errorf(
				"load sync config error: %v",
				err,
			)
		}

		logger.Info(
			"sync config loaded",
		)

		syncClient, err := client.NewSyncClient(
			cfg,
			logPath,
			successImagesFile,
			outputImagesFormat,
			procNum,
			retries,
			utils.RemoveEmptyItems(
				osFilterList,
			),
			utils.RemoveEmptyItems(
				archFilterList,
			),
			forceUpdate,
		)

		if err != nil {

			logger.WithError(err).
				Error(
					"init sync client failed",
				)

			return fmt.Errorf(
				"init sync client error: %v",
				err,
			)
		}

		logger.Info(
			"sync client initialized",
		)

		cmd.SilenceUsage = true

		logger.Info(
			"starting synchronization",
		)

		if err := syncClient.Run(); err != nil {

			logger.WithError(err).
				Error(
					"synchronization failed",
				)

			return err
		}

		logger.Info(
			"synchronization completed",
		)

		return nil
	},
}

func init() {

	RootCmd.PersistentFlags().StringVar(
		&syncFile,
		"sync",
		"",
		"sync yaml file path",
	)

	RootCmd.PersistentFlags().StringVar(
		&logPath,
		"log",
		"",
		"log file path",
	)

	RootCmd.PersistentFlags().IntVarP(
		&procNum,
		"proc",
		"p",
		5,
		"numbers of working goroutines",
	)

	RootCmd.PersistentFlags().IntVarP(
		&retries,
		"retries",
		"r",
		2,
		"times to retry failed task",
	)

	RootCmd.PersistentFlags().StringArrayVar(
		&osFilterList,
		"os",
		[]string{},
		"os list to filter source tags",
	)

	RootCmd.PersistentFlags().StringArrayVar(
		&archFilterList,
		"arch",
		[]string{},
		"architecture list to filter source tags",
	)

	RootCmd.PersistentFlags().BoolVar(
		&forceUpdate,
		"force",
		false,
		"force update manifest",
	)

	RootCmd.PersistentFlags().StringVar(
		&successImagesFile,
		"output-success-images",
		"",
		"output success images file",
	)

	RootCmd.PersistentFlags().StringVar(
		&outputImagesFormat,
		"output-images-format",
		"yaml",
		"success images output format",
	)

	RootCmd.PersistentFlags().BoolVar(
		&debug,
		"debug",
		false,
		"enable debug log",
	)
}

func Execute() {

	if err := RootCmd.Execute(); err != nil {

		fmt.Println(err)

		os.Exit(-1)
	}
}
