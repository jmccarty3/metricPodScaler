package api

import (
	"math"
	"time"
)

//Time should be in seconds
func calcDelta(startVal, endVal int64, startTime, endTime time.Time) float32 {
	return float32(endVal-startVal) / float32(endTime.Sub(startTime).Seconds())
}

func perPodDelta(delta float32, podCount int32) float32 {
	return delta / float32(podCount)
}

func estimateTargetTime(currentValue, targetValue int64, delta float32) time.Duration {
	return time.Duration(float64(targetValue-currentValue)/float64(delta)) * time.Second // Truncation?
}

func calcNeededPods(podCount int32, neededDelta, avgDelta float32) int32 {
	if avgDelta == 0 { // Goal progress is not being made.
		if neededDelta > 0 { // Below the goal; scale down
			return podCount - 1
		} else if neededDelta < 0 { // Above the goal; scale up
			return podCount + 1
		}
	}
	return int32(math.Ceil(float64(neededDelta*(1/avgDelta)))) * podCount
}
