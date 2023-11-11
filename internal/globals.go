package utils

type SimulationMode int

const (
	All = iota
	OnlySkill
)

var (
	GTrustedNameTrue  = "trusted_true"
	GTrustedNameFalse = "trusted_false"

	GameModes     = []string{"bank_it", "quick_cash", "tournament_unranked", "tournament_ranked"}
	GRegions      = []string{"europe", "us"}
	GPasswordName = "password"

	GPoolName           = "all"
	GSkillArg           = "skill"
	GTrustedArg         = "trusted"
	GPasswordArg        = "password"
	GLatencyArg         = "latency"
	GBeginnerName       = "beginner"
	GMaxPlayersKey      = "max_players"
	GBestRegionKey      = "best_region"
	GProfileRegion      = "profile_region"
	GMaxSkillDifference = "match_skill"
	GSimulationMode     = All

	GMaxLatency = 1000
	GMaxSkill   = 50

	GCurrentNumTickets = "num_tickets"
	GCurrentNumMatches = "num_matches"
)
