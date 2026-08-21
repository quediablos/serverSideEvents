

**1.How to connect and post ticket data in Redis**

Start a docker container in local with port mapping of 6379:6379.

Exec into the container.

Run "redis-cli" command.

To insert a ticket data run either one of below:

    HSET <currency> price "<price>" at "<timestamp (UTC) with seconds which represents the second the ticket belongs to>"
    HSET USD price "47.02" at "2026-08-19 14:53:00"
or

    HSET <currency> price "<price>" at "last"
    HSET USD price "47.02"

The second entry represents the default ticket as the last one, the other ones can be used for historic data as well.