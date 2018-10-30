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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// gmtCmd represents the gmt command
var gmtCmd = &cobra.Command{
	Use:   "gmt",
	Short: "管理服务",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gmt called")
	},
}

func init() {
	RootCmd.AddCommand(gmtCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gmtCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gmtCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	gmtCmd.Flags().StringP("name", "n", "gmt", "服务名称")
	gmtCmd.Flags().BoolP("debug", "d", true, "调试模式")

	viper.BindPFlag("gmt.name", gmtCmd.Flags().Lookup("name"))
	viper.BindPFlag("gmt.debug", gmtCmd.Flags().Lookup("debug"))
}
