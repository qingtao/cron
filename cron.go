//Package cron 提供一个基本的定时任务管理工具
//  使用方法在测试文件和Readme
package cron

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	//duration of send time to job's chan
	jobTimeout = 10

	//time used usually
	//每月
	Monthly = "0 3  3 1 * *"
	//每周
	Weekly = "0 13 3 * * 0"
	//每天
	Daily = "0 23 3 * * *"
	//每小时
	Hourly = "0 33 * * * *"
	//每分钟
	Minute = "0 *  * * * *"
	//每秒
	Second = "* *  * * * *"
)

var (
	//ErrCronClosed关闭cron
	ErrCronClosed = errors.New("Cron schedule stopped")
	//ErrCleanChannel清理channel
	ErrCleanChannel = errors.New("Cron stopped, clean error channel")
	//ErrChanClosed channel已关闭
	ErrChanClosed = errors.New("Channel closed")
)

//Job is cron schedule work
type Job struct {
	//Name used for store or delete job
	Name string
	//Time任务时间定义
	Time *Time
	//Func is what to do
	Func func()

	//used for stop the job
	cancel context.CancelFunc
	//receive time
	c chan time.Time
}

//Cancel the job
func (job *Job) Cancel() {
	job.cancel()
	select {
	case <-job.c:
	default:
	}
}

//Run job, return when  read from ctx.Done()
func (job *Job) Run(ctx context.Context, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			send(errCh, fmt.Errorf("Job [%s] %s", job.Name, ctx.Err()))
			return
		case u, ok := <-job.c:
			// check time, run job.Func or not
			if ok && job.Time.Check(u) {
				go job.Func()
			}
		}
	}
}

//Cron 管理所有任务，每一秒向Jobs内的所有成员发送当前时间
type Cron struct {
	Jobs *sync.Map

	err    chan error
	cancel context.CancelFunc
}

//New创建一个新的Cron, 使用context.WithCancel新建ctx和cancel, cancel在New函数使用，ctx分别用于Start函数和AddFunc函数
func New(cancel context.CancelFunc) *Cron {
	jobs, ch := new(sync.Map), make(chan error)
	return &Cron{jobs, ch, cancel}
}

//Wait等待读取cron发送的错误，f的参数是error
func (c *Cron) Wait(f func(error)) {
	for {
		err, ok := <-c.err
		if !ok {
			f(ErrChanClosed)
			return
		}
		f(err)
	}
}

//AddFunc添加成员到Cron，ctx是context.WithCancel的返回值
func (c *Cron) AddFunc(ctx context.Context, name, s string, f func()) {
	t, err := Parse(s)
	if err != nil {
		send(c.err, err)
	}
	ch := make(chan time.Time)
	ctx, cancel := context.WithCancel(ctx)

	job := &Job{name, t, f, cancel, ch}
	_, ok := c.Jobs.LoadOrStore(name, job)
	//如果任务已经存在，返回错误到c.err，否则执行job.Func
	if ok {
		send(c.err, fmt.Errorf("Job [%s] is already exists", name))
		return
	}
	go job.Run(ctx, c.err)
}

//Delete删除名称为name的任务
func (c *Cron) Delete(name string) {
	job, ok := c.Jobs.Load(name)
	if ok {
		c.Jobs.Delete(name)
		if v, ok := job.(*Job); ok {
			v.Cancel()
		}
	}
}

func send(ch chan<- error, err error) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	ch <- err
}

//Start启动Cron，定时器为1秒
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
						send(c.err, fmt.Errorf("Cron wake job [%s] timeout", job.Name))
					}
				}
				return true
			})
		//等待退出命令
		case <-ctx.Done():
			send(c.err, ErrCronClosed)
			close(c.err)
			return
		}
	}
}

//Stop停止计划任务管理器
func (c *Cron) Stop() {
	c.cancel()
}
