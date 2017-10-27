// Package cron 提供一个基本的定时任务管理工具
// 只实现了功能，使用方法在测试文件和Readme

package cron

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// duration of send time to job's chan
	jobTimeout = 10

	// time used usually
	Monthly = "0 3 3 1 * *"
	Weekly  = "0 10 3 * * 0"
	Daily   = "0 23 3 * * *"
	Hourly  = "0 33 * * * *"
	Minute  = "0 * * * * *"
	Second  = "* * * * * *"
)

// get last day of month, m is the month, y is year
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
	//the last day February
	loc := time.Now().Location()
	last := time.Date(y, time.March, 1, 0, 0, 0, 0, loc).AddDate(0, 0, -1)
	return last.Day()
}

// Job is cron schedule work
type Job struct {
	// Name used for store or delete job
	Name string
	Time *Time
	// what to do
	Func func()

	// used for stop the job
	cancel context.CancelFunc
	// receive time
	c chan time.Time
}

// cancel the job
func (job *Job) Cancel() {
	job.cancel()
	select {
	case <-job.c:
	default:
	}
	close(job.c)
}

// run job, return when  read from ctx.Done()
func (job *Job) Run(ctx context.Context, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			errCh <- fmt.Errorf("job: %s - %s", job.Name, ctx.Err())
			return
		case u := <-job.c:
			// check time, run job.Func or not
			if job.Time.Check(u) {
				go job.Func()
			}
		}
	}
}

// cron 管理所有任务，每一秒向Jobs内的所有成员发送当前时间
type Cron struct {
	Jobs *sync.Map
	Err  chan error

	cancel context.CancelFunc
}

// 创建一个新的Cron, 使用context.Context和context.CancelFunc
func New(ctx context.Context, cancel context.CancelFunc) *Cron {
	jobs, ch := new(sync.Map), make(chan error)
	return &Cron{jobs, ch, cancel}
}

// 添加成员到Cron
func (c *Cron) AddFunc(ctx context.Context, name, s string, f func()) {
	t, err := Parse(s)
	if err != nil {
		c.Err <- err
	}
	ch := make(chan time.Time)
	ctx, cancel := context.WithCancel(ctx)

	job := &Job{name, t, f, cancel, ch}
	_, ok := c.Jobs.LoadOrStore(name, job)
	// 如果任务已经存在，返回错误到c.Err，否则执行job.Func
	if ok {
		c.Err <- fmt.Errorf("job already exists: %s", name)
		return
	}
	go job.Run(ctx, c.Err)
}

// 删除名称为name的任务
func (c *Cron) Delete(name string) {
	job, ok := c.Jobs.Load(name)
	if ok {
		c.Jobs.Delete(name)
		if v, ok := job.(*Job); ok {
			v.cancel()
		}
	}
}

// 启动Cron，定时器为1秒
func (c *Cron) Start(ctx context.Context) {
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case t := <-tick:
			//遍历Jobs的成员，发送当前时间到每个job.c
			c.Jobs.Range(func(k, v interface{}) bool {
				if job, ok := v.(*Job); ok {
					select {
					case job.c <- t:
					//发送超时时间
					case <-time.After(time.Duration(jobTimeout) * time.Microsecond):
						c.Err <- errors.New("schedule check job timeout")
					}
				}
				return true
			})
		//等待退出命令
		case <-ctx.Done():
			c.Err <- errors.New("cron schedule stopped")
			return
		}
	}
}

func (c *Cron) Stop() {
	c.cancel()
}
