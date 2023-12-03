package snowflake

import (
  "github.com/stretchr/testify/assert"
  "testing"
  "time"
)

func TestSnowflake_Generate(t *testing.T) {
  option := Option{
    Mode:        Normal,
    StartTime:   time.Date(2023, 12, 1, 0, 0, 0, 0, time.Local),
    NodeID:      1,
    MaxWaitTime: 0,
  }
  snowflake, err := NewSnowflake(option)
  assert.NoError(t, err)
  id, err := snowflake.GenerateID()
  id, err = snowflake.GenerateID()
  assert.NoError(t, err)
  idTime, nodeID, step, err := snowflake.Parse(id)
  assert.NoError(t, err)
  if idTime < option.StartTime.UnixMilli() {
    t.Error("Generate time is error.")
  }
  if int16(nodeID) != option.NodeID {
    t.Error("Generate nodeID is error.")
  }
  t.Log(step)
}

func TestSnowflake_Generate_Auto(t *testing.T) {
  option := Option{
    Mode:        AutoTime,
    StartTime:   time.Date(2023, 12, 1, 0, 0, 0, 0, time.Local),
    NodeID:      1,
    MaxWaitTime: 0,
  }
  snowflake, err := NewSnowflake(option)
  assert.NoError(t, err)
  id, err := snowflake.GenerateID()
  id, err = snowflake.GenerateID()
  assert.NoError(t, err)
  idTime, nodeID, step, err := snowflake.Parse(id)
  assert.NoError(t, err)
  if idTime < option.StartTime.UnixMilli() {
    t.Error("Generate time is error.")
  }
  if int16(nodeID) != option.NodeID {
    t.Error("Generate nodeID is error.")
  }
  sn := snowflake.(*Snowflake)
  sn.lastTimestamp = sn.lastTimestamp + 1000
  lastTimestamp := sn.lastTimestamp
  sn.step = 1<<StepBits - 1
  id, err = sn.GenerateID()
  assert.NoError(t, err)
  idTime, nodeID, step, err = snowflake.Parse(id)
  assert.NoError(t, err)
  if idTime != (lastTimestamp + 1) {
    t.Error("Generate time is error.")
  }
  idTime, nodeID, step, err = snowflake.Parse(id)
  assert.NoError(t, err)
  t.Log(step)

  time.Sleep(1050 * time.Millisecond)
  idTime, nodeID, step, err = snowflake.Parse(id)
  assert.NoError(t, err)
  if idTime < lastTimestamp+1 {
    t.Error("Generate time is error.")
  }
  t.Log(time.UnixMilli(lastTimestamp).String())
  t.Log(time.UnixMilli(idTime).String())
}
