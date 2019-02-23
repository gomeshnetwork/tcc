package notifier

import "testing"

func TestShuffle(t *testing.T) {
	m := []string{
		"1", "2", "3", "4", "5", "6", "7",
	}

	for i := 0; i < 10; i++ {

		m1 := shuffle(m)

		for k, v := range m1 {
			print(k, v, ",")
		}

		println("")
	}

}
