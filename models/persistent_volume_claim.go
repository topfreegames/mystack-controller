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
	"html/template"

	"github.com/topfreegames/mystack-controller/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

const persistentVolumeClaimYaml = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  annotations:
    volume.alpha.kubernetes.io/storage-class: "gp2"
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{.Storage}}
`

//PersistentVolumeClaim gets volume configs from yaml
type PersistentVolumeClaim struct {
	Name      string `yaml:"name"`
	Storage   string `yaml:"storage"`
	Namespace string
}

//NewPVC is the PersistentVolumeClaim constructor
func NewPVC(name, username, storage string) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{
		Name:      name,
		Namespace: usernameToNamespace(username),
		Storage:   storage,
	}
}

func (p *PersistentVolumeClaim) Start(clientset kubernetes.Interface) (*v1.PersistentVolumeClaim, error) {
	if !NamespaceExists(clientset, p.Namespace) {
		err := fmt.Errorf("namespace %s not found", p.Namespace)
		return nil, errors.NewKubernetesError("create namespace error", err)
	}

	tmpl, err := template.New("pvc").Parse(persistentVolumeClaimYaml)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, p)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	decoder := api.Codecs.UniversalDecoder()
	obj, _, err := decoder.Decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	src := obj.(*api.PersistentVolumeClaim)
	dst := &v1.PersistentVolumeClaim{}

	err = api.Scheme.Convert(src, dst, 0)
	if err != nil {
		return nil, errors.NewYamlError("parse yaml error", err)
	}

	pvc, err := clientset.CoreV1().PersistentVolumeClaims(p.Namespace).Create(dst)
	if err != nil {
		return nil, errors.NewKubernetesError("create deployment error", err)
	}

	return pvc, nil
}
