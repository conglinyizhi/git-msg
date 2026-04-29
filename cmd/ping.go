package cmd

import (
	"gitmsg/internal/config"
	"gitmsg/internal/core"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "向配置文件中的 API 地址发起一次请求",
	Long:  `向配置文件 llm.toml 等配置的模型端口发送一次请求以判断是否配置正常`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetConfigValue(afero.NewOsFs())
		if err != nil {
			return err
		}
		core.SubcommandPing(cfg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
