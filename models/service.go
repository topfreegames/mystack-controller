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
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/topfreegames/mystack-controller/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

const serviceYaml = `
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    mystack/routable: "true"
    mystack/service: "{{.IsMystackSvc}}"
    mystack/socket: "{{.IsSocket}}"
    mystack/socketPorts: "{{.SocketPorts}}"
spec:
  selector:
    app: {{.Name}}
  ports:
    {{range .Ports}}
    - name: {{.Name}}
      protocol: TCP
      port: {{.Port}}
      targetPort: {{.TargetPort}}
    {{end}}
  type: ClusterIP
`

//PortMap maps a port to a target por on service
type PortMap struct {
	Port       int
	TargetPort int
	Name       string
}

//Service represents a service
type Service struct {
	Name         string
	Namespace    string
	Ports        []*PortMap
	IsMystackSvc bool
	SocketPorts  string
	IsSocket     bool
}

//NewService is the service ctor
func NewService(name, username string, ports []*PortMap, isMystackSvc, isSocket bool) *Service {
	namespace := usernameToNamespace(username)
	for i, port := range ports {
		if len(port.Name) == 0 {
			port.Name = fmt.Sprintf("port-%d", i)
		}
	}

	var socketPorts []string
	if isSocket {
		socketPorts = make([]string, len(ports))
		for idx := range ports {
			r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
			socketPort := r1.Intn(20000) + 40000
			socketPorts[idx] = fmt.Sprintf("%d", socketPort)
		}
	}

	return &Service{
		Name:         name,
		Namespace:    namespace,
		Ports:        ports,
		IsMystackSvc: isMystackSvc,
		IsSocket:     isSocket,
		SocketPorts:  strings.Join(socketPorts, ","),
	}
}

//Expose exposes a deployment
func (s *Service) Expose(clientset kubernetes.Interface) (*v1.Service, error) {
	tmpl, err := template.New("expose").Parse(serviceYaml)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	decoder := api.Codecs.UniversalDecoder()
	obj, _, err := decoder.Decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	src := obj.(*api.Service)
	dst := &v1.Service{}

	err = api.Scheme.Convert(src, dst, 0)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	service, err := clientset.CoreV1().Services(s.Namespace).Create(dst)

	if err != nil {
		return nil, errors.NewKubernetesError("create service error", err)
	}

	return service, nil
}

//Delete deletes service
func (s *Service) Delete(clientset kubernetes.Interface) error {
	deleteOptions := &v1.DeleteOptions{}

	err := clientset.CoreV1().Services(s.Namespace).Delete(s.Name, deleteOptions)
	if err != nil {
		return errors.NewKubernetesError("create service error", err)
	}

	return nil
}

// ServicePort ...
func ServicePort(clientset kubernetes.Interface, name, username string) (int, error) {
	namespace := usernameToNamespace(username)
	service, err := clientset.CoreV1().Services(namespace).Get(name)
	if err != nil {
		return 0, err
	}

	port := service.Spec.Ports[0].Port
	return int(port), nil
}
