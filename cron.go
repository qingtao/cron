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
	Name string
	Time *Time
	Func func()

	cancel context.CancelFunc
	c      chan time.Time
}

func (job *Job) Cancel() {
	job.cancel()
	select {
	case <-job.c:
	default:
	}
	close(job.c)
}

func (job *Job) Run(ctx context.Context, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			errCh <- fmt.Errorf("job: %s - %s", job.Name, ctx.Err())
			return
		case u := <-job.c:
			if job.Time.Check(u) {
				go job.Func()
			}
		}
	}
}

type Cron struct {
	Jobs *sync.Map
	Err  chan error

	cancel context.CancelFunc
}

func New(ctx context.Context, cancel context.CancelFunc) *Cron {
	jobs, ch := new(sync.Map), make(chan error)
	return &Cron{jobs, ch, cancel}
}

func (c *Cron) AddFunc(ctx context.Context, name, s string, f func()) {
	t, err := Parse(s)
	if err != nil {
		c.Err <- err
	}
	ch := make(chan time.Time)
	ctx, cancel := context.WithCancel(ctx)

	job := &Job{name, t, f, cancel, ch}
	_, ok := c.Jobs.LoadOrStore(name, job)
	if ok {
		c.Err <- fmt.Errorf("job already exists: %s", name)
		return
	}
	go job.Run(ctx, c.Err)
}

func (c *Cron) Delete(name string) {
	job, ok := c.Jobs.Load(name)
	if ok {
		c.Jobs.Delete(name)
		if v, ok := job.(*Job); ok {
			v.cancel()
		}
	}
}

func (c *Cron) Start(ctx context.Context) {
	tick := time.Tick(1 * time.Second)
	for {
		select {
		case t := <-tick:
			c.Jobs.Range(func(k, v interface{}) bool {
				if job, ok := v.(*Job); ok {
					select {
					case job.c <- t:
					case <-time.After(time.Duration(jobTimeout) * time.Microsecond):
						c.Err <- errors.New("schedule check job timeout")
					}
				}
				return true
			})
		case <-ctx.Done():
			c.Err <- errors.New("cron schedule stopped")
			return
		}
	}
}

func (c *Cron) Stop() {
	c.cancel()
}
