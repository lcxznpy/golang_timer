package cron

import (
	"fmt"
	"time"

	"github.com/gorhill/cronexpr"
)

// cron表达式解析器 结构
type CronParser struct {
}

// cron解析器构造函数
func NewCronParser() *CronParser {
	return &CronParser{}
}

// 判断cron表达式合不合法
func (c *CronParser) IsValidCronExpr(cron string) bool {
	_, err := cronexpr.Parse(cron)
	return err == nil
}

// 获取距离当前时间最近且匹配cron表达式的的一个时间点
func (c *CronParser) NextFromNow(cron string) (time.Time, error) {
	expr, err := cronexpr.Parse(cron)
	if err != nil {
		return time.Time{}, err
	}

	nextTime := expr.Next(time.Now())
	if nextTime.UnixNano() < 0 {
		return time.Time{}, fmt.Errorf("fail to parse time from cron: %s", cron)
	}
	return nextTime, nil
}

// NextBefore 返回在 time.now 和 end 时间区间内符合cron表达式的时间点切片
func (c *CronParser) NextsBefore(cron string, end time.Time) ([]time.Time, error) {
	return c.NextsBetween(cron, time.Now(), end)
}

// NextsBetween 返回在start 和 end 时间区间内符合cron表达式的时间点切片
func (c *CronParser) NextsBetween(cron string, start, end time.Time) ([]time.Time, error) {
	if end.Before(start) {
		return nil, fmt.Errorf("end can not earlier than start, start: %v, end: %v", start, end)
	}

	expr, err := cronexpr.Parse(cron)
	if err != nil {
		return nil, err
	}

	var nexts []time.Time
	for start.Before(end) {
		next := expr.Next(start)
		if next.UnixNano() < 0 {
			return nil, fmt.Errorf("fail to parse time from cron: %s", cron)
		}
		nexts = append(nexts, next)
		start = next
	}
	return nexts, nil
}
