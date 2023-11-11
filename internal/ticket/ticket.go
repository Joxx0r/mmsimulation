// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ticket

import (
	// Uncomment if following the tutorial
	// "math/rand"

	"sim/cmd/frontend/client"
	utils "sim/internal"
	"sim/internal/random"

	"google.golang.org/protobuf/types/known/anypb"
	"open-match.dev/open-match/pkg/pb"
)

type ClientMatchmakingData struct {
	RegionData client.ClientRegionData
	Trusted    string
	Password   string
	Skill      float64
	GameMode   string
	Beginner   bool
}

func CreateRandomMatchmakingData() ClientMatchmakingData {
	returnData := ClientMatchmakingData{
		RegionData: client.ClientRegionData{
			Pings: make(map[string]float64),
		},
		Trusted:  random.FindTrustedState(),
		Password: random.FindPassword(),
		Skill:    random.FindSkill(),
		GameMode: random.FindGameMode(),
	}

	allRegions := utils.GRegions

	for _, region := range allRegions {
		returnData.RegionData.Pings[region] = random.FindRegionRandom(region)
	}

	return returnData
}

// Ticket generates a Ticket with a mode search field that has one of the
// randomly selected modes.
func MakeTicket(clientData ClientMatchmakingData) *pb.Ticket {
	ticket := &pb.Ticket{
		SearchFields: &pb.SearchFields{
			// Tags can support multiple values but for simplicity, the demo function
			// assumes only single mode selection per Ticket.
			Tags: []string{
				clientData.GameMode,
				clientData.Trusted,
			},
			DoubleArgs: map[string]float64{
				utils.GSkillArg: clientData.Skill,
			},
		},
		Extensions: make(map[string]*anypb.Any),
	}

	if clientData.Password != "" {
		ticket.SearchFields.Tags = append(ticket.SearchFields.Tags, clientData.Password)
	}

	desiredRegions := client.GetDesiredRegions(clientData.RegionData.Pings)
	if len(desiredRegions) == 0 {
		panic("expected regions to be filled in ")
	}

	for index := range desiredRegions {
		ticket.SearchFields.Tags = append(ticket.SearchFields.Tags, desiredRegions[index].Region)
	}
	utils.AddExtensionString(ticket.Extensions, utils.GBestRegionKey, desiredRegions[0].Region)
	for region, v := range clientData.RegionData.Pings {
		utils.AddExtensionFloat64(ticket.Extensions, region, float64(v))
	}
	if len(desiredRegions) > 0 {
	}

	return ticket
}

func GetSkillFromTicket(t *pb.Ticket) float64 {
	return t.SearchFields.DoubleArgs[utils.GSkillArg]
}

func GetLatencyFromTicket(t *pb.Ticket, region string, bestRegionMaxPing int) float64 {
	regionPing := utils.GetExtensionFloat64(t.Extensions, region)
	bestRegion := utils.GetExtensionString(t.Extensions, utils.GBestRegionKey)
	if bestRegion == region && regionPing > float64(bestRegionMaxPing) {
		return float64(bestRegionMaxPing)
	}
	return regionPing
}
