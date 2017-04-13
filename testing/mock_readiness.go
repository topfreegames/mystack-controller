// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package testing

import "k8s.io/client-go/kubernetes"

//MockReadiness implements Readiness interface
type MockReadiness struct{}

//WaitForCompletion waits until job has completed its task
func (*MockReadiness) WaitForCompletion(kubernetes.Interface, interface{}) error {
	return nil
}
