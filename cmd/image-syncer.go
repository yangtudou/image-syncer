package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AliyunContainerService/image-syncer/pkg/client"
	"github.com/AliyunContainerService/image-syncer/pkg/config"
	"github.com/AliyunContainerService/image-syncer/pkg/mapper"
	"github.com/AliyunContainerService/image-syncer/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	logPath, authFile, syncFile, successImagesFile, outputImagesFormat string

	procNum, retries int

	osFilterList, archFilterList []string

	forceUpdate bool
)

var RootCmd = &cobra.Command{
	Use:     "image-syncer",
	Aliases: []string{"image-syncer"},
	Short:   "A docker registry image synchronization tool",
	Long:    "A Fast and Flexible docker registry image synchronization tool implement by Go.",

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceErrors = true

		if syncFile == "" {
			return fmt.Errorf("sync config is required")
		}

		if authFile == "" {
			return fmt.Errorf("auth file is required")
		}

		cfg, err := config.Load(syncFile)
		if err != nil {
			return fmt.Errorf("load sync config error: %v", err)
		}

		mappings := mapper.Generate(cfg)

		if len(mappings) == 0 {
			return fmt.Errorf("no image mappings found")
		}

		tempDir, err := os.MkdirTemp("", "image-syncer-*")
		if err != nil {
			return fmt.Errorf("create temp dir error: %v", err)
		}

		defer os.RemoveAll(tempDir)

		imagesFile := filepath.Join(tempDir, "images.yaml")

		if err := mapper.WriteImageSyncer(imagesFile, mappings); err != nil {
			return fmt.Errorf("write generated images config error: %v", err)
		}

		syncClient, err := client.NewSyncClient(
			"",
			authFile,
			imagesFile,
			logPath,
			successImagesFile,
			outputImagesFormat,
			procNum,
			retries,
			utils.RemoveEmptyItems(osFilterList),
			utils.RemoveEmptyItems(archFilterList),
			forceUpdate,
		)

		if err != nil {
			return fmt.Errorf("init sync client error: %v", err)
		}

		cmd.SilenceUsage = true

		return syncClient.Run()
	},
}

func init() {
	RootCmd.PersistentFlags().StringVar(
		&authFile,
		"auth",
		"",
		"auth file path",
	)

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
		"log file path (default in os.Stderr)",
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
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
