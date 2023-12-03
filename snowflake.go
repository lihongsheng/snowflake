package snowflake

import (
	"errors"
	"sync"
	"time"
)

type Generate interface {
	GenerateID() (int64, error)
	Parse(id int64) (time int64, node int64, seq int64, err error)
}

type Mode int8

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

const (
	// NodeBits holds the number of bits to use for Node
	// Remember, you have a total 22 bits to share between Node/Step
	NodeBits uint8 = 10
	// StepBits holds the number of bits to use for Step
	// Remember, you have a total 22 bits to share between Node/Step
	StepBits uint8 = 12
)

type Option struct {
	Mode        Mode
	StartTime   time.Time
	NodeID      int16
	MaxWaitTime time.Duration
}

// Snowflake
// +--------------------------------------------------------------------------+
// | 1 Bit Unused | 41 Bit Timestamp |  10 Bit NodeID  |   12 Bit Sequence ID |
// +--------------------------------------------------------------------------+
type Snowflake struct {
	node          int64
	step          int64
	MaxStep       int64
	epoch         int64
	lastTimestamp int64
	mode          Mode
	mutex         sync.Mutex
	maxWaitTime   time.Duration
}

var (
	ErrStartTimeAhead   = errors.New("start time is ahead of now")
	ErrOverTimeRollback = errors.New("time rollback")
	ErrInvalidMachineID = errors.New("invalid machine id")
	ErrNoMode           = errors.New("mode  is nil")
)

func NewSnowflake(option Option) (Generate, error) {
	if option.StartTime.After(time.Now()) {
		return nil, ErrStartTimeAhead
	}

	if option.NodeID > (1<<NodeBits - 1) {
		return nil, ErrInvalidMachineID
	}

	return &Snowflake{
		node:    int64(option.NodeID),
		step:    0,
		MaxStep: -1 ^ (-1 << StepBits),
		epoch:   option.StartTime.UnixMilli(),
		mode:    option.Mode,
		mutex:   sync.Mutex{},
		//	timestampLeftShift: stepBits + nodeBits,
		maxWaitTime: option.MaxWaitTime,
	}, nil
}

func (s *Snowflake) GenerateID() (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	switch s.mode {
	case Normal: // 直接返回错，让客户端重试
		return s.normalNextID()
	case AutoTime: // 自增模式, 自动增加 lastTimestamp
		return s.autoNextID()
	}
	return 0, ErrNoMode
}
func (s *Snowflake) autoNextID() (int64, error) {
	currentTime := s.getCurrentTime()
	if currentTime < s.lastTimestamp {
		currentTime = s.lastTimestamp
	}
	if currentTime == s.lastTimestamp {
		s.step = (s.step + 1) & s.MaxStep
		if s.step == 0 {
			currentTime++
		}
	} else {
		s.step = 0
	}
	s.lastTimestamp = currentTime
	return ((s.lastTimestamp - s.epoch) << (NodeBits + StepBits)) | (s.node << StepBits) | s.step, nil
}

func (s *Snowflake) normalNextID() (int64, error) {
	currentTime := s.getCurrentTime()
	if currentTime < s.lastTimestamp {
		return 0, ErrOverTimeRollback
	}
	if currentTime == s.lastTimestamp {
		s.step = (s.step + 1) & s.MaxStep
		if s.step == 0 {
			currentTime = s.timeMills(s.lastTimestamp)
		}
	} else {
		s.step = 0
	}
	s.lastTimestamp = currentTime
	return ((s.lastTimestamp - s.epoch) << (NodeBits + StepBits)) | (s.node << StepBits) | s.step, nil
}

func (s *Snowflake) getCurrentTime() int64 {
	currentTime := time.Now().UnixMilli()
	if currentTime < s.lastTimestamp {
		if s.maxWaitTime > 0 {
			time.Sleep(s.maxWaitTime)
		}
		currentTime = time.Now().UnixMilli()
	}
	return currentTime
}

// timeMills
func (s *Snowflake) timeMills(lastTime int64) int64 {
	currentTime := time.Now().UnixMilli()
	for currentTime <= lastTime {
		currentTime = time.Now().UnixMilli()
	}
	return currentTime
}

func (s *Snowflake) Parse(id int64) (time int64, node int64, seq int64, err error) {
	time = id>>(NodeBits+StepBits) + s.epoch
	node = (1<<NodeBits - 1) << StepBits
	node = id & node >> StepBits
	seq = id & s.MaxStep
	return
}
