package normalizer

import (
	"fmt"
	"strconv"
)

type ExtraArg struct {
	Column string
	Value  interface{}
}

type ExtractedArgs struct {
	Query     string
	ExtraArgs []ExtraArg
}

func unwrapInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("failed to unwrap int: %w", err))
	}
	return i
}
