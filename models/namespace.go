// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/labels"
)

var (
	labelMap    = labels.Set{"mystack/routable": "true"}
	listOptions = v1.ListOptions{
		LabelSelector: labelMap.AsSelector().String(),
		FieldSelector: fields.Everything().String(),
	}
)

//CreateNamespace creates a namespace
func CreateNamespace(clientset kubernetes.Interface, username string) error {
	namespaceStr := usernameToNamespace(username)
	namespace := &v1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: namespaceStr,
			Labels: map[string]string{
				"mystack/routable": "true",
			},
		},
	}
	_, err := clientset.CoreV1().Namespaces().Create(namespace)
	return err
}

//DeleteNamespace delete the namespace
func DeleteNamespace(clientset kubernetes.Interface, username string) error {
	namespace := usernameToNamespace(username)
	deleteOptions := &v1.DeleteOptions{}
	return clientset.CoreV1().Namespaces().Delete(namespace, deleteOptions)
}

//ListNamespaces returns a list of namespaces
func ListNamespaces(clientset kubernetes.Interface) (*v1.NamespaceList, error) {
	return clientset.CoreV1().Namespaces().List(listOptions)
}

//NamespaceExists return true if namespace is already created
func NamespaceExists(clientset kubernetes.Interface, namespace string) bool {
	_, err := clientset.CoreV1().Namespaces().Get(namespace)
	return err == nil
}
