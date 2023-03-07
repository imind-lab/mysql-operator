package utils

import (
	"bytes"
	"text/template"

	"github.com/imind-lab/mysql-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func parseTemplate(templateName string, cluster *v1beta1.MySQLCluster) []byte {
	tmpl, err := template.ParseFiles("controllers/template/" + templateName + ".yaml")
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, cluster)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func NewService(cluster *v1beta1.MySQLCluster) *corev1.Service {
	svc := &corev1.Service{}
	err := yaml.Unmarshal(parseTemplate("service", cluster), svc)
	if err != nil {
		panic(err)
	}
	return svc
}

func NewStatefulSet(cluster *v1beta1.MySQLCluster) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{}
	err := yaml.Unmarshal(parseTemplate("statefulset", cluster), sts)
	if err != nil {
		panic(err)
	}
	return sts
}

func NewConfigMap(cluster *v1beta1.MySQLCluster) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{}
	err := yaml.Unmarshal(parseTemplate("configmap", cluster), cm)
	if err != nil {
		panic(err)
	}
	return cm
}
