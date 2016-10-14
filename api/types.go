package api

import (
	"time"

	providers "github.com/jmccarty3/metricPodScaler/api/providers"
	"k8s.io/client-go/1.4/kubernetes"
)

//Config object
type Config struct {
	MasterURL  string `yaml:"masterURL"`
	KubeConfig string `yaml:"kubeConfig"`
	Scalers    []Scaler
}

//ScaleObject is a scaleable object in k8s
type ScaleObject struct {
	Namespace string
	Name      string
	Kind      string
	k8sClient *kubernetes.Clientset
}

type scaleFunc func() (bool, error)

//Scaler reprsents an object to be scaled
type Scaler struct {
	Object          ScaleObject
	TimeStepSeconds int32
	StepQuantity    int64
	MaxReplicas     int32
	Provider        providers.Provider

	lastScaledTime time.Time
	scalers        []scaleFunc
}
