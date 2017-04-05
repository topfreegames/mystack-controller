// mystack-controller api
// https://github.com/topfreegames/mystack-controller
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2017 Top Free Games <backend@tfgco.com>

package models

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

const deployYaml = `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    mystack/routable: "true"
    mystack/owner: {{.Username}}
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
        - name: {{.Name}}
          image: {{.Image}}
          ports:
            - containerPort: {{.Port}}
`

//Deployment represents a deployment
type Deployment struct {
	Name      string
	Namespace string
	Username  string
	Image     string
	Port      int
}

//NewDeployment is the deployment ctor
func NewDeployment(name, username, image string, port int) *Deployment {
	username = strings.Replace(username, ".", "-", -1)
	namespace := usernameToNamespace(username)
	return &Deployment{
		Name:      name,
		Namespace: namespace,
		Username:  username,
		Image:     image,
		Port:      port,
	}
}

//Deploy creates a deployment from yaml
func (d *Deployment) Deploy(clientset kubernetes.Interface) (*v1beta1.Deployment, error) {
	if !NamespaceExists(clientset, d.Namespace) {
		err := fmt.Errorf("Namespace %s not found", d.Namespace)
		return nil, err
	}

	tmpl, err := template.New("deploy").Parse(deployYaml)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, d)
	if err != nil {
		return nil, err
	}

	decoder := api.Codecs.UniversalDecoder()
	obj, _, err := decoder.Decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, err
	}

	src := obj.(*extensions.Deployment)
	dst := &v1beta1.Deployment{}

	err = api.Scheme.Convert(src, dst, 0)
	if err != nil {
		return nil, err
	}

	return clientset.ExtensionsV1beta1().Deployments(d.Namespace).Create(dst)
}

//Delete deletes deployment from cluster
func (d *Deployment) Delete(clientset kubernetes.Interface) error {
	deleteOptions := &v1.DeleteOptions{}
	return clientset.ExtensionsV1beta1().Deployments(d.Namespace).Delete(d.Namespace, deleteOptions)
}
