// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"enen/game"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// gameCmd represents the game command
var gameCmd = &cobra.Command{
	Use:   "game",
	Short: "逻辑服务",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("game called")

		game.Run()
	},
}

func init() {
	RootCmd.AddCommand(gameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gameCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//gameCmd.Flags().StringP("rpcaddr", "r", "127.0.0.1:6003", "rpc服务地址")
	//gameCmd.Flags().StringP("logdir", "l", "./log", "log存放路径")
	gameCmd.Flags().StringP("name", "n", "game", "服务名称")
	//gameCmd.Flags().StringP("path", "p", "", "应用路径")
	gameCmd.Flags().BoolP("debug", "d", true, "调试模式")

	//viper.BindPFlag("game.rpcaddr", gameCmd.Flags().Lookup("rpcaddr"))
	//viper.BindPFlag("game.logdir", gameCmd.Flags().Lookup("logdir"))
	viper.BindPFlag("game.name", gameCmd.Flags().Lookup("name"))
	//viper.BindPFlag("game.path", gameCmd.Flags().Lookup("path"))
	viper.BindPFlag("game.debug", gameCmd.Flags().Lookup("debug"))
}
