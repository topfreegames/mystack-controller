// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"github.com/topfreegames/mystack-controller/errors"
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

	if err != nil {
		return errors.NewKubernetesError("create namespace error", err)
	}

	return nil
}

//DeleteNamespace delete the namespace
func DeleteNamespace(clientset kubernetes.Interface, username string) error {
	namespace := usernameToNamespace(username)
	deleteOptions := &v1.DeleteOptions{}

	err := clientset.CoreV1().Namespaces().Delete(namespace, deleteOptions)
	if err != nil {
		return errors.NewKubernetesError("delete namespace error", err)
	}

	return nil
}

//ListNamespaces returns a list of namespaces
func ListNamespaces(clientset kubernetes.Interface) (*v1.NamespaceList, error) {
	list, err := clientset.CoreV1().Namespaces().List(listOptions)
	if err != nil {
		return nil, errors.NewKubernetesError("list namespaces error", err)
	}
	return list, nil
}

//NamespaceExists return true if namespace is already created
func NamespaceExists(clientset kubernetes.Interface, namespace string) bool {
	_, err := clientset.CoreV1().Namespaces().Get(namespace)
	return err == nil
}
