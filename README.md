# RediQueue

Pure Go queue based on REDIS protocol.


## Commands

Implemented commands:

 - Connection (complete)
   - AUTH -- see RequireAuth()
   - ECHO
   - PING
   - SELECT
   - QUIT
 - Key 
   - DEL
   - EXISTS
   - EXPIRE
   - EXPIREAT
   - KEYS
   - MOVE
   - ~~PERSIST~~
   - ~~PEXPIRE~~
   - ~~PEXPIREAT~~
   - ~~PTTL~~
   - RENAME
   - RENAMENX
   - RANDOMKEY -- call math.rand.Seed(...) once before using.
   - ~~TTL~~
   - TYPE
   - SCAN
 - Transactions (complete)
   - DISCARD
   - EXEC
   - MULTI
   - UNWATCH
   - WATCH
 - Server
   - DBSIZE
   - FLUSHALL
   - FLUSHDB
 - List keys (complete)
   - BLPOP
   - BRPOP
   - BRPOPLPUSH
   - LINDEX
   - LINSERT
   - LLEN
   - LPOP
   - LPUSH
   - LPUSHX
   - LRANGE
   - LREM
   - LSET
   - LTRIM
   - RPOP
   - RPOPLPUSH
   - RPUSH
   - RPUSHX
 - Set keys (complete)
   - SADD
   - SCARD
   - SDIFF
   - SDIFFSTORE
   - SINTER
   - SINTERSTORE
   - SISMEMBER
   - SMEMBERS
   - SMOVE
   - SPOP -- call math.rand.Seed(...) once before using.
   - SRANDMEMBER -- call math.rand.Seed(...) once before using.
   - SREM
   - SUNION
   - SUNIONSTORE
   - SSCAN

## Not supported

Commands which will probably not be implemented:

 - CLUSTER (all)
    - ~~CLUSTER *~~
    - ~~READONLY~~
    - ~~READWRITE~~
 - GEO (all) -- unless someone needs these
    - ~~GEOADD~~
    - ~~GEODIST~~
    - ~~GEOHASH~~
    - ~~GEOPOS~~
    - ~~GEORADIUS~~
    - ~~GEORADIUSBYMEMBER~~
 - HyperLogLog (all) -- unless someone needs these
    - ~~PFADD~~
    - ~~PFCOUNT~~
    - ~~PFMERGE~~
 - Key
    - ~~DUMP~~
    - ~~MIGRATE~~
    - ~~OBJECT~~
    - ~~RESTORE~~
    - ~~WAIT~~
 - Pub/Sub (all)
    - ~~PSUBSCRIBE~~
    - ~~PUBLISH~~
    - ~~PUBSUB~~
    - ~~PUNSUBSCRIBE~~
    - ~~SUBSCRIBE~~
    - ~~UNSUBSCRIBE~~
 - Scripting (all)
    - ~~EVAL~~
    - ~~EVALSHA~~
    - ~~SCRIPT *~~
 - Server
    - ~~BGSAVE~~
    - ~~BGWRITEAOF~~
    - ~~CLIENT *~~
    - ~~COMMAND *~~
    - ~~CONFIG *~~
    - ~~DEBUG *~~
    - ~~INFO~~
    - ~~LASTSAVE~~
    - ~~MONITOR~~
    - ~~ROLE~~
    - ~~SAVE~~
    - ~~SHUTDOWN~~
    - ~~SLAVEOF~~
    - ~~SLOWLOG~~
    - ~~SYNC~~
    - ~~TIME~~
 - Hash
 - String
 - SortedSet
