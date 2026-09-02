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

outer:
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

			second := t.Second()

			now := time.Now()
			timestampTags := j.FindTimestampTagsOfPeriod(now, second, Period1Min)

			keyCandle := fmt.Sprintf("CANDLE_1_MIN_%s:%s", j.currency, timestampTags[0])

			//keyTicker := fmt.Sprintf("TICKER_%s:%s", j.currency, t.UTC().Format(timestampLayout))
			keyTicker := j.TimingUtils.MakeTickerKey(j.currency, t.UTC().Format(timestampLayout))
			priceTicker, err := j.redis.Get(ctx, keyTicker, "price")
			if err != nil {
				continue outer
			}

			//Build the candle for the first time and post it.
			if second == 0 {

				cnd := candle.Candle{
					Open:  priceTicker,
					High:  priceTicker,
					Low:   priceTicker,
					Close: priceTicker,
				}

				j.SaveCandle(ctx, keyCandle, &cnd)

				fmt.Println("CandleBuilderJob: candle(0):", cnd.Open, cnd.High, cnd.Low, cnd.Close)
			} else {

				candleExists := true
				//If the candle exists from the previous run, update it.
				candlePrevious, err := j.redis.GetCandle(ctx, keyCandle)
				if err != nil {
					continue outer
				}

				if candlePrevious.Open == "" {
					candleExists = false
				}

				//Candle from previous ticker exists, update it.
				if candleExists {

					cnd := j.UpdateFromPreviousCandle(candlePrevious, priceTicker)

					j.SaveCandle(ctx, keyCandle, &cnd)

					fmt.Println("CandleBuilderJob: candle(u):", cnd.Open, cnd.High, cnd.Low, cnd.Close)

				} else {
					//Cold start case where there is no previous data for the timeframe.
					//TODO:to be implemented later.
				}
			}

		}
	}
}

func (j *CandleBuilderJob) SaveCandle(ctx context.Context, keyCandle string, candle *candle.Candle) {
	j.redis.HSetMultiple(ctx, keyCandle,
		"open", candle.Open,
		"high", candle.High,
		"low", candle.Low, "close", candle.Close)
}

func (j *CandleBuilderJob) UpdateFromPreviousCandle(candlePrevious candle.Candle,
	priceTicker string) candle.Candle {
	cnd := candle.Candle{}

	//Compare high.
	highPrevious, _ := strconv.ParseFloat(candlePrevious.High, 32)
	tickerValue, _ := strconv.ParseFloat(priceTicker, 32)
	if tickerValue > highPrevious {
		cnd.High = priceTicker
	} else {
		cnd.High = candlePrevious.High
	}

	//Compare low
	lowPrevious, _ := strconv.ParseFloat(candlePrevious.Low, 32)
	if tickerValue < lowPrevious {
		cnd.Low = priceTicker
	} else {
		cnd.Low = candlePrevious.Low
	}

	cnd.Open = candlePrevious.Open
	cnd.Close = priceTicker
	return cnd
}

func (j *CandleBuilderJob) FindOHLCFromTickers(prices []string) candle.Candle {

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

func (j *CandleBuilderJob) FindTimestampTagsOfPeriod(now time.Time, second int, timeFrame TimeFrame) []string {
	var timeFrameStart time.Time

	switch timeFrame {
	case Period1Min:
		timeFrameStart = now.Truncate(1 * time.Minute)
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

	tags := make([]string, 0, second)

	if second == 0 {
		t := timeFrameStart.UTC().Add(time.Duration(0) * time.Second)
		tags = append(tags, t.Format(timestampLayout))
		return tags
	}

	for i := range second {
		t := timeFrameStart.UTC().Add(time.Duration(i) * time.Second)
		tags = append(tags, t.Format(timestampLayout))
	}

	return tags
}
