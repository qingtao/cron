/*
Package cron 提供一个基本的定时任务管理工具.

Example:

	package main

	import (
		"context"
		"cron"
		"fmt"
		"time"
	)

	func main() {
		//新的context和cancel
		ctx, cancel := context.WithCancel(context.Background())

		//创建新的cron.Cron
		c := cron.New(cancel)
		//处理错误
		go cr.Wait(func(err error) {
			fmt.Println(err)
		})

		//启动计划任务管理进程
		go c.Start(ctx)

		//定义时间
		s1 := "1/2 * * * * *"
		s2 := "15 13 * * * *"
		//添加计划任务
		c.AddFunc(ctx, "s1", s1, func() {
			fmt.Printf("s1 %s: %s\n", s1, time.Now())
		})
		c.AddFunc(ctx, "s2", s2, func() {
			fmt.Printf("s2 %s: %s\n", s2, time.Now())
		})
	}
*/
package cron
