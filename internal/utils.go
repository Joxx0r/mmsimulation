package utils

import (
	"math"

	simproto "sim/proto"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"open-match.dev/open-match/pkg/pb"
)

func GetAnyFromValue(value float64) *anypb.Any {
	protoMessage := &pb.DefaultEvaluationCriteria{Score: float64(value)}
	return MustAny(protoMessage)
}

func MustAny(m proto.Message) *anypb.Any {
	result, err := anypb.New(m)
	if err != nil {
		panic(err)
	}
	return result
}

func AddExtensionFloat64(Extensions map[string]*anypb.Any, key string, value float64) {
	Extensions[key] = MustAny(&pb.DefaultEvaluationCriteria{Score: float64(value)})
}

func AddExtensionString(Extensions map[string]*anypb.Any, key string, stringValue string) {
	Extensions[key] = MustAny(&simproto.DefaultEvaluationString{Value: stringValue})
}

func GetExtensionFloat64(Extensions map[string]*anypb.Any, key string) float64 {
	if a, ok := Extensions[key]; ok {
		imp := &pb.DefaultEvaluationCriteria{
			Score: math.Inf(-1),
		}
		err := a.UnmarshalTo(imp)
		if err != nil {
			return math.Inf(-1)
		}
		return imp.Score
	}
	return math.Inf(-1)
}

func GetExtensionString(Extensions map[string]*anypb.Any, key string) string {
	if a, ok := Extensions[key]; ok {
		imp := &simproto.DefaultEvaluationString{
			Value: "not_assigned",
		}
		err := a.UnmarshalTo(imp)
		if err != nil {
			return "invalid_parse"
		}
		return imp.Value
	}
	return "not_assigned"
}
