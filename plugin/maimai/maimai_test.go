package maimai

import (
	"fmt"
	"testing"
)

func TestFetchScore(t *testing.T) {
	fmt.Println(GetMinMaxChart("tamce", ""))
	fmt.Println(GetMinMaxChart("", "876472013"))
}

func TestRa(t *testing.T) {
	fmt.Println(CalcRa(12.7, 100.1))
	fmt.Println(CalcRa(13.3, 99.8))
	fmt.Println(FindDifficultyForRaAcc(175, 99.5))
}
