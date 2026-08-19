package job

import (
	"context"
	"fmt"
	"math/rand"
	redisclient "serverSideEvents/client/redis"
	"time"
)

const (
	seederDuration  = 10 * time.Minute
	seederInterval  = 1 * time.Second
	priceMin        = 47.00
	priceMax        = 48.00
	timestampLayout = "2006_01_02_15_04_05"
)

type UsdPriceSeederJob struct {
	redis *redisclient.Client
}

func NewUsdPriceSeederJob(redis *redisclient.Client) *UsdPriceSeederJob {
	return &UsdPriceSeederJob{redis: redis}
}

func (j *UsdPriceSeederJob) Run(ctx context.Context) {
	deadline := time.Now().Add(seederDuration)
	ticker := time.NewTicker(seederInterval)
	defer ticker.Stop()

	fmt.Println("UsdPriceSeederJob: started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("UsdPriceSeederJob: stopped (context cancelled)")
			return

		case t := <-ticker.C:
			if t.After(deadline) {
				fmt.Println("UsdPriceSeederJob: completed (10 minutes elapsed)")
				return
			}

			key := fmt.Sprintf("USD:%s", t.UTC().Format(timestampLayout))
			price := fmt.Sprintf("%.2f", priceMin+rand.Float64()*(priceMax-priceMin))

			if err := j.redis.HSet(ctx, key, "price", price); err != nil {
				fmt.Printf("UsdPriceSeederJob: failed to write %s: %v\n", key, err)
				continue
			}

			fmt.Printf("UsdPriceSeederJob: set %s price=%s\n", key, price)
		}
	}
}
