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
	"github.com/topfreegames/mystack-controller/errors"
	"text/template"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/apis/batch"
	"k8s.io/client-go/pkg/apis/batch/v1"
)

const jobYaml = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    mystack/owner: {{.Username}}
    app: {{.Name}}
    heritage: mystack
spec:
  template:
    metadata:
      name: {{.Name}}
    spec:
      containers:
      - name: {{.Name}}
        {{with .Setup}}
        image: {{.Image}}
        imagePullPolicy: {{.ImagePullPolicy}}
        {{if .Command}}
        command:
        {{range .Command}}
          - "{{.}}"
        {{end}}{{end}}
        {{end}}
        env:
          {{range .Environment}}
          - name: {{.Name}}
            value: "{{.Value}}"
          {{end}}
      restartPolicy: OnFailure
`

//Job represents a Kubernetes job
type Job struct {
	Name        string
	Namespace   string
	Username    string
	Environment []*EnvVar
	Setup       *Setup
}

//NewJob is the job ctor
func NewJob(username string, setup *Setup, environment []*EnvVar) *Job {
	if setup == nil || len(setup.Image) == 0 {
		return nil
	}
	var env []*EnvVar
	for _, setupEnv := range setup.Environment {
		env = append(env, setupEnv)
	}
	for _, environmentEnv := range environment {
		env = append(env, environmentEnv)
	}
	namespace := usernameToNamespace(username)
	return &Job{
		Name:        "setup",
		Namespace:   namespace,
		Username:    username,
		Environment: env,
		Setup:       setup,
	}
}

//Run starts the Job
func (j *Job) Run(clientset kubernetes.Interface) (*v1.Job, error) {
	if j == nil {
		return nil, nil
	}

	if !NamespaceExists(clientset, j.Namespace) {
		err := fmt.Errorf("namespace %s not found", j.Namespace)
		return nil, errors.NewKubernetesError("create namespace error", err)
	}

	tmpl, err := template.New("job").Parse(jobYaml)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, j)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	decoder := api.Codecs.UniversalDecoder()
	obj, _, err := decoder.Decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	src := obj.(*batch.Job)
	dst := &v1.Job{}

	err = api.Scheme.Convert(src, dst, 0)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	job, err := clientset.BatchV1().Jobs(j.Namespace).Create(dst)
	if err != nil {
		return nil, errors.NewKubernetesError("create job error", err)
	}

	return job, nil
}
