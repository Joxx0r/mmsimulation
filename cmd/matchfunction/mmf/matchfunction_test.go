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

package mmf

import (
	"testing"

	"sim/internal/ticket"

	"github.com/stretchr/testify/require"
	"open-match.dev/open-match/pkg/pb"
)

func getRandomClientData(number int) []ticket.ClientMatchmakingData {
	returnData := []ticket.ClientMatchmakingData{}
	for i := 0; i < number; i++ {
		clientData := ticket.CreateRandomMatchmakingData()
		returnData = append(returnData, clientData)
	}
	return returnData
}

func getTicketsFromClientData(mmData []ticket.ClientMatchmakingData) []*pb.Ticket {
	returnData := []*pb.Ticket{}
	for index := range mmData {
		returnData = append(returnData, ticket.MakeTicket(mmData[index]))
	}
	return returnData
}

func getRandomTicketDataFromNum(number int) []*pb.Ticket {
	clientData := getRandomClientData(number)
	return getTicketsFromClientData(clientData)
}

func TestBasicSkill(t *testing.T) {
	require := require.New(t)
	numPlayersPerMatch := 10

	profileData := ProfileData{
		ProfileName: "test_profile",
		Region:      "europe",
		MaxPlayer:   10,
		MaxPing:     100000,
		MaxSkill:    50,
	}

	{
		clientData := getRandomClientData(numPlayersPerMatch)
		for index := range clientData {
			clientData[index].Skill = 0
			clientData[index].RegionData.Pings = map[string]float64{
				"europe": 0.0,
			}
		}
		tickets := getTicketsFromClientData(clientData)
		matches, _ := makeMatches2(tickets, profileData)
		require.True(len(matches) > 0, "Created match")
	}

	{
		tickets := getRandomTicketDataFromNum(numPlayersPerMatch / 2)
		matches, _ := makeMatches2(tickets, profileData)
		require.True(len(matches) == 0, "Did not create match with too few people")
	}

	{
		clientData := getRandomClientData(numPlayersPerMatch * 2)
		for index := range clientData {
			clientData[index].Skill = float64(index) * 1
			clientData[index].RegionData.Pings = map[string]float64{
				"europe": 0.0,
			}
		}
		tickets := getTicketsFromClientData(clientData)
		matches, _ := makeMatches2(tickets, profileData)
		require.True(len(matches) > 0, "Created match")

	}
}
