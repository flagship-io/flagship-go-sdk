package bucketing

import (
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/rs/xid"
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
		ID:         "vgID1",
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 33})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 34})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
		ID:         "vgID2",
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 25})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 35})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 30})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
		ID:         "vgID3",
	}
	testVariationGroupAlloc(variationsGroupInfo, t)

	variationArray = []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 10})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 25})

	variationsGroupInfo = VariationGroup{
		Variations: variationArray,
		ID:         "vgID4",
	}
	testVariationGroupAlloc(variationsGroupInfo, t)
}

// Test that same visitors got their variation assignment changed between multiple variation groups
func TestVisitorVGVariationChange(t *testing.T) {
	variationArray := []*Variation{}
	variationArray = append(variationArray, &Variation{ID: "1", Allocation: 20})
	variationArray = append(variationArray, &Variation{ID: "2", Allocation: 40})
	variationArray = append(variationArray, &Variation{ID: "3", Allocation: 60})
	variationArray = append(variationArray, &Variation{ID: "4", Allocation: 80})
	variationArray = append(variationArray, &Variation{ID: "5", Allocation: 100})

	vg1 := VariationGroup{
		ID:         xid.New().String(),
		Variations: variationArray,
	}

	vg2 := VariationGroup{
		ID:         xid.New().String(),
		Variations: variationArray,
	}

	s2 := rand.NewSource(time.Now().Unix())
	r2 := rand.New(s2)
	visitorIDs := []string{}
	for i := 0; i < 1000; i++ {
		visitorIDs = append(visitorIDs, strconv.Itoa(r2.Int()))
	}

	var counts map[string][]int = make(map[string][]int)
	for i := 1; i < 3; i++ {
		vg := vg1
		if i > 1 {
			vg = vg2
		}
		for _, visID := range visitorIDs {
			v, _ := GetRandomAllocation(visID, &vg)
			_, exists := counts[visID]

			if !exists {
				counts[visID] = []int{0, 0, 0, 0, 0}
			}
			vIDint, _ := strconv.Atoi(v.ID)
			counts[visID][vIDint-1]++
		}
	}

	visChanged := 0
	for _, count := range counts {
		for _, c := range count {
			if c == 1 {
				visChanged++
				break
			}
		}
	}

	t.Logf("Count %d visitor that changed var between vg", visChanged)
	if visChanged < 100 {
		t.Errorf("Expecting at least 10 vis that had different var between vg. Got %d", visChanged)
	}
}
