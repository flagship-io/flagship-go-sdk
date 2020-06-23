package bucketing

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func testVariationGroupAlloc(vg VariationGroup, t *testing.T) {
	counts := []int{}
	errs := []error{}
	countTotal := 100000

	for i := 1; i < countTotal; i++ {
		vAllocated, err := GetRandomAllocation(strconv.Itoa(rand.Int()), &vg)

		if err != nil {
			errs = append(errs, err)
			continue
		}

		for i, v := range vg.Variations {
			if v.ID == vAllocated.ID {
				for len(counts) <= i {
					counts = append(counts, 0)
				}
				counts[i]++
			}
		}
	}

	for i, v := range counts {
		t.Logf("Count v%d : %d", i+1, v)
	}

	for i, v := range counts {
		correctRatio := float64(vg.Variations[i].Allocation) / 100
		ratio := float64(v) / float64(countTotal)
		if math.Abs(correctRatio-ratio) > 0.05 {
			t.Errorf("Problem with stats: ratio %f, correctRatio : %f", ratio, correctRatio)
		}
	}

	sumAllocVars := 0
	for _, v := range vg.Variations {
		sumAllocVars += v.Allocation
	}
	correctErrorRatio := float64((100 - sumAllocVars)) / 100
	errorRatio := float64(len(errs)) / float64(countTotal)
	if math.Abs(correctErrorRatio-errorRatio) > 0.05 {
		t.Errorf("Problem with stats: error ratio %f, correctErrorRatio : %f", errorRatio, correctErrorRatio)
	}
}

func TestVariationAllocation(t *testing.T) {
	variationArray := []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 50})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 50})

	variationsGroupInfo := VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 34})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 25})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 35})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 30})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 25})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
	}
	testVariationGroupAlloc(variationsGroupInfo, t)
}
