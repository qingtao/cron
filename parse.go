package cron

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//Time 存放任务的时间
//格式:
//	second	minute	hour	dom		month	dow
//	秒	分	时	每月的第一天	月	每周的第几天
//	second:	0-59 , - / * 秒
//	minute:	0-59 , - / * 分
//	hour:	0-23 , - / * 时
//	dom:	1-31 , - / * 每月第几天
//	month:	1-12 , - / * 月
//	dow:	0-6  , - / * 每周第几天
//	LastDayOfMonth: 指定每月中的最后一天
//
//
//	例:
//	每天6点:
//		s = "0 0 6 * * *"
//	每个月最后一天:
//		s = "0 0 0 * * * L"
// 注意: dom和dow都不是*时，时间是两者交集，指定LastDayOfMonth，Dom和Dow必须同时是"*"
type Time struct {
	Second []int `json:"second"`
	Minute []int `json:"minute"`
	Hour   []int `json:"hour"`
	Dom    []int `json:"dom"`
	Month  []int `json:"month"`
	Dow    []int `json:"dow"`
	//指定LastDayOfMonth时忽略Dom和Dow
	LastDayOfMonth bool
}

const (
	asterisk   = "*"
	hyphen     = "-"
	slash      = "/"
	comma      = ","
	lengthTime = 6
)

//slashOption 定义'/'递增的数值
var slashOption = []string{"2", "3", "4", "5", "6", "10", "12", "15", "20", "30"}

//timeOption 时间的最小和最大值
type timeOption struct {
	Min, Max int
}

var timeOptions = map[string]timeOption{
	"second": {0, 59},
	"minute": {0, 59},
	"hour":   {0, 23},
	"dom":    {1, 31},
	"month":  {1, 12},
	"dow":    {0, 6},
}

//LastDayOfMonth 是Time的第七个字段
const LastDayOfMonth = "L"

var (
	//ErrNoSlash slash不存在
	ErrNoSlash = errors.New("no slash")
	//ErrNoHyphen hyphen不存在
	ErrNoHyphen = errors.New("no hyphen")
	//ErrTimeInvalid time格式错误
	ErrTimeInvalid = errors.New("invalid times")
	//ErrField 字符串字段不足
	ErrField = errors.New("field not enough")
	//ErrLastDayOfMonth 指定LastDayOfMonth的格式错误
	ErrLastDayOfMonth = errors.New(`LastDayOfMonth only accept "L", "dom" and "dow" must be "*"`)
)

func checkSlash(s string) bool {
	for i := 0; i < len(slashOption); i++ {
		if slashOption[i] == s {
			return true
		}
	}
	return false
}

//分割'/'获取int类型的时间切片
func splitSlash(s, typ string) ([]int, error) {
	ss := strings.Split(s, slash)
	//如果不是1/2这种格式，返回错误
	if len(ss) != 2 {
		return nil, ErrNoSlash
	}
	//检查间隔是否有效
	if !checkSlash(ss[1]) {
		return nil, fmt.Errorf(`"%s": the number after %s must be one of: %s`, typ, slash, slashOption)
	}

	op := timeOptions[typ]
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
		return nil, fmt.Errorf(`"%s": use %s: %d-%d`, typ, slash, min, max)
	}

	res := make([]int, 0)
	for i := s1; i < max; i += s2 {
		res = append(res, i)
	}
	return res, nil
}

//Add 新增时间数值a到现有切片s，a可能是[]int，或者是int
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

//分割"-"
func splitHyphen(s, typ string) ([]int, error) {
	ss := strings.Split(s, hyphen)
	if len(ss) != 2 {
		return nil, ErrNoHyphen
	}
	op := timeOptions[typ]
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
		return nil, fmt.Errorf(`"%s": the number before %s can not less than %d or more than %d`, typ, hyphen, min, max)
	}
	if s2 < s1 || s2 > max {
		return nil, fmt.Errorf(`"%s": the number after %s can not less than %d or more than %d`, typ, hyphen, s1, max)
	}

	res := make([]int, 0)
	for i := s1; i <= s2; i++ {
		res = append(res, i)
	}
	return res, nil
}

//获取以","组成的数值
func splitComma(s, typ string) ([]int, error) {
	times := make([]int, 0)
	//s是"*",返回所有有效值
	if s == "*" {
		op := timeOptions[typ]
		min, max := op.Min, op.Max
		for i := min; i <= max; i++ {
			times = append(times, i)
		}
	} else {
		//首先以","分割字符串
		ss := strings.Split(s, comma)
		for _, v := range ss {
			//分割"/"
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
				//分割"-"
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
					//单个数字的情况,或者以","分割的单个数据
				} else if len(ss) > 0 {
					m, err := strconv.Atoi(v)
					if err != nil {
						return nil, err
					}
					times = append(times, m)
				}

			}
		}
	}
	if len(times) < 1 {
		return nil, ErrTimeInvalid
	}
	return times, nil
}

//Parse 解析字符串s到*Time，失败返回错误
func Parse(s string) (*Time, error) {
	ss := strings.Fields(s)
	lastDom := false
	if len(ss) != 6 {
		if len(ss) == 7 {
			if ss[6] == LastDayOfMonth && ss[3] == "*" && ss[5] == "*" {
				lastDom = true
				ss = ss[0:6]
			} else {
				return nil, ErrLastDayOfMonth
			}
		} else {
			return nil, ErrField
		}
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
	//月份
	month, err := splitComma(ss[4], "month")
	if err != nil {
		return nil, err
	}
	if !lastDom {
		//每月第几天
		dom, err := splitComma(ss[3], "dom")
		if err != nil {
			return nil, err
		}
		//每周几
		dow, err := splitComma(ss[5], "dow")
		if err != nil {
			return nil, err
		}
		return &Time{second, minute, hour, dom, month, dow, lastDom}, nil
	}
	return &Time{second, minute, hour, nil, month, nil, lastDom}, nil
}

//checkInt 检查a是否存在b
func checkInt(a []int, b int) bool {
	for i := 0; i < len(a); i++ {
		if a[i] == b {
			return true
		}
	}
	return false
}

//lastday get last day of month, m is the month, y is year
func lastDay(m, y int) int {
	switch m {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	}
	if m > 12 || m < 1 {
		return -1
	}
	//the last day of February
	loc := time.Now().Location()
	last := time.Date(y, time.March, 1, 0, 0, 0, 0, loc).AddDate(0, 0, -1)
	return last.Day()
}

//Check 检查时间u是否符合时间t的定义
func (t *Time) Check(u time.Time) bool {
	if time.Now().Sub(u) > 500*time.Millisecond {
		return false
	}
	y, m, d := u.Date()
	H, M, S := u.Clock()
	D := u.Weekday()
	if checkInt(t.Second, S) {
		if checkInt(t.Minute, M) {
			if checkInt(t.Hour, H) {
				if checkInt(t.Month, int(m)) {
					//如果指定最后一天，否则检查Dom和Dow
					if t.LastDayOfMonth {
						if d == lastDay(int(m), y) {
							return true
						}
					} else {
						if checkInt(t.Dom, d) {
							if checkInt(t.Dow, int(D)) {
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
