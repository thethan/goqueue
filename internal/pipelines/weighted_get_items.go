package pipelines

import "sync/atomic"

type WwightedGetItems struct {
	counts map[string]atomic.Int64
}

func NewWeightedGetItems() {

}
