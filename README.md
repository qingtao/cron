# cron
## 新建立的定时任务模块
### 使用方法：
```go
package main

import (
	"context"
	"cron"
	"fmt"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	cron := cron.New(ctx, cancel)

	go func() {
		for err := range cron.Err {
			fmt.Printf("ERROR: %s\n", err)
		}
	}()
	go cron.Start(ctx)

	s1 := "1/2 * * * * *"
	s1 := "15 13 * * * *"
	cron.AddFunc(ctx, "s1", s1, func() {
		fmt.Printf("s1 %s: %s\n", s1, time.Now())
	})
	cron.AddFunc(ctx, "s2", s2, func() {
		fmt.Printf("s2 %s: %s\n", s2, time.Now())
	})
.
.
.
}
```

只完成了功能测试，未完整验证性能和全部时间有效性

。
