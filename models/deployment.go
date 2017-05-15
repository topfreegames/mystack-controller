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
	"text/template"

	"github.com/topfreegames/mystack-controller/errors"
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
    app: {{.Name}}
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: {{.Name}}
        mystack/owner: {{.Username}}
        heritage: mystack
    spec:
      containers:
        - name: {{.Name}}
          image: {{.Image}}
          env:
            {{range .Environment}}
            - name: {{.Name}}
              value: "{{.Value}}"
            {{end}}
          ports:
            {{range .Ports}}
            - containerPort: {{.}}
            {{end}}
          {{if .ReadinessProbe}}
          readinessProbe:
            {{with .ReadinessProbe}}
            exec:
              command:
              {{range .Command}}
                - "{{.}}"
              {{end}}
            {{if .PeriodSeconds}}
            periodSeconds: {{.PeriodSeconds}}
            {{end}}{{end}}{{end}}
          {{with .Volume}}
          volumeMounts: 
            - name: {{.Name}}
              mountPath: {{.MountPath}}
          {{end}}
      {{with .Volume}}
      volumes: 
        - name: {{.Name}}
          persistentVolumeClaim:
            claimName: {{.Name}}
      {{end}}
`

//Deployment represents a deployment
type Deployment struct {
	Name           string
	Namespace      string
	Username       string
	Image          string
	Ports          []int
	Environment    []*EnvVar
	ReadinessProbe *Probe
	Volume         *VolumeMount
	Links          []*Deployment
}

//NewDeployment is the deployment ctor
func NewDeployment(
	name, username, image string,
	ports []int,
	environment []*EnvVar,
	readinessProbe *Probe,
	volume *VolumeMount,
) *Deployment {
	namespace := usernameToNamespace(username)

	return &Deployment{
		Name:           name,
		Namespace:      namespace,
		Username:       username,
		Image:          image,
		Ports:          ports,
		Environment:    environment,
		ReadinessProbe: readinessProbe,
		Volume:         volume,
		Links:          []*Deployment{},
	}
}

//Deploy creates a deployment from yaml
func (d *Deployment) Deploy(clientset kubernetes.Interface) (*v1beta1.Deployment, error) {
	if !NamespaceExists(clientset, d.Namespace) {
		err := fmt.Errorf("namespace %s not found", d.Namespace)
		return nil, errors.NewKubernetesError("create namespace error", err)
	}

	tmpl, err := template.New("deploy").Parse(deployYaml)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, d)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	decoder := api.Codecs.UniversalDecoder()
	obj, _, err := decoder.Decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	src := obj.(*extensions.Deployment)
	dst := &v1beta1.Deployment{}

	err = api.Scheme.Convert(src, dst, 0)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	deployment, err := clientset.ExtensionsV1beta1().Deployments(d.Namespace).Create(dst)
	if err != nil {
		return nil, errors.NewKubernetesError("create deployment error", err)
	}

	return deployment, nil
}

//Delete deletes deployment from cluster
func (d *Deployment) Delete(clientset kubernetes.Interface) error {
	deleteOptions := &v1.DeleteOptions{}
	err := clientset.ExtensionsV1beta1().Deployments(d.Namespace).Delete(d.Name, deleteOptions)

	if err != nil {
		return errors.NewKubernetesError("delete deployment error", err)
	}

	return nil
}
