package bucketing

import (
	"fmt"

	"github.com/twmb/murmur3"
)

// GetRandomAllocation returns a random allocation for a variationGroup
func GetRandomAllocation(visitorID string, variationGroup *VariationGroup) (*Variation, error) {
	hash := murmur3.New32()
	_, err := hash.Write([]byte(visitorID))

	if err != nil {
		return nil, err
	}

	hashed := hash.Sum32()
	z := hashed % 100

	summedAlloc := 0
	for _, v := range variationGroup.Variations {
		summedAlloc += v.Allocation
		if int(z) < summedAlloc {
			return v, nil
		}
	}

	// If no variation alloc, returns empty
	return nil, fmt.Errorf("Visitor untracked for vg ID : %s", variationGroup.ID)
}
