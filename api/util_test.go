package api

import (
	"testing"
	"time"
)

func TestEstimateTargetTime(t *testing.T) {
	tests := []struct {
		Current      int64
		Target       int64
		Delta        float32
		ExpectedTime time.Duration
	}{
		{
			Current:      10,
			Target:       0,
			Delta:        -1,
			ExpectedTime: 10 * time.Second,
		},
		{
			Current:      10,
			Target:       0,
			Delta:        1,
			ExpectedTime: -10 * time.Second,
		},
		{
			Current:      10,
			Target:       0,
			Delta:        -.5,
			ExpectedTime: 20 * time.Second,
		},
	}

	for _, test := range tests {
		if actual := estimateTargetTime(test.Current, test.Target, test.Delta); actual != test.ExpectedTime {
			t.Errorf("Expected: %v Actual %v", test.ExpectedTime, actual)
		}
	}
}

func TestCalcNeededPods(t *testing.T) {
	tests := []struct {
		Pods        int32
		NeededDelta float32
		AvgDelta    float32
		Expected    int32
	}{
		{
			Pods:        1,
			NeededDelta: 2,
			AvgDelta:    1,
			Expected:    2,
		},
		{
			Pods:        10,
			NeededDelta: .5,
			AvgDelta:    .1,
			Expected:    50,
		},
		{
			Pods:        1,
			NeededDelta: 1,
			AvgDelta:    0,
			Expected:    0,
		},
		{
			Pods:        1,
			NeededDelta: -1,
			AvgDelta:    0,
			Expected:    2,
		},
	}

	for _, test := range tests {
		if actual := calcNeededPods(test.Pods, test.NeededDelta, test.AvgDelta); actual != test.Expected {
			t.Errorf("Expected: %v Actual %v", test.Expected, actual)
		}
	}
}
