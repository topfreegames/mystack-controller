// mystack/mystack-cli api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package main

import "github.com/topfreegames/mystack-controller/cmd"

func main() {
	cmd.Execute(cmd.RootCmd)
}
