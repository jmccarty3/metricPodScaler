package api

import (
	"math"
	"time"

	"gopkg.in/yaml.v1"

	"github.com/golang/glog"
	"github.com/jmccarty3/metricPodScaler/api/providers"
)

//DefaultCooldown time between run cycles
const DefaultCooldown = 60

//DefaultThresholdPercent is a threshold to avoid thrashing on value change
const DefaultThresholdPercent = .10

//Init initializes the scaler object
//TODO Consider just passing a client object
func (s *Scaler) Init(masterURL, kubeConfig string) {
	s.Object.Init(masterURL, kubeConfig)
}

//Run main control loop for scaler
func (s *Scaler) Run() {
	s.scalers = []scaleFunc{s.scaleByValue, s.scaleByTime}
	s.lastScaledTime = time.Now()

	for {
		if s.Object.Exists() {
			s.work()
		}

		time.Sleep(DefaultCooldown * time.Second) //Take a breather
	}
}

func inThreshold(target, quantity int64, thresholdPercent float64) bool {
	jitter := int64(math.Ceil(float64(target) * thresholdPercent))
	min := int64(math.Max(0, float64(target-jitter)))
	max := target + jitter
	return quantity > min && quantity < max
}

func (s *Scaler) inThreshold(quantity int64) bool {
	baseLimit := int64(s.Object.CurrentCount()) * s.StepQuantity
	percentage := int64(float32(baseLimit) * DefaultThresholdPercent)

	if quantity > (percentage + baseLimit) {
		return false
	}
	if quantity < (baseLimit - percentage) {
		return false
	}
	return true
}

func (s *Scaler) scaleByValue() (bool, error) {
	currentCount, err := s.Provider.CurrentCount()
	if err != nil {
		return false, err
	}

	if inThreshold(int64(s.Object.CurrentCount())*s.StepQuantity, currentCount, DefaultThresholdPercent) {
		glog.V(2).Info("Current count in threshold")
		return false, nil
	}
	return s.scale(int32(currentCount / s.StepQuantity)), nil
}

func (s *Scaler) scaleByTime() (bool, error) {
	if s.TimeStepSeconds == 0 {
		glog.V(4).Info("No timestep given. skipping")
		return false, nil
	}
	if s.lastScaledTime.Add(time.Duration(s.TimeStepSeconds) * time.Second).After(time.Now()) {
		glog.V(4).Info("Time threshold fine")
		return false, nil
	}
	glog.V(2).Info("Time threshold crossed. Scaling.")
	return s.scale(s.Object.CurrentCount() + 1), nil
}

func (s *Scaler) work() {
	for _, scaler := range s.scalers {
		scaled, err := scaler()

		if scaled {
			return
		}
		if err != nil {
			glog.Errorf("Could not scale: %v", err)
		}
	}
}

func (s *Scaler) scale(desiredSize int32) bool {
	currentObjects := s.Object.CurrentCount()
	glog.V(2).Infof("Current object count: %d Max Count: %d", currentObjects, s.MaxReplicas)

	if desiredSize >= s.MaxReplicas {
		glog.V(4).Info("Desired size greater then max. Setting to max")
		desiredSize = s.MaxReplicas
	}

	if desiredSize == currentObjects {
		glog.V(4).Info("Desired size matches current size. Skipping")
		return false
	}

	s.Object.Scale(desiredSize)
	s.lastScaledTime = time.Now()
	return true
}

type scalerYaml struct {
	Object          ScaleObject
	TimeStepSeconds int32 `yaml:"timeStepSeconds"`
	StepQuantity    int64 `yaml:"stepQuantity"`
	MaxReplicas     int32 `yaml:"maxReplicas"`
	Provider        map[string]interface{}
}

//UnmarshalYAML performs custom unmarshalling
func (s *Scaler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	in := &scalerYaml{}

	if err := unmarshal(&in); err != nil {
		return err
	}

	var q providers.Provider
	for k, v := range in.Provider {
		q = providers.Providers[k]()
		reMarsh, _ := yaml.Marshal(v)
		if err := yaml.Unmarshal(reMarsh, q); err != nil {
			return err
		}
	}
	glog.Infof("%v", in.Object)
	glog.Info(in)
	s.Object = in.Object
	s.TimeStepSeconds = in.TimeStepSeconds
	s.StepQuantity = in.StepQuantity
	s.MaxReplicas = in.MaxReplicas
	s.Provider = q
	return nil
}
