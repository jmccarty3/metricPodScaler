package api

import (
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/tools/clientcmd"
)

const defaultWaitPeriod = 30

//Init performs initializaion work
func (scale *ScaleObject) Init(masterURL, configFile string) {
	//TODO Support more than just in cluster
	config, err := clientcmd.BuildConfigFromFlags(masterURL, configFile)
	if err != nil {
		glog.Exit(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Exit(err)
	}
	glog.V(2).Info("Connected to k8s")

	scale.k8sClient = clientset
}

//Exists checks that the object actually exists to act on
func (scale *ScaleObject) Exists() bool {
	_, err := scale.k8sClient.Scales(scale.Namespace).Get(scale.Kind, scale.Name)
	glog.Infof("Looking for %s/%s/%s", scale.Kind, scale.Namespace, scale.Name)
	if err != nil {
		glog.V(2).Infof("Scale object %s may not exist. Error: %v", scale.Name, err)
		return false
	}
	return true
}

//CurrentCount checks the current count of the scalable object
func (scale *ScaleObject) CurrentCount() int32 {
	s, _ := scale.k8sClient.Scales(scale.Namespace).Get(scale.Kind, scale.Name)
	return s.Status.Replicas
}

//Scale attempts to adjust the size of the scalable object
func (scale *ScaleObject) Scale(newSize int32) {
	s, _ := scale.k8sClient.Scales(scale.Namespace).Get(scale.Kind, scale.Name)
	s.Spec.Replicas = newSize
	scale.k8sClient.Scales(scale.Namespace).Update(scale.Kind, s)
	glog.Infof("Scaled %s to %d", scale.Name, newSize)
	for scale.CurrentCount() != newSize {
		glog.V(2).Info("Waiting for scale to finish")
		time.Sleep(defaultWaitPeriod * time.Second)
	}
}
