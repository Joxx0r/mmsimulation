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
	"fmt"
	"log"
	"sort"
	"time"

	"sim/internal/ticket"

	utils "sim/internal"

	// Uncomment if following the tutorial
	// "fmt"
	// "time"

	"google.golang.org/protobuf/types/known/anypb"
	"open-match.dev/open-match/pkg/matchfunction"
	"open-match.dev/open-match/pkg/pb"
)

const (
	matchName = "basic-matchfunction"
)

type CalculationMode int

const (
	All CalculationMode = iota
	Skill
)

type ProfileData struct {
	ProfileName string
	Region      string
	MaxPlayer   int
	MaxPing     int
	MaxSkill    int
}

var GCalculationMode = Skill

// Run is this match function's implementation of the gRPC call defined in api/matchfunction.proto.
func (s *MatchFunctionService) Run(req *pb.RunRequest, stream pb.MatchFunction_RunServer) error {
	// Fetch tickets for the pools specified in the Match Profile.
	log.Printf("Generating proposals for function %v", req.GetProfile().GetName())

	poolTickets, err := matchfunction.QueryPools(stream.Context(), s.queryServiceClient, req.GetProfile().GetPools())
	if err != nil {
		log.Printf("Failed to query tickets for the given pools, got %s", err.Error())
		return err
	}

	matchProfile := req.GetProfile()

	profileData := ProfileData{
		ProfileName: matchProfile.Name,
		Region:      utils.GetExtensionString(matchProfile.Extensions, utils.GMaxPlayersKey),
		MaxPlayer:   int(utils.GetExtensionFloat64(matchProfile.Extensions, utils.GMaxPlayersKey)),
		MaxPing:     100000,
		MaxSkill:    int(utils.GetExtensionFloat64(matchProfile.Extensions, utils.GMaxSkillDifference)),
	}

	proposals := []*pb.Match{}
	if GCalculationMode == All {
		proposals, err = makeMatches(matchProfile, poolTickets, profileData.MaxPlayer)
		if err != nil {
			log.Printf("Failed to generate matches, got %s", err.Error())
			return err
		}
	} else {
		// Generate proposals.
		proposals, err = makeMatches2(poolTickets[utils.GPoolName], profileData)
		if err != nil {
			log.Printf("Failed to generate matches, got %s", err.Error())
			return err
		}
	}

	log.Printf("Streaming %v proposals to Open Match", len(proposals))
	// Stream the generated proposals back to Open Match.
	for _, proposal := range proposals {
		if err := stream.Send(&pb.RunResponse{Proposal: proposal}); err != nil {
			log.Printf("Failed to stream proposals to Open Match, got %s", err.Error())
			return err
		}
	}

	return nil
}

func makeMatches2(tickets []*pb.Ticket, profile ProfileData) ([]*pb.Match, error) {
	matches := []*pb.Match{}
	skillTickets := tickets

	sort.Slice(skillTickets, func(i, j int) bool {
		return ticket.GetSkillFromTicket(skillTickets[i]) < ticket.GetSkillFromTicket(skillTickets[j])
	})

	count := 0
	for ticketIndex := 0; ticketIndex+profile.MaxPlayer-1 < len(skillTickets); ticketIndex++ {
		mt := skillTickets[ticketIndex : ticketIndex+profile.MaxPlayer]
		if ticket.GetSkillFromTicket(mt[len(mt)-1])-ticket.GetSkillFromTicket(mt[0]) < float64(profile.MaxSkill) {

			avgLatency := 0.0
			for _, t := range mt {
				avgLatency += ticket.GetLatencyFromTicket(t, profile.Region, profile.MaxPing)
			}
			avgLatency /= float64(len(mt))

			qLatency := float64(0)
			for _, t := range mt {
				diff := ticket.GetLatencyFromTicket(t, profile.Region, profile.MaxPing) - avgLatency
				qLatency -= diff * diff
			}

			avgSkill := 0.0
			for _, t := range mt {
				avgSkill += ticket.GetSkillFromTicket(t)
			}
			avgSkill /= float64(len(mt))

			qSkill := 0.0
			for _, t := range mt {
				diff := ticket.GetSkillFromTicket(t) - avgSkill
				qSkill -= diff * diff
			}

			matches = append(matches, &pb.Match{
				MatchId:       fmt.Sprintf("profile-%v-time-%v-%v", profile.ProfileName, time.Now().Format("2006-01-02T15:04:05.00"), count),
				MatchProfile:  profile.ProfileName,
				MatchFunction: matchName,
				Tickets:       mt,
				Extensions: map[string]*anypb.Any{
					utils.GCurrentNumTickets: utils.GetAnyFromValue(float64(len(mt))),
					utils.GCurrentNumMatches: utils.GetAnyFromValue(float64(count)),
				},
			})
			count++
		}
	}

	// loop through and assign matches
	return matches, nil
}

func makeMatches(p *pb.MatchProfile, poolTickets map[string][]*pb.Ticket, matchPerProfile int) ([]*pb.Match, error) {
	var matches []*pb.Match
	count := 0
	for {
		insufficientTickets := false
		desiredRegionskillTicketsskillTickets := []*pb.Ticket{}
		log.Printf("Match profile data Num Players Per Match %d %+v", matchPerProfile, p)
		for pool, tickets := range poolTickets {
			if len(tickets) < matchPerProfile {
				// This pool is completely drained out. Stop creating matches.
				insufficientTickets = true
				break
			}

			// Remove the Tickets from this pool and add to the match proposal.
			desiredRegionskillTicketsskillTickets = append(desiredRegionskillTicketsskillTickets, tickets[0:matchPerProfile]...)
			poolTickets[pool] = tickets[matchPerProfile:]
		}

		if insufficientTickets {
			break
		}

		matches = append(matches, &pb.Match{
			MatchId:       fmt.Sprintf("profile-%v-time-%v-%v", p.GetName(), time.Now().Format("2006-01-02T15:04:05.00"), count),
			MatchProfile:  p.GetName(),
			MatchFunction: matchName,
			Tickets:       desiredRegionskillTicketsskillTickets,
			Extensions: map[string]*anypb.Any{
				utils.GCurrentNumTickets: utils.GetAnyFromValue(float64(len(desiredRegionskillTicketsskillTickets))),
				utils.GCurrentNumMatches: utils.GetAnyFromValue(float64(matchPerProfile)),
			},
		})

		count++
	}

	return matches, nil
}
