package job

import (
	"context"
	"fmt"
	"math/rand"
	redisclient "serverSideEvents/client/redis"
	"serverSideEvents/timing"
	"time"
)

const (
	seederDuration  = 10 * time.Minute
	seederInterval  = 1 * time.Second
	priceMin        = 47.00
	priceMax        = 48.00
	timestampLayout = "2006_01_02_15_04_05"
	millisecondMark = 20
)

type PriceSeederJob struct {
	redis       *redisclient.Client
	TimingUtils *timing.Utils
	currency    string
}

func NewPriceSeederJob(redis *redisclient.Client, TimingUtils *timing.Utils, currency string) *PriceSeederJob {
	return &PriceSeederJob{redis: redis,
		TimingUtils: TimingUtils,
		currency:    currency}
}

func (j *PriceSeederJob) Run(ctx context.Context) {

	j.TimingUtils.Synchronize(ctx, millisecondMark)

	deadline := time.Now().Add(seederDuration)
	ticker := time.NewTicker(seederInterval)
	defer ticker.Stop()

	fmt.Println("PriceSeederJob: started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("PriceSeederJob: stopped (context cancelled)")
			return

		case t := <-ticker.C:
			if t.After(deadline) {
				fmt.Println("PriceSeederJob: completed (10 minutes elapsed)")
				return
			}

			key := fmt.Sprintf("TICKER_%s:%s", j.currency, t.UTC().Format(timestampLayout))
			price := fmt.Sprintf("%.2f", priceMin+rand.Float64()*(priceMax-priceMin))

			if err := j.redis.HSet(ctx, key, "price", price); err != nil {
				fmt.Printf("PriceSeederJob: failed to write %s: %v\n", key, err)
				continue
			}

			//fmt.Printf("PriceSeederJob: set %s price=%s\n", key, price)
		}
	}
}
