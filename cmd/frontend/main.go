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

// The Frontend in this tutorial continuously creates Tickets in batches in Open Match.

import (
	"context"
	"log"

	"sim/internal/ticket"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

const (
	// The endpoint for the Open Match Frontend service.
	omFrontendEndpoint = "open-match-frontend.open-match.svc.cluster.local:50504"
	// Number of tickets created per iteration
	ticketsGeneratedPerIteration = 20
)

func main() {
	// Connect to Open Match Frontend.
	conn, err := grpc.Dial(omFrontendEndpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Open Match, got %v", err)
	}

	defer conn.Close()
	fe := pb.NewFrontendServiceClient(conn)

	log.Printf("Amount of tickets attempted to be generated %d", ticketsGeneratedPerIteration)

	for {
		clientData := ticket.CreateRandomMatchmakingData()
		req := &pb.CreateTicketRequest{
			Ticket: ticket.MakeTicket(clientData),
		}
		resp, err := fe.CreateTicket(context.Background(), req)
		if err != nil {
			log.Printf("Failed to Create Ticket, got %s for client %+v", err.Error(), clientData)
			continue
		}

		log.Printf("Created ticket %s with client %+v", resp.GetId(), clientData)
	}
}
