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

	"enen/test"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "脚本",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("test called")

		test.Run()
	},
}

func init() {
	RootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	testCmd.Flags().StringP("func", "f", "robot", "执行模块")
	testCmd.Flags().BoolP("debug", "d", true, "调试模式")
	testCmd.Flags().StringP("robot", "r", "user_1", "机器人设置[id_count]")

	viper.BindPFlag("test.func", testCmd.Flags().Lookup("func"))
	viper.BindPFlag("test.debug", testCmd.Flags().Lookup("debug"))
	viper.BindPFlag("test.robot", testCmd.Flags().Lookup("robot"))
}
