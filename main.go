// kubecos/kubecos-cli api
// https://github.com/topfreegames/kubecos/kubecos-cli
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package main

import "github.com/topfreegames/kubecos/kubecos-controller/cmd"

func main() {
	cmd.Execute(cmd.RootCmd)
}
