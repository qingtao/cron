package cron

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Time struct {
	Second []int `json:"second"`
	Minute []int `json:"minute"`
	Hour   []int `json:"hour"`
	Dom    []int `json:"dom"`
	Month  []int `json:"month"`
	Dow    []int `json:"dow"`
}

const (
	asterisk   = "*"
	hyphen     = "-"
	slash      = "/"
	comma      = ","
	lengthTime = 6
)

var slashOption = []string{"2", "3", "4", "5", "6", "10", "12", "15", "20", "30"}

type TimeOption struct {
	Min, Max int
}

var timeOption = map[string]TimeOption{
	"second": TimeOption{0, 59},
	"minute": TimeOption{0, 59},
	"hour":   TimeOption{0, 23},
	"dom":    TimeOption{1, 31},
	"month":  TimeOption{1, 12},
	"dow":    TimeOption{0, 6},
}

var (
	ErrNoSlash     = errors.New("no slash")
	ErrNoHyphen    = errors.New("no hyphen")
	ErrTimeInvalid = errors.New("invalid times")
	ErrField       = errors.New("field not enough")
)

func checkSlash(s string) bool {
	for i := 0; i < len(slashOption); i++ {
		if slashOption[i] == s {
			return true
		}
	}
	return false
}

func splitSlash(s, typ string) ([]int, error) {
	ss := strings.Split(s, slash)
	if len(ss) != 2 {
		return nil, ErrNoSlash
	}
	if !checkSlash(ss[1]) {
		return nil, errors.New(fmt.Sprintf(`"%s": the number after %s must be one of: %s`, typ, slash, slashOption))
	}

	op := timeOption[typ]
	min, max := op.Min, op.Max
	s1, err := strconv.Atoi(ss[0])
	if err != nil {
		return nil, err
	}
	s2, err := strconv.Atoi(ss[1])
	if err != nil {
		return nil, err
	}
	if s1 < min || s1 > max {
		return nil, errors.New(fmt.Sprintf(`"%s": use %s: %d-%d`, typ, slash, min, max))
	}

	res := make([]int, 0)
	for i := s1; i < max; i += s2 {
		res = append(res, i)
	}
	return res, nil
}

func Add(s []int, a interface{}) []int {
	switch a.(type) {
	case int:
		v, _ := a.(int)
		if len(s) > 0 {
			for i := 0; i < len(s); i++ {
				if v == s[i] {
					return s
				}
			}
		}
		s = append(s, v)
	case []int:
		v, _ := a.([]int)
		if len(s) < 1 {
			s = append(s, v...)
		} else {
		TOP:
			for i := 0; i < len(v); i++ {
				for j := 0; j < len(s); j++ {
					if v[i] == s[j] {
						continue TOP
					}
				}
				s = append(s, v[i])
			}
		}
	}
	return s
}

func splitHyphen(s, typ string) ([]int, error) {
	ss := strings.Split(s, hyphen)
	if len(ss) != 2 {
		return nil, ErrNoHyphen
	}
	op := timeOption[typ]
	min, max := op.Min, op.Max
	s1, err := strconv.Atoi(ss[0])
	if err != nil {
		return nil, err
	}
	s2, err := strconv.Atoi(ss[1])
	if err != nil {
		return nil, err
	}
	if s1 < min || s1 > max {
		return nil, errors.New(fmt.Sprintf(`"%s": the number before %s can not less than %d or more than %d`, typ, hyphen, min, max))
	}
	if s2 < s1 || s2 > max {
		return nil, errors.New(fmt.Sprintf(`"%s": the number after %s can not less than %d or more than %d`, typ, hyphen, s1, max))
	}

	res := make([]int, 0)
	for i := s1; i <= s2; i++ {
		res = append(res, i)
	}
	return res, nil
}

func splitComma(s, typ string) ([]int, error) {
	times := make([]int, 0)
	if s == "*" {
		op := timeOption[typ]
		min, max := op.Min, op.Max
		for i := min; i <= max; i++ {
			times = append(times, i)
		}
	} else {
		ss := strings.Split(s, comma)
		for _, v := range ss {
			n := strings.Index(v, slash)
			if n > 0 {
				t, err := splitSlash(v, typ)
				if err != nil {
					if err.Error() == ErrNoSlash.Error() {
						continue
					}
					return nil, err
				}
				times = append(times, t...)
			} else {
				n = strings.Index(v, hyphen)
				if n > 0 {
					t, err := splitHyphen(v, typ)
					if err != nil {
						if err.Error() == ErrNoHyphen.Error() {
							continue
						}
						return nil, err
					}
					times = Add(times, t)
				} else if len(ss) == 1 {
					i, err := strconv.Atoi(s)
					if err != nil {
						return nil, err
					}
					times = append(times, i)
				}

			}
		}
	}
	if len(times) < 1 {
		return nil, ErrTimeInvalid
	}
	return times, nil

}

func Parse(s string) (*Time, error) {
	ss := strings.Fields(s)
	if len(ss) != 6 {
		return nil, ErrField
	}
	second, err := splitComma(ss[0], "second")
	if err != nil {
		return nil, err
	}
	minute, err := splitComma(ss[1], "minute")
	if err != nil {
		return nil, err
	}
	hour, err := splitComma(ss[2], "hour")
	if err != nil {
		return nil, err
	}
	dom, err := splitComma(ss[3], "dom")
	if err != nil {
		return nil, err
	}
	month, err := splitComma(ss[4], "month")
	if err != nil {
		return nil, err
	}
	dow, err := splitComma(ss[5], "dow")
	if err != nil {
		return nil, err
	}
	return &Time{second, minute, hour, dom, month, dow}, nil
}

func checkInt(a []int, b int) bool {
	for i := 0; i < len(a); i++ {
		if a[i] == b {
			return true
		}
	}
	return false
}

func (t *Time) Check(u time.Time) bool {
	_, m, d := u.Date()
	H, M, S := u.Clock()
	D := u.Weekday()
	if checkInt(t.Second, S) {
		if checkInt(t.Minute, M) {
			if checkInt(t.Hour, H) {
				if checkInt(t.Dom, d) {
					if checkInt(t.Dow, int(D)) {
						if checkInt(t.Month, int(m)) {
							if time.Now().Sub(u) < 500*time.Millisecond {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}
