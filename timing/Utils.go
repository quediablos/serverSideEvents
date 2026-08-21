package timing

import (
	"context"
	"fmt"
	"time"
)

type Utils struct {
}

func New() *Utils {
	return &Utils{}
}

func (u *Utils) Synchronize(ctx context.Context, millisecondMark time.Duration) {

	/* First synchronize the loop to start in the beginning of the new second. In order to so, the thread sleeps
	until the 50th millisecond (roughly) mark of the next second.
	*/
	now := time.Now()
	ms := now.Nanosecond() / int(time.Millisecond)
	fmt.Printf("currently at %dms within the second\n", ms)

	nextSecond := now.Truncate(time.Second).Add(time.Second)
	target := nextSecond.Add(millisecondMark * time.Millisecond)

	select {
	case <-time.After(time.Until(target)):
		// aligned to next second's 50ms mark
	case <-ctx.Done():
		return
	}
}
