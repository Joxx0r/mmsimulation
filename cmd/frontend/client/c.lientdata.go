package client

import (
	"sort"
)

type ClientRegionData struct {
	Pings map[string]float64
}

type PingData struct {
	Region string
	Ping   float64
}

func GetDesiredRegions(Pings map[string]float64) []PingData {
	returnData := []PingData{}
	bestRegionName := ""
	bestPing := 100000.0
	for k := range Pings {
		value := Pings[k]
		if bestPing > value {
			bestPing = value
			bestRegionName = k
		}
		// allow to queue towards region that are smaller
		if value < 300 {
			returnData = append(returnData, PingData{
				Region: k,
				Ping:   value,
			})
		}
	}

	// if we found 0 regions with below x ping, pick the one that was best
	if len(returnData) == 0 {
		returnData = append(returnData, PingData{Region: bestRegionName, Ping: bestPing})
	}

	sort.SliceStable(returnData, func(i int, j int) bool {
		return returnData[i].Ping < returnData[j].Ping
	})

	return returnData
}
