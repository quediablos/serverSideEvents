package job

import (
	"context"
	"fmt"
	"serverSideEvents/candle"
	redisclient "serverSideEvents/client/redis"
	"serverSideEvents/timing"
	"strconv"
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
			prices := make([]string, 60)

			for _, tag := range timestampTags {
				key := fmt.Sprintf("%s:%s", j.currency, tag)
				price, err := j.redis.Get(ctx, key, "price")

				//Too early to see all the tickers of the whole time frame.
				/*if err != nil {
					continue outer
				}*/

				prices = append(prices, price)

				if err != nil {
					continue
				}
			}

			cnd := j.FindOHLC(prices)
			_ = cnd

		}
	}
}

func (j *CandleBuilderJob) BuildCandle() string {
	return ""
}

func (j *CandleBuilderJob) FindOHLC(prices []string) candle.Candle {

	cnd := candle.Candle{}
	cnd.Open = prices[0]
	cnd.Close = prices[len(prices)-1]

	var high, _ = strconv.ParseFloat(prices[0], 32)
	var low, _ = strconv.ParseFloat(prices[0], 32)

	for _, price := range prices {

		p, err := strconv.ParseFloat(price, 64)
		if err != nil {
			continue
		}
		if p > high {
			high = p
		}
		if p < low {
			low = p
		}
	}

	cnd.High = strconv.FormatFloat(high, 'f', -1, 32)
	cnd.Low = strconv.FormatFloat(low, 'f', -1, 32)
	return cnd
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
