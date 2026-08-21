package job

import (
	"context"
	"fmt"
	redisclient "serverSideEvents/client/redis"
	"serverSideEvents/timing"
	"time"
)

type TimeFrame int

const (
	Period1Min TimeFrame = iota
	Period5Min
	Period15Min
	Period1Hour
	Period1Day
)

type CandleBuilderJob struct {
	redis       *redisclient.Client
	TimingUtils *timing.Utils
	currency    string
	timeFrame   TimeFrame
}

func NewCandleBuilderJob(redis *redisclient.Client, timingUtils *timing.Utils, currency string, timeFrame TimeFrame) *CandleBuilderJob {
	return &CandleBuilderJob{
		redis:       redis,
		TimingUtils: timingUtils,
		currency:    currency,
		timeFrame:   timeFrame,
	}
}

func (j *CandleBuilderJob) Run(ctx context.Context) {

	j.TimingUtils.Synchronize(ctx, 30)

	deadline := time.Now().Add(seederDuration)
	ticker := time.NewTicker(seederInterval)
	defer ticker.Stop()

	fmt.Println("CandleBuilderJob: started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("CandleBuilderJob: stopped (context cancelled)")
			return

		case t := <-ticker.C:
			if t.After(deadline) {
				fmt.Println("CandleBuilderJob: completed (10 minutes elapsed)")
				return
			}

			timestampTags := j.FindTimestampTagsOfPeriod(time.Now(), Period1Min)

			//prices := make(map[string]string, len(timestampTags))
			for _, tag := range timestampTags {
				key := fmt.Sprintf("%s:%s", j.currency, tag)
				price, err := j.redis.Get(ctx, key, "price")
				_ = price
				if err != nil {
					continue
				}
				//prices[tag] = price
			}

			_ = t
		}
	}
}

func (j *CandleBuilderJob) BuildCandle() string {
	return ""
}

func (j *CandleBuilderJob) FindTimestampTagsOfPeriod(now time.Time, timeFrame TimeFrame) []string {
	var timeFrameStart time.Time
	var duration time.Duration

	switch timeFrame {
	case Period1Min:
		timeFrameStart = now.Truncate(1 * time.Minute)
		duration = 1 * time.Minute
		//TODO:implementation for the other timestamps will be different.
		/*case Period5Min:
			timeFrameStart = now.Truncate(5 * time.Minute)
			duration = 5 * time.Minute
		case Period15Min:
			timeFrameStart = now.Truncate(15 * time.Minute)
			duration = 15 * time.Minute
		case Period1Hour:
			timeFrameStart = now.Truncate(1 * time.Hour)
			duration = 1 * time.Hour
		case Period1Day:
			timeFrameStart = now.UTC().Truncate(24 * time.Hour)
			duration = 24 * time.Hour
		default:
			timeFrameStart = now.Truncate(1 * time.Minute)
			duration = 1 * time.Minute*/
	}

	totalSeconds := int(duration.Seconds())
	tags := make([]string, 0, totalSeconds)

	for i := range totalSeconds {
		t := timeFrameStart.UTC().Add(time.Duration(i) * time.Second)
		tags = append(tags, t.Format(timestampLayout))
	}

	return tags
}
