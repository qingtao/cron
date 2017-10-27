package cron

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	s1 := "1/2 * * * * *"
	s2 := "1/2 3 * * * *"
	s3 := "1/5 * * * * *"
	s4 := "8/12 * * * * *"
	ctx, cancel := context.WithCancel(context.Background())
	cr := New(ctx, cancel)

	go func() {
		for err := range cr.Err {
			fmt.Printf("ERROR: %s\n", err)
		}
	}()

	fmt.Printf("%s\n", time.Now())
	fmt.Println("add s1")
	cr.AddFunc(ctx, "s1", s1, func() {
		fmt.Printf("s1 %s: %s\n", s1, time.Now())
	})
	go cr.Start(ctx)
	fmt.Println("add s2")
	cr.AddFunc(ctx, "s2", s2, func() {
		fmt.Printf("s2 %s: %s\n", s2, time.Now())
	})
	fmt.Println("add s3")
	cr.AddFunc(ctx, "s3", s3, func() {
		fmt.Printf("s3 %s: %s\n", s3, time.Now())
	})
	fmt.Println("add s4, but name is s3")
	cr.AddFunc(ctx, "s3", s4, func() {
		fmt.Printf("s4 %s: %s\n", s4, time.Now())
	})

	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("delete s1")
		cr.Delete("s1")
	}()
	time.Sleep(10 * time.Second)
	fmt.Println("stop cron")
	cr.Stop()
	fmt.Println("wait stop!")
	time.Sleep(1 * time.Second)
}
