package runner

import "sync/atomic"

// Summary holds aggregate conversion counts.
type Summary struct {
	OK   int64
	Fail int64
	Skip int64
}

// collector accumulates job results atomically.
type collector struct {
	ok   atomic.Int64
	fail atomic.Int64
	skip atomic.Int64
}

func (c *collector) incOK()   { c.ok.Add(1) }
func (c *collector) incFail() { c.fail.Add(1) }
func (c *collector) incSkip() { c.skip.Add(1) }

func (c *collector) summary() Summary {
	return Summary{
		OK:   c.ok.Load(),
		Fail: c.fail.Load(),
		Skip: c.skip.Load(),
	}
}
