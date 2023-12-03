### snowflake
Snowflake is a distributed unique ID generator inspired by [Twitter's Snowflake](https://blog.twitter.com/2010/announcing-snowflake).
```dtd
  +--------------------------------------------------------------------------+
  | 1 Bit Unused | 41 Bit Timestamp |  10 Bit NodeID  |   12 Bit Sequence ID |
  +--------------------------------------------------------------------------+
```
### Installation
```dtd
go get github.com/lihongsheng/snowflake
```
### Usage
The function New creates a new snowflake instance.
```
func NewSnowflake(opt Option) (*snowflake, error)
```

You can configure snowflake by the struct Settings:
```go
const (
// Normal is dependent time. if Time rollback when return error.
// 正常模式下 snowflake 依赖时钟，如果出现时间回滚会返回error
Normal Mode = iota
// AutoTime
// if Time rollback , AutoTime is auto add mills.
// When the time is greater than the current time, it will switch to time dependent mode again。
// 自动模式下，如果出现时钟回滚，当步长超过最大值时候会自动追加时间毫秒数。此时不在依赖时钟，当获取的系统时间再次大于snowflake时间时候
// 会恢复到正常模式。
AutoTime
)
type Option struct {
  Mode        Mode
  StartTime   time.Time
  NodeID      int16
  MaxWaitTime time.Duration
}
```
StartTime is the time since which the snowflake time is defined as the elapsed time. If StartTime is 0, the start time of the Sonyflake is set to "2014-09-01 00:00:00 +0000 UTC". If StartTime is ahead of the current time, Sonyflake is not created.

NodeID returns the unique ID of the snowflake instance. If NodeID returns an error, snowflake is not created.

Mode What mode to handle when clock rollback occurs. normal is return error. but AutoTime is to auto add time millis, when sys time is greater than  last time, Will use system time again。

In order to get a new unique ID, you just have to call the method NextID.
```go
func (s *snowflake) GenerateID() (uint64, error)
```


