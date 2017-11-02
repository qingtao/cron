package cron

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

var years = []int{2016, 2017, 2018}

func TestLastDay(t *testing.T) {
	for _, y := range years {
		t.Logf("year: %d\n", y)
		for i := 1; i <= 12; i++ {
			last := lastDay(i, y)
			t.Logf("year: %d, month: %d, last day: %d\n", y, i, last)
		}
		t.Log("---------------\n")
	}
}

func TestCheckSlash(t *testing.T) {
	s := []string{"12", "1", "11", "*", "30"}
	for k, v := range s {
		t.Logf("%d: %4s: %v\n", k, v, checkSlash(v))
	}
}

func TestSplitSlash(t *testing.T) {
	s := []string{
		"1/2", "1/20", "3/12", "4/13", "6/33", "30/2", "5/14",
	}
	for k, v := range s {
		a, err := splitSlash(v, "minute")
		if err != nil {
			t.Logf("fail %d: %4s: %s\n", k, v, err)
		} else {
			t.Logf("ok %d: %4s: %v\n", k, v, a)
		}
	}
}

type Data struct {
	A interface{}
	S []int
}

func TestAdd(t *testing.T) {
	data := []*Data{
		{A: 2, S: []int{1, 3, 5, 7, 20}},
		{A: 22, S: []int{1, 2, 3, 6, 7, 22}},
		{A: []int{2, 22, 33, 40}, S: []int{1, 2, 3, 6, 7, 22}},
		{A: []int{1, 2, 6, 7, 8, 10, 22, 9, 20}, S: []int{3, 5, 9, 22}},
	}

	t.Log("before add action:")
	for k, v := range data {
		t.Logf("%d s: %#v, a: %#v\n", k, v.S, v.A)
	}
	t.Log("after add action:")
	for k, v := range data {
		s := Add(v.S, v.A)
		sort.Slice(s, func(i, j int) bool {
			return s[i] < s[j]
		})
		t.Logf("%d s: %#v\n", k, s)
	}
}

func TestSplit(t *testing.T) {
	var s = []string{
		"1-10",
		"0-6",
		"3-33",
		"33-3",
		"33-50",
		"1-12",
		"3-13",
		"40-60",
		"5-30",
		"5-31",
		"1-23",
		"0-23",
		"1-6",
	}
	for name := range timeOptions {
		t.Run(name, func(t *testing.T) {
			for k, v := range s {
				a, err := splitHyphen(v, name)
				if err != nil {
					t.Logf("fail %d: %4s: %s\n", k, v, err)
				} else {
					t.Logf("ok %d: %4s: %v\n", k, v, a)
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	s := []string{
		// second minute hour dom month dow
		"1/3 * * * * *",
		"* 1-3 * * * *",
		"0 1-10,1/3 * * * *",
		"* * 3 * * *",
		"* * * * * 6",
		"0 * * * 5 *",
		"0 * * 20 * *",
		Monthly,
		Weekly,
		Daily,
		Hourly,
		Minute,
		Second,
		"0 0 0 * 10 * L",
		"0 0 0 * 9 * A",
		"0 0 0 1 8 * L",
		"0 0 0 * 7 1 L",
		"0 0 0 1 7 6 L",
	}
	for k, v := range s {
		a, err := Parse(v)
		if err != nil {
			t.Logf("fail %d: %4s: %s\n", k, v, err)
		} else {
			t.Logf("ok %d: %4s: %v\n", k, v, a)
		}
	}
}

func TestCheck(t *testing.T) {
	a := []*Time{
		{[]int{0}, []int{0}, []int{0}, nil, []int{1}, nil, true},
		{[]int{0}, []int{0}, []int{0}, nil, []int{2}, nil, true},
		{[]int{0}, []int{0}, []int{0}, nil, []int{4}, nil, true},
		{[]int{0}, []int{0}, []int{0}, []int{29}, []int{1, 3, 4}, []int{0, 1, 2, 3, 4, 5, 6}, false},
	}
	u := make([]time.Time, 0)
	loc := time.Now().Location()
	for j := 2016; j < 2018; j++ {
		for i := 1; i < 5; i++ {
			for n := 28; n < 32; n++ {
				u = append(u, time.Date(j, time.Month(i), n, 0, 0, 0, 0, loc))
			}
		}
	}
	for k, v := range a {
		fmt.Printf("%2d - %v\n", k, v)
		for kk, vv := range u {
			fmt.Printf("%2d - %d-%d-%d %v\n", kk, vv.Year(), vv.Month(), vv.Day(), v.Check(vv))
		}
	}

}
