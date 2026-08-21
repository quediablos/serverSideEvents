
**1.Data Format in Redis**

1.1.Example price ticker for each second:

    HSET TICKER_USDTRY:2026_08_21_14_13_01 price "47.08"

1.1.Price candle for each minute: 

    HSET CANDLE_1MIN_USDTRY:2026_08_21_14_13_00 open "47.08" high "48.10" low "47.05" close "48.00"

