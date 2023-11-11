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

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	utils "sim/internal"
	grpccontext "sim/internal/grpc"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"app":       "openmatch",
	"component": "scale.mmf",
})

// The Director in this tutorial continously polls Open Match for the Match
// Profiles and makes random assignments for the Tickets in the returned matches.

const (
	// The endpoint for the Open Match Backend service.
	omBackendEndpoint = "open-match-backend.open-match.svc.cluster.local:50505"
	// The Host and Port for the Match Function service endpoint.
	functionHostName = "matchfunction.mm.svc.cluster.local"
	functionPort     = 50502
	// The endpoint for the Open Match Frontend service.
	omFrontendEndpoint = "open-match-frontend.open-match.svc.cluster.local:50504"
)

func main() {
	log.Printf("Starting Director")

	// Connect to Open Match Backend.
	conn, err := grpc.Dial(omBackendEndpoint, grpccontext.NewGRPCDialOptions(logger)...)
	if err != nil {
		log.Fatalf("Failed to connect to Open Match Backend, got %s", err.Error())
	}

	log.Printf("Endpoint being used %s", omBackendEndpoint)
	defer conn.Close()
	be := pb.NewBackendServiceClient(conn)

	// Connect to Open Match Backend.
	conn2, err := grpc.Dial(omFrontendEndpoint, grpccontext.NewGRPCDialOptions(logger)...)
	if err != nil {
		log.Fatalf("Failed to connect to Open Match Frontend, got %s", err.Error())
	}

	defer conn2.Close()
	fe := pb.NewFrontendServiceClient(conn2)

	scenario := &FinalsGameScenario{
		modeData:        []GameModeData{},
		regions:         utils.GRegions,
		activePasswords: []string{utils.GPasswordArg},
	}

	for _, mode := range utils.GameModes {
		modeData := GameModeData{
			modeName:           mode,
			skillBoundaries:    []float64{0, 500, 1500},
			maxSkillDifference: float64(utils.GMaxSkill),
			trustedQueues:      true,
			playersPerGame:     16,
			skillDiffBand:      50,
			backfill:           true,
		}
		scenario.modeData = append(scenario.modeData, modeData)
	}

	profiles := profilesCall(scenario)
	log.Printf("Fetching matches for %v profiles", len(profiles))

	for range time.Tick(time.Second * 5) {
		// Fetch matches for each profile and make random assignments for Tickets in
		// the matches returned.
		var wg sync.WaitGroup
		for _, p := range profiles {
			wg.Add(1)
			go func(wg *sync.WaitGroup, p *pb.MatchProfile) {
				defer wg.Done()
				matches, err := fetch(be, p)
				if err != nil {
					log.Printf("Failed to fetch matches for profile %v, got %s", p.GetName(), err.Error())
					return
				}

				count := 0
				for _, match := range matches {
					count += len(match.GetTickets())
				}

				if count > 0 {
					log.Printf("Generated %d matches for profile %s amount of tickets %d", len(matches), p.Name, count)
				}
				if err := assign(be, matches, p, fe); err != nil {
					log.Printf("Failed to assign servers to matches, got %s", err.Error())
					return
				}
			}(&wg, p)
		}

		wg.Wait()
	}
}

func fetch(be pb.BackendServiceClient, p *pb.MatchProfile) ([]*pb.Match, error) {
	req := &pb.FetchMatchesRequest{
		Config: &pb.FunctionConfig{
			Host: functionHostName,
			Port: functionPort,
			Type: pb.FunctionConfig_GRPC,
		},
		Profile: p,
	}

	startTime := time.Now()
	stream, err := be.FetchMatches(context.Background(), req)
	if err != nil {
		log.Println()
		return nil, err
	}

	var result []*pb.Match
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		result = append(result, resp.GetMatch())
	}
	endTime := time.Now()
	difference := endTime.Sub(startTime)
	log.Printf("Time to fetch took %s for profile %s", difference.String(), p.GetName())

	return result, nil
}

func assign(be pb.BackendServiceClient, matches []*pb.Match, matchProfile *pb.MatchProfile, fe pb.FrontendServiceClient) error {
	for _, match := range matches {
		ticketIDs := []string{}
		for _, t := range match.GetTickets() {
			ticketIDs = append(ticketIDs, t.Id)
		}

		conn := fmt.Sprintf("%d.%d.%d.%d:2222", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
		req := &pb.AssignTicketsRequest{
			Assignments: []*pb.AssignmentGroup{
				{
					TicketIds: ticketIDs,
					Assignment: &pb.Assignment{
						Connection: conn,
					},
				},
			},
		}

		if _, err := be.AssignTickets(context.Background(), req); err != nil {
			return fmt.Errorf("AssignTickets failed for match %v, got %w", match.GetMatchId(), err)
		}
	}

	return nil
}
