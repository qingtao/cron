package cron

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var delayTime = 900

func TestCron(t *testing.T) {
	s1 := "1/2 * * * * *"
	s2 := fmt.Sprintf("1/2 %d * * * *", time.Now().Minute())
	s21 := fmt.Sprintf("1/3 %d * * * *", time.Now().Minute())
	s22 := fmt.Sprintf("1/4 %d * * * *", time.Now().Minute())
	s23 := fmt.Sprintf("1/6 %d * * * *", time.Now().Minute())
	s3 := "1/5 * * * * *"
	s4 := "8/12 * * * * *"
	ctx, cancel := context.WithCancel(context.Background())
	cr := New(cancel)

	go cr.Wait(func(err error) {
		fmt.Println(err)
	})

	fmt.Printf("%s\n", time.Now())
	fmt.Println("add s1")
	cr.AddFunc(ctx, "s1", s1, delayTime, func() {
		fmt.Printf("s1 %s: %s\n", s1, time.Now())
	})
	go cr.Start(ctx)
	fmt.Println("add s2")
	cr.AddFunc(ctx, "s2", s2, delayTime, func() {
		fmt.Printf("s2 %s: %s\n", s2, time.Now())
	})
	fmt.Println("add s21")
	cr.AddFunc(ctx, "s21", s21, delayTime, func() {
		fmt.Printf("s21 %s: %s\n", s21, time.Now())
	})
	fmt.Println("add s22")
	cr.AddFunc(ctx, "s22", s22, delayTime, func() {
		fmt.Printf("s22 %s: %s\n", s22, time.Now())
	})
	fmt.Println("add s23")
	cr.AddFunc(ctx, "s23", s23, delayTime, func() {
		fmt.Printf("s23 %s: %s\n", s23, time.Now())
	})
	fmt.Println("add s3")
	cr.AddFunc(ctx, "s3", s3, delayTime, func() {
		fmt.Printf("s3 %s: %s\n", s3, time.Now())
	})
	fmt.Println("add s4, but name is s3")
	cr.AddFunc(ctx, "s3", s4, delayTime, func() {
		fmt.Printf("s4 %s: %s\n", s4, time.Now())
	})

	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("delete s1")
		cr.Delete("s1")
	}()
	time.Sleep(20 * time.Second)
	fmt.Println("stop cron")
	cr.Stop()
	time.Sleep(2 * time.Second)
	fmt.Println("wait stop!")
}
