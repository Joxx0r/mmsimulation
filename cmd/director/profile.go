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
	"fmt"

	utils "sim/internal"

	"google.golang.org/protobuf/types/known/anypb"
	"open-match.dev/open-match/pkg/pb"
)

type GameModeData struct {
	modeName           string
	skillBoundaries    []float64
	maxSkillDifference float64
	trustedQueues      bool
	playersPerGame     int
	skillDiffBand      int
	backfill           bool
	beginner           bool
}

// TeamShooterScenario provides the required methods for running a scenario.
type FinalsGameScenario struct {
	modeData        []GameModeData
	regions         []string
	activePasswords []string
	maxLatency      float64
}

const (
	poolName    = "all"
	skillArg    = "skill"
	modeArg     = "mode"
	trustedArg  = "trusted"
	passwordArg = "password"
	latencyArg  = "latency"
)

func profilesCall(t *FinalsGameScenario) []*pb.MatchProfile {
	p := []*pb.MatchProfile{}
	for _, region := range t.regions {
		for _, mode := range t.modeData {
			for i := 0; i+1 < len(mode.skillBoundaries); i++ {
				trustedNum := 1
				if mode.trustedQueues {
					trustedNum += 1
				}

				for trustedIndex := 0; trustedIndex < trustedNum; trustedIndex++ {
					beginnerNum := 1
					if mode.beginner {
						beginnerNum = 2
					}
					for beginnerIndex := 0; beginnerIndex < beginnerNum; beginnerIndex++ {
						skillMin := mode.skillBoundaries[i] - mode.maxSkillDifference/2
						skillMax := mode.skillBoundaries[i+1] + mode.maxSkillDifference/2

						trustedName := utils.GTrustedNameFalse
						if trustedIndex > 0 {
							trustedName = utils.GTrustedNameTrue
						}

						matchProfile := &pb.MatchProfile{
							Name: fmt.Sprintf("%s_%s", region, mode),
							Pools: []*pb.Pool{
								{
									Name: poolName,
									DoubleRangeFilters: []*pb.DoubleRangeFilter{
										{
											DoubleArg: skillArg,
											Min:       skillMin,
											Max:       skillMax,
										},
									},
									TagPresentFilters: []*pb.TagPresentFilter{
										{
											Tag: region,
										},
										{
											Tag: mode.modeName,
										},
										{
											Tag: trustedName,
										},
									},
								},
							},
							Extensions: make(map[string]*anypb.Any),
						}
						utils.AddExtensionFloat64(matchProfile.Extensions, utils.GMaxPlayersKey, float64(mode.playersPerGame))
						utils.AddExtensionFloat64(matchProfile.Extensions, utils.GMaxSkillDifference, float64(mode.skillDiffBand))
						utils.AddExtensionString(matchProfile.Extensions, utils.GProfileRegion, region)
						if beginnerIndex > 0 {
							filter := []*pb.TagPresentFilter{
								{
									Tag: utils.GBeginnerName,
								},
							}
							matchProfile.Pools[0].TagPresentFilters = append(matchProfile.Pools[0].TagPresentFilters, filter...)
						}

						p = append(p, matchProfile)
					}
				}
			}
		}
	}

	if t.activePasswords != nil {
		for _, password := range t.activePasswords {
			p = append(p, &pb.MatchProfile{
				Name: fmt.Sprintf("password_%s", password),
				Pools: []*pb.Pool{
					{
						Name: poolName,
						TagPresentFilters: []*pb.TagPresentFilter{
							{
								Tag: password,
							},
						},
					},
				},
			})
		}
	}

	return p
}
