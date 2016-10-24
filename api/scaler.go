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
	s.delayDuration = time.Duration(s.ScaleDelaySeconds) * time.Second
	s.Object.Init(masterURL, kubeConfig)
}

//Run main control loop for scaler
func (s *Scaler) Run() {
	s.lastScaledTime = time.Now()
	s.updateValues() // Get initial values
	for {
		time.Sleep(s.delayDuration) //Take a breather
		if s.Object.Exists() {
			if s.updateValues() != nil {
				continue
			}
			s.work()
		}
	}
}

func (s *Scaler) updateValues() error {
	val, err := s.Provider.CurrentCount()
	if err != nil {
		glog.Errorf("Could not update values: %v", err)
		return err
	}
	s.lastMeasuredTime = s.currentMeasuredTime
	s.lastMeasuredValue = s.currentValue
	s.currentValue = val
	s.currentMeasuredTime = time.Now()

	glog.V(4).Infof("Last Time/Value: %v/%v  Current Time/Value: %v/%v", s.lastMeasuredTime, s.lastMeasuredValue, s.currentMeasuredTime, s.currentValue)
	return nil
}

func (s *Scaler) getPods() int32 {
	currentPods := s.Object.CurrentCount()
	d := calcDelta(s.lastMeasuredValue, s.currentValue,
		s.lastMeasuredTime, s.currentMeasuredTime)

	targetTimeSecs := estimateTargetTime(s.currentValue, s.TargetValue, d)
	deltaPerPod := perPodDelta(d, currentPods)
	goalDelta := calcDelta(s.lastMeasuredValue, s.TargetValue, s.lastMeasuredTime, time.Now().Add(60*time.Second))
	newPodCount := calcNeededPods(currentPods, goalDelta, deltaPerPod)
	glog.V(4).Infof("Time Diff: %v", s.currentMeasuredTime.Sub(s.lastMeasuredTime).Seconds())
	glog.V(4).Infof("Current Pods: %d", currentPods)
	glog.V(4).Infof("Delta: %v", d)
	glog.V(4).Infof("Intercept time estimate: %v seconds", targetTimeSecs)
	glog.V(4).Infof("PerPodDelta: %v", deltaPerPod)
	glog.V(4).Infof("Needed Delta: %v", goalDelta)

	if targetTimeSecs > 0 && targetTimeSecs <= s.delayDuration {
		glog.V(2).Info("Intercept time is within threshold. Do nothing")
		return currentPods
	}

	glog.V(4).Infof("Current Pods: %d Desired Pods: %v", currentPods, newPodCount)
	return newPodCount
}

func inThreshold(target, quantity int64, thresholdPercent float64) bool {
	jitter := int64(math.Ceil(float64(target) * thresholdPercent))
	min := int64(math.Max(0, float64(target-jitter)))
	max := target + jitter
	return quantity > min && quantity < max
}

/*
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
*/

func (s *Scaler) work() {
	s.scale(s.getPods())
}

func (s *Scaler) scale(desiredSize int32) bool {
	currentObjects := s.Object.CurrentCount()
	glog.V(2).Infof("Current object count: %d Max Count: %d", currentObjects, s.MaxReplicas)

	if desiredSize >= s.MaxReplicas {
		glog.V(4).Info("Desired size greater then max. Setting to max")
		desiredSize = s.MaxReplicas
	}

	if desiredSize < s.MinReplicas {
		glog.V(4).Info("Desired size less then min. Setting to min.")
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
	Object            ScaleObject
	ScaleDelaySeconds int32 `yaml:"scaleDelaySeconds"`
	TargetValue       int64 `yaml:"targetValue"`
	MaxReplicas       int32 `yaml:"maxReplicas"`
	MinReplicas       int32 `yaml:"minReplicas"`
	Provider          map[string]interface{}
}

//UnmarshalYAML performs custom unmarshalling
func (s *Scaler) UnmarshalYAML(unmarshal func(interface{}) error) error {
	in := &scalerYaml{
		ScaleDelaySeconds: DefaultCooldown,
	}

	if err := unmarshal(&in); err != nil {
		return err
	}

	var q providers.Provider
	for k, v := range in.Provider {
		glog.Info(providers.Providers)
		q = providers.Providers[k]()
		reMarsh, _ := yaml.Marshal(v)
		if err := yaml.Unmarshal(reMarsh, q); err != nil {
			return err
		}
	}

	s.Object = in.Object
	s.ScaleDelaySeconds = in.ScaleDelaySeconds
	s.MaxReplicas = in.MaxReplicas
	s.MinReplicas = in.MinReplicas
	s.Provider = q
	return nil
}
