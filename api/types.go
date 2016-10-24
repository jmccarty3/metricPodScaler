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
	Object            ScaleObject
	ScaleDelaySeconds int32
	TargetValue       int64
	MaxReplicas       int32
	MinReplicas       int32
	Provider          providers.Provider

	delayDuration       time.Duration
	lastScaledTime      time.Time
	lastMeasuredValue   int64
	currentValue        int64
	lastMeasuredTime    time.Time
	currentMeasuredTime time.Time

	scalers []scaleFunc
}
