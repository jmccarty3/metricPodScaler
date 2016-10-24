package k8s

import (
	"github.com/golang/glog"
	"github.com/jmccarty3/metricPodScaler/api/providers"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/labels"
	"k8s.io/client-go/1.4/tools/clientcmd"
)

type podCount struct {
	MasterURL  string            `yaml:"masterURL"`
	KubeConfig string            `yaml:"kubeConfig"`
	Namespace  string            `yaml:"namespace"`
	Labels     map[string]string `yaml:"labels"`

	client        *kubernetes.Clientset
	labelSelector labels.Selector
}

func init() {
	providers.AddProvider("podCount", func() providers.Provider {
		return &podCount{
			Namespace: "default",
			Labels:    make(map[string]string),
		}
	})
}

func (p *podCount) Connect() error {
	glog.V(2).Info("Making Pod")
	config, err := clientcmd.BuildConfigFromFlags(p.MasterURL, p.KubeConfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	glog.V(2).Info("Connected to k8s")

	p.client = clientset
	p.labelSelector = labels.SelectorFromSet(p.Labels)
	glog.V(4).Infof("LabelSelector: %v", p.labelSelector)
	return nil
}

func (p *podCount) CurrentCount() (int64, error) {
	glog.V(4).Info("Fetching pod count")
	pods, err := p.client.Pods(p.Namespace).List(api.ListOptions{
		LabelSelector: p.labelSelector,
	})

	if err != nil {
		return 0, err
	}
	return int64(len(pods.Items)), nil
}
