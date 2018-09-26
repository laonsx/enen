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

package main

import (
	"log"

	"enen/cmd"
)

const LOGO = `
  ___  ____  ___  ____ 
 / _ \/ __ \/ _ \/ __ \
/  __/ / / /  __/ / / /
\___/_/ /_/\___/_/ /_/ 

Contact: cn.laonsx@gmail.com
Version: v0.0.2 -- master(f19751d)
`

func main() {

	log.Println(LOGO)

	cmd.Execute()
}
