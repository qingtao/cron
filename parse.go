package cron

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Time存放任务的时间
// 格式：
// second minute hour dom         month  dow
// 秒     分     时   每月的一天  月    周几
// second:	"0-59", ",", "-", "/", "*"
// minute:	"0-59", ",", "-", "/", "*"
// hour:	"0-23", ",", "-", "/", "*"
// dom:		"1-31", ",", "-", "/", "*"
// month:	"1-12", ",", "-", "/", "*"
// second:	"0-6",  ",", "-", "/", "*"
//
// 注意：dom和dow都不是*时，时间是两者交集
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

// 定义'/'递增的数值
var slashOption = []string{"2", "3", "4", "5", "6", "10", "12", "15", "20", "30"}

// 时间的最小和最大值
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

// 分割'/'获取int类型的时间切片
func splitSlash(s, typ string) ([]int, error) {
	ss := strings.Split(s, slash)
	//如果不是1/2这种格式，返回错误
	if len(ss) != 2 {
		return nil, ErrNoSlash
	}
	//检查间隔是否有效
	if !checkSlash(ss[1]) {
		return nil, errors.New(fmt.Sprintf(`"%s": the number after %s must be one of: %s`, typ, slash, slashOption))
	}

	op := timeOption[typ]
	min, max := op.Min, op.Max
	//转换字符串到整型数值
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

// 新增时间数值a到现有切片s，a可能是[]int，或者是int
func Add(s []int, a interface{}) []int {
	switch a.(type) {
	case int:
		v, _ := a.(int)
		if len(s) > 0 {
			//检查a是否已存在
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
			//检查a中的每个值是否s已含有
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

//分割-
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

// 获取以","组成的数值
func splitComma(s, typ string) ([]int, error) {
	times := make([]int, 0)
	//s是"*",返回所有有效值
	if s == "*" {
		op := timeOption[typ]
		min, max := op.Min, op.Max
		for i := min; i <= max; i++ {
			times = append(times, i)
		}
	} else {
		//首先以","分割字符串
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
					//单个数字的情况
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

// 解析字符串s到*Time，失败返回错误
func Parse(s string) (*Time, error) {
	ss := strings.Fields(s)
	if len(ss) != 6 {
		return nil, ErrField
	}
	//秒
	second, err := splitComma(ss[0], "second")
	if err != nil {
		return nil, err
	}
	//分钟
	minute, err := splitComma(ss[1], "minute")
	if err != nil {
		return nil, err
	}
	//小时
	hour, err := splitComma(ss[2], "hour")
	if err != nil {
		return nil, err
	}
	//每月第几天
	dom, err := splitComma(ss[3], "dom")
	if err != nil {
		return nil, err
	}
	//月份
	month, err := splitComma(ss[4], "month")
	if err != nil {
		return nil, err
	}
	//每周几
	dow, err := splitComma(ss[5], "dow")
	if err != nil {
		return nil, err
	}
	return &Time{second, minute, hour, dom, month, dow}, nil
}

// 检查a是否存在b
func checkInt(a []int, b int) bool {
	for i := 0; i < len(a); i++ {
		if a[i] == b {
			return true
		}
	}
	return false
}

// 检查时间u是否符合时间t的定义
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
