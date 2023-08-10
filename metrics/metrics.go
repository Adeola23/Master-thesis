package metrics

import (
	"fmt"
	"strings"
	"time"
)

type Metrics struct {
	topic string
	addr  string

	tm time.Time

	handshake time.Duration
	read      time.Duration
	handle    time.Duration
	write     time.Duration

	stat []string
}

func NewMetrics(addr string) (m *Metrics) {
	return &Metrics{
		addr: addr,
		tm:   time.Now(),

		handshake: -1,
		read:      -1,
		handle:    -1,
		write:     -1,

		stat: []string{fmt.Sprintf("addr (%s)", addr)},
	}
}

func (m *Metrics) Reset() {
	m.tm = time.Now()
}

func (m *Metrics) SetTopic(topic string) {
	m.topic = topic
}

const statPattern = "%s (%s)"

func (m *Metrics) FixHandshake() {
	m.handshake = time.Since(m.tm)
	m.stat = append(m.stat, fmt.Sprintf(statPattern, "handshake", prepareValue(m.handshake)))
	m.Reset()
}

func (m *Metrics) FixReadDuration() {
	m.read = time.Since(m.tm)
	m.stat = append(m.stat, fmt.Sprintf(statPattern, "read", prepareValue(m.read)))
	m.Reset()
}

func (m *Metrics) FixHandleDuration() {
	m.handle = time.Since(m.tm)
	m.stat = append(m.stat, fmt.Sprintf(statPattern, "handle", prepareValue(m.handle)))
	m.Reset()
}

func (m *Metrics) FixWriteDuration() {
	m.write = time.Since(m.tm)
	m.stat = append(m.stat, fmt.Sprintf(statPattern, "write", prepareValue(m.write)))
	m.Reset()
}

func (m *Metrics) String() (line string) {
	var total time.Duration

	if m.handshake >= 0 {
		total += m.handshake
	}

	if m.read >= 0 {
		total += m.read
	}

	if m.handle >= 0 {
		total += m.handle
	}

	if m.write >= 0 {
		total += m.write
	}

	if total > 0 {
		m.stat = append(m.stat, fmt.Sprintf(statPattern, "total", prepareValue(total)))
	}

	line = fmt.Sprintf("%s: %s", m.topic, strings.Join(m.stat, ", "))

	return line
}

func prepareValue(dur time.Duration) (size string) {
	const valuePattern = "%d %ss"

	ns := dur.Nanoseconds()
	if ns < 1000 {
		return fmt.Sprintf(valuePattern, ns, "n")
	}

	mcs := dur.Microseconds()
	if mcs < 1000 {
		return fmt.Sprintf(valuePattern, mcs, "µ")
	}

	ms := dur.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf(valuePattern, ms, "m")
	}

	return fmt.Sprintf("%.2f s", dur.Seconds())
}
