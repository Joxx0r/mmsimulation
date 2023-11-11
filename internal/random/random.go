package random

import (
	"math/rand"

	utils "sim/internal"
)

func FindGameMode() string {
	return utils.GameModes[rand.Intn(len(utils.GameModes))]
}

func FindTrustedState() string {
	if utils.GSimulationMode == utils.All {
		modes := []string{utils.GTrustedNameFalse, utils.GTrustedNameTrue}
		return modes[rand.Intn(len(modes))]
	} else {
		return utils.GTrustedNameTrue
	}
}

func FindPassword() string {
	if utils.GSimulationMode == utils.All {
		randomSeed := rand.Float64()
		if randomSeed > 0.8 {
			return utils.GPasswordName
		}
	}
	return ""
}

func FindSkill() float64 {
	if utils.GSimulationMode == utils.All || utils.GSimulationMode == utils.OnlySkill {
		return rand.Float64() * 500
	}
	return 0
}

func FindRegionRandom(region string) float64 {
	if utils.GSimulationMode == utils.All {
		return rand.Float64() * 500
	}
	return 0
}
