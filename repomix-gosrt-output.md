This file is a merged representation of a subset of the codebase, containing specifically included files, combined into a single document by Repomix.

# File Summary

## Purpose
This file contains a packed representation of a subset of the repository's contents that is considered the most important context.
It is designed to be easily consumable by AI systems for analysis, code review,
or other automated processes.

## File Format
The content is organized as follows:
1. This summary section
2. Repository information
3. Directory structure
4. Repository files (if enabled)
5. Multiple file entries, each consisting of:
  a. A header with the file path (## File: path/to/file)
  b. The full contents of the file in a code block

## Usage Guidelines
- This file should be treated as read-only. Any changes should be made to the
  original repository files, not this packed version.
- When processing this file, use the file path to distinguish
  between different files in the repository.
- Be aware that this file may contain sensitive information. Handle it with
  the same level of security as you would the original repository.

## Notes
- Some files may have been excluded based on .gitignore rules and Repomix's configuration
- Binary files are not included in this packed representation. Please refer to the Repository Structure section for a complete list of file paths, including binary files
- Only files matching these patterns are included: **/*.go, **/*.md
- Files matching patterns in .gitignore are excluded
- Files matching default ignore patterns are excluded
- Files are sorted by Git change count (files with more changes are at the bottom)

# Directory Structure
```
circular/circular_test.go
circular/circular.go
CODE_OF_CONDUCT.md
config_test.go
config.go
congestion/congestion.go
congestion/live/doc.go
congestion/live/fake.go
congestion/live/receive_test.go
congestion/live/receive.go
congestion/live/send_test.go
congestion/live/send.go
conn_request.go
connection_test.go
connection.go
contrib/client/main.go
contrib/client/reader.go
contrib/client/writer.go
contrib/server/main.go
crypto/crypto_test.go
crypto/crypto.go
dial_test.go
dial.go
doc.go
listen_test.go
listen.go
log_test.go
log.go
net_windows.go
net.go
net/ip_test.go
net/ip.go
net/syncookie_test.go
net/syncookie.go
packet_conn.go
packet/packet_test.go
packet/packet.go
pubsub_test.go
pubsub.go
rand/rand_test.go
rand/rand.go
README.md
SECURITY.md
server_test.go
server.go
statistics.go
```

# Files

## File: circular/circular_test.go
````go
package circular

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const max uint32 = 0b11111111_11111111_11111111_11111111

func ExampleNumber_Inc() {
	a := New(42, max)
	b := a.Inc()

	fmt.Println(b.Val())
	// Output: 43
}

func TestIncNoWrap(t *testing.T) {
	a := New(42, max)

	require.Equal(t, uint32(42), a.Val())

	a = a.Inc()

	require.Equal(t, uint32(43), a.Val())
}

func TestIncWrap(t *testing.T) {
	a := New(max-1, max)

	require.Equal(t, max-1, a.Val())

	a = a.Inc()

	require.Equal(t, max, a.Val())

	a = a.Inc()

	require.Equal(t, uint32(0), a.Val())
}

func TestDecNoWrap(t *testing.T) {
	a := New(42, max)

	require.Equal(t, uint32(42), a.Val())

	a = a.Dec()

	require.Equal(t, uint32(41), a.Val())
}

func TestDecWrap(t *testing.T) {
	a := New(0, max)

	require.Equal(t, uint32(0), a.Val())

	a = a.Dec()

	require.Equal(t, max, a.Val())

	a = a.Dec()

	require.Equal(t, max-1, a.Val())
}

func TestDistanceNoWrap(t *testing.T) {
	a := New(42, max)
	b := New(50, max)

	d := a.Distance(b)

	require.Equal(t, uint32(8), d)

	d = b.Distance(a)

	require.Equal(t, uint32(8), d)
}

func TestDistanceWrap(t *testing.T) {
	a := New(2, max)
	b := New(max-2, max)

	d := a.Distance(b)

	require.Equal(t, uint32(5), d)

	d = b.Distance(a)

	require.Equal(t, uint32(5), d)
}

func TestLt(t *testing.T) {
	a := New(42, max)
	b := New(50, max)
	c := New(max-10, max)

	x := a.Lt(b)

	require.Equal(t, true, x)

	x = b.Lt(a)

	require.Equal(t, false, x)

	x = a.Lt(c)

	require.Equal(t, false, x)

	x = c.Lt(a)

	require.Equal(t, true, x)
}

func TestGt(t *testing.T) {
	a := New(42, max)
	b := New(50, max)
	c := New(max-10, max)

	x := a.Gt(b)

	require.Equal(t, false, x)

	x = b.Gt(a)

	require.Equal(t, true, x)

	x = a.Gt(c)

	require.Equal(t, true, x)

	x = c.Gt(a)

	require.Equal(t, false, x)
}

func TestAdd(t *testing.T) {
	a := New(max-42, max)

	a = a.Add(42)

	require.Equal(t, max, a.Val())

	a = a.Add(1)

	require.Equal(t, uint32(0), a.Val())
}

func TestSub(t *testing.T) {
	a := New(42, max)

	a = a.Sub(42)

	require.Equal(t, uint32(0), a.Val())

	a = a.Sub(1)

	require.Equal(t, max, a.Val())
}
````

## File: circular/circular.go
````go
// Package circular implements "circular numbers". This is a number that can be
// increased (or decreased) indefinitely while only using up a limited amount of
// memory. This feature comes with the limitiation in how distant two such
// numbers can be. Circular numbers have a maximum. The maximum distance is
// half the maximum value. If a number that has the maximum value is
// increased by 1, it becomes 0. If a number that has the value of 0 is
// decreased by 1, it becomes the maximum value. By comparing two circular
// numbers it is not possible to tell how often they wrapped. Therefore these
// two numbers must come from the same domain in order to make sense of the
// camparison.
package circular

// Number represents a "circular number". A Number is immutable. All modification
// to a Number will result in a new instance of a Number.
type Number struct {
	max       uint32
	threshold uint32
	value     uint32
}

// New returns a new circular number with the value of x and the maximum of max.
func New(x, max uint32) Number {
	c := Number{
		value:     0,
		max:       max,
		threshold: max / 2,
	}

	if x > max {
		return c.Add(x)
	}

	c.value = x

	return c
}

// Val returns the current value of the number.
func (a Number) Val() uint32 {
	return a.value
}

// Equals returns whether two circular numbers have the same value.
func (a Number) Equals(b Number) bool {
	return a.value == b.value
}

// Distance returns the distance of two circular numbers.
func (a Number) Distance(b Number) uint32 {
	if a.Equals(b) {
		return 0
	}

	d := uint32(0)

	if a.value > b.value {
		d = a.value - b.value
	} else {
		d = b.value - a.value
	}

	if d >= a.threshold {
		d = a.max - d + 1
	}

	return d
}

// Lt returns whether the circular number is lower than the circular number b.
func (a Number) Lt(b Number) bool {
	if a.Equals(b) {
		return false
	}

	d := uint32(0)
	altb := false

	if a.value > b.value {
		d = a.value - b.value
	} else {
		d = b.value - a.value
		altb = true
	}

	if d < a.threshold {
		return altb
	}

	return !altb
}

// Lte returns whether the circular number is lower than or equal to the circular number b.
func (a Number) Lte(b Number) bool {
	if a.Equals(b) {
		return true
	}

	return a.Lt(b)
}

// Gt returns whether the circular number is greather than the circular number b.
func (a Number) Gt(b Number) bool {
	if a.Equals(b) {
		return false
	}

	d := uint32(0)
	agtb := false

	if a.value > b.value {
		d = a.value - b.value
		agtb = true
	} else {
		d = b.value - a.value
	}

	if d < a.threshold {
		return agtb
	}

	return !agtb
}

// Gte returns whether the circular number is greather than or equal to the circular number b.
func (a Number) Gte(b Number) bool {
	if a.Equals(b) {
		return true
	}

	return a.Gt(b)
}

// Inc returns a new circular number with a value that is increased by 1.
func (a Number) Inc() Number {
	b := a

	if b.value == b.max {
		b.value = 0
	} else {
		b.value++
	}

	return b
}

// Add returns a new circular number with a value that is increased by b.
func (a Number) Add(b uint32) Number {
	c := a
	x := c.max - c.value

	if b <= x {
		c.value += b
	} else {
		c.value = b - x - 1
	}

	return c
}

// Dec returns a new circular number with a value that is decreased by 1.
func (a Number) Dec() Number {
	b := a

	if b.value == 0 {
		b.value = b.max
	} else {
		b.value--
	}

	return b
}

// Sub returns a new circular number with a value that is decreased by b.
func (a Number) Sub(b uint32) Number {
	c := a

	if b <= c.value {
		c.value -= b
	} else {
		c.value = c.max - (b - c.value) + 1
	}

	return c
}
````

## File: CODE_OF_CONDUCT.md
````markdown
Contributor Code of Conduct
===========================

As contributors and maintainers of this project, we pledge to respect all
people who contribute through reporting issues, posting feature requests,
updating documentation, submitting pull requests or patches, and other
activities.

Examples of unacceptable behavior by participants include:
- Sexual language or imagery.
- Derogatory comments or personal attacks.
- Trolling, public or private harassment.
- Insults.
- Other unprofessional conduct.

Examples of unacceptable behavior by participants include sexual
language or imagery, derogatory comments or personal attacks, trolling, public
or private harassment, insults, or other unprofessional conduct.

Project maintainers have the right and responsibility to remove, edit, or reject 
comments, commits, code, wiki edits, issues, and other contributions that are not
aligned with this Code of Conduct. In addition, project maintainers who do not
follow the Code of Conduct may be removed from the project team.

This code of conduct applies both within a project and in public spaces when an
individual represents the project or its community.

Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported by opening an issue or contacting one or more of the project maintainers.

This Code of Conduct is adapted from the [Contributor
Covenant](https://contributor-covenant.org/), version 1.1.0, available at
[https://contributor-covenant.org/version/1/1/0/](https://contributor-covenant.org/version/1/1/0/)
````

## File: config_test.go
````go
package srt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	err := config.Validate()

	if err != nil {
		require.NoError(t, err, "Failed to verify the default configuration: %s", err)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	wantConfig := Config{
		Congestion:            "xxx",
		ConnectionTimeout:     42 * time.Second,
		DriftTracer:           false,
		EnforcedEncryption:    false,
		FC:                    42,
		GroupConnect:          true,
		GroupStabilityTimeout: 42 * time.Second,
		InputBW:               42,
		IPTOS:                 42,
		IPTTL:                 42,
		IPv6Only:              42,
		KMPreAnnounce:         42,
		KMRefreshRate:         42,
		Latency:               42 * time.Second,
		LossMaxTTL:            42,
		MaxBW:                 42,
		MessageAPI:            true,
		MinInputBW:            42,
		MSS:                   42,
		NAKReport:             false,
		OverheadBW:            42,
		PacketFilter:          "FEC",
		Passphrase:            "foobar",
		PayloadSize:           42,
		PBKeylen:              42,
		PeerIdleTimeout:       42 * time.Second,
		PeerLatency:           42 * time.Second,
		ReceiverBufferSize:    42,
		ReceiverLatency:       42 * time.Second,
		SendBufferSize:        42,
		SendDropDelay:         42 * time.Second,
		StreamId:              "foobaz",
		TooLatePacketDrop:     false,
		TransmissionType:      "yyy",
		TSBPDMode:             false,
		Logger:                nil,
	}

	url := wantConfig.MarshalURL("localhost:6000")

	config := Config{}
	config.UnmarshalURL(url)

	require.Equal(t, wantConfig, config)
}
````

## File: config.go
````go
package srt

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

const (
	UDP_HEADER_SIZE     = 28
	SRT_HEADER_SIZE     = 16
	MIN_MSS_SIZE        = 76
	MAX_MSS_SIZE        = 1500
	MIN_PAYLOAD_SIZE    = MIN_MSS_SIZE - UDP_HEADER_SIZE - SRT_HEADER_SIZE
	MAX_PAYLOAD_SIZE    = MAX_MSS_SIZE - UDP_HEADER_SIZE - SRT_HEADER_SIZE
	MIN_PASSPHRASE_SIZE = 10
	MAX_PASSPHRASE_SIZE = 80
	MAX_STREAMID_SIZE   = 512
	SRT_VERSION         = 0x010401
)

// Config is the configuration for a SRT connection
type Config struct {
	// Type of congestion control. 'live' or 'file'
	// SRTO_CONGESTION
	Congestion string

	// Connection timeout.
	// SRTO_CONNTIMEO
	ConnectionTimeout time.Duration

	// Enable drift tracer.
	// SRTO_DRIFTTRACER
	DriftTracer bool

	// Reject connection if parties set different passphrase.
	// SRTO_ENFORCEDENCRYPTION
	EnforcedEncryption bool

	// Flow control window size. Packets.
	// SRTO_FC
	FC uint32

	// Accept group connections.
	// SRTO_GROUPCONNECT
	GroupConnect bool

	// Group stability timeout.
	// SRTO_GROUPSTABTIMEO
	GroupStabilityTimeout time.Duration

	// Input bandwidth. Bytes.
	// SRTO_INPUTBW
	InputBW int64

	// IP socket type of service
	// SRTO_IPTOS
	IPTOS int

	// Defines IP socket "time to live" option.
	// SRTO_IPTTL
	IPTTL int

	// Allow only IPv6.
	// SRTO_IPV6ONLY
	IPv6Only int

	// Duration of Stream Encryption key switchover. Packets.
	// SRTO_KMPREANNOUNCE
	KMPreAnnounce uint64

	// Stream encryption key refresh rate. Packets.
	// SRTO_KMREFRESHRATE
	KMRefreshRate uint64

	// Defines the maximum accepted transmission latency.
	// SRTO_LATENCY
	Latency time.Duration

	// Packet reorder tolerance.
	// SRTO_LOSSMAXTTL
	LossMaxTTL uint32

	// Bandwidth limit in bytes/s.
	// SRTO_MAXBW
	MaxBW int64

	// Enable SRT message mode.
	// SRTO_MESSAGEAPI
	MessageAPI bool

	// Minimum input bandwidth
	// This option is effective only if both SRTO_MAXBW and SRTO_INPUTBW are set to 0. It controls the minimum allowed value of the input bitrate estimate.
	// SRTO_MININPUTBW
	MinInputBW int64

	// Minimum SRT library version of a peer.
	// SRTO_MINVERSION
	MinVersion uint32

	// MTU size
	// SRTO_MSS
	MSS uint32

	// Enable periodic NAK reports
	// SRTO_NAKREPORT
	NAKReport bool

	// Limit bandwidth overhead, percents
	// SRTO_OHEADBW
	OverheadBW int64

	// Set up the packet filter.
	// SRTO_PACKETFILTER
	PacketFilter string

	// Password for the encrypted transmission.
	// SRTO_PASSPHRASE
	Passphrase string

	// Maximum payload size. Bytes.
	// SRTO_PAYLOADSIZE
	PayloadSize uint32

	// Crypto key length in bytes.
	// SRTO_PBKEYLEN
	PBKeylen int

	// Peer idle timeout.
	// SRTO_PEERIDLETIMEO
	PeerIdleTimeout time.Duration

	// Minimum receiver latency to be requested by sender.
	// SRTO_PEERLATENCY
	PeerLatency time.Duration

	// Receiver buffer size. Bytes.
	// SRTO_RCVBUF
	ReceiverBufferSize uint32

	// Receiver-side latency.
	// SRTO_RCVLATENCY
	ReceiverLatency time.Duration

	// Sender buffer size. Bytes.
	// SRTO_SNDBUF
	SendBufferSize uint32

	// Sender's delay before dropping packets.
	// SRTO_SNDDROPDELAY
	SendDropDelay time.Duration

	// Stream ID (settable in caller mode only, visible on the listener peer)
	// SRTO_STREAMID
	StreamId string

	// Drop too late packets.
	// SRTO_TLPKTDROP
	TooLatePacketDrop bool

	// Transmission type. 'live' or 'file'.
	// SRTO_TRANSTYPE
	TransmissionType string

	// Timestamp-based packet delivery mode.
	// SRTO_TSBPDMODE
	TSBPDMode bool

	// An implementation of the Logger interface
	Logger Logger

	// if a new IP starts sending data on an existing socket id, allow it
	AllowPeerIpChange bool
}

// DefaultConfig is the default configuration for a SRT connection
// if no individual configuration has been provided.
var defaultConfig Config = Config{
	Congestion:            "live",
	ConnectionTimeout:     3 * time.Second,
	DriftTracer:           true,
	EnforcedEncryption:    true,
	FC:                    25600,
	GroupConnect:          false,
	GroupStabilityTimeout: 0,
	InputBW:               0,
	IPTOS:                 0,
	IPTTL:                 0,
	IPv6Only:              -1,
	KMPreAnnounce:         1 << 12,
	KMRefreshRate:         1 << 24,
	Latency:               -1,
	LossMaxTTL:            0,
	MaxBW:                 -1,
	MessageAPI:            false,
	MinVersion:            SRT_VERSION,
	MSS:                   MAX_MSS_SIZE,
	NAKReport:             true,
	OverheadBW:            25,
	PacketFilter:          "",
	Passphrase:            "",
	PayloadSize:           MAX_PAYLOAD_SIZE,
	PBKeylen:              16,
	PeerIdleTimeout:       5 * time.Second,
	PeerLatency:           120 * time.Millisecond,
	ReceiverBufferSize:    0,
	ReceiverLatency:       120 * time.Millisecond,
	SendBufferSize:        0,
	SendDropDelay:         1 * time.Second,
	StreamId:              "",
	TooLatePacketDrop:     true,
	TransmissionType:      "live",
	TSBPDMode:             true,
	AllowPeerIpChange:     false,
}

// DefaultConfig returns the default configuration for Dial and Listen.
func DefaultConfig() Config {
	return defaultConfig
}

// UnmarshalURL takes a SRT URL and parses out the configuration. A SRT URL is
// srt://[host]:[port]?[key1]=[value1]&[key2]=[value2]... It returns the host:port
// of the URL.
func (c *Config) UnmarshalURL(srturl string) (string, error) {
	u, err := url.Parse(srturl)
	if err != nil {
		return "", err
	}

	if u.Scheme != "srt" {
		return "", fmt.Errorf("the URL doesn't seem to be an srt:// URL")
	}

	return u.Host, c.UnmarshalQuery(u.RawQuery)
}

// UnmarshalQuery parses a query string and interprets it as a configuration
// for a SRT connection. The key in each key/value pair corresponds to the
// respective field in the Config type, but with only lower case letters. Bool
// values can be represented as "true"/"false", "on"/"off", "yes"/"no", or "0"/"1".
func (c *Config) UnmarshalQuery(query string) error {
	v, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	// https://github.com/Haivision/srt/blob/master/docs/apps/srt-live-transmit.md

	if s := v.Get("congestion"); len(s) != 0 {
		c.Congestion = s
	}

	if s := v.Get("conntimeo"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.ConnectionTimeout = time.Duration(d) * time.Millisecond
		}
	}

	if s := v.Get("drifttracer"); len(s) != 0 {
		switch s {
		case "yes", "on", "true", "1":
			c.DriftTracer = true
		case "no", "off", "false", "0":
			c.DriftTracer = false
		}
	}

	if s := v.Get("enforcedencryption"); len(s) != 0 {
		switch s {
		case "yes", "on", "true", "1":
			c.EnforcedEncryption = true
		case "no", "off", "false", "0":
			c.EnforcedEncryption = false
		}
	}

	if s := v.Get("fc"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 32); err == nil {
			c.FC = uint32(d)
		}
	}

	if s := v.Get("groupconnect"); len(s) != 0 {
		switch s {
		case "yes", "on", "true", "1":
			c.GroupConnect = true
		case "no", "off", "false", "0":
			c.GroupConnect = false
		}
	}

	if s := v.Get("groupstabtimeo"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.GroupStabilityTimeout = time.Duration(d) * time.Millisecond
		}
	}

	if s := v.Get("inputbw"); len(s) != 0 {
		if d, err := strconv.ParseInt(s, 10, 64); err == nil {
			c.InputBW = d
		}
	}

	if s := v.Get("iptos"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.IPTOS = d
		}
	}

	if s := v.Get("ipttl"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.IPTTL = d
		}
	}

	if s := v.Get("ipv6only"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.IPv6Only = d
		}
	}

	if s := v.Get("kmpreannounce"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 64); err == nil {
			c.KMPreAnnounce = d
		}
	}

	if s := v.Get("kmrefreshrate"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 64); err == nil {
			c.KMRefreshRate = d
		}
	}

	if s := v.Get("latency"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.Latency = time.Duration(d) * time.Millisecond
		}
	}

	if s := v.Get("lossmaxttl"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 32); err == nil {
			c.LossMaxTTL = uint32(d)
		}
	}

	if s := v.Get("maxbw"); len(s) != 0 {
		if d, err := strconv.ParseInt(s, 10, 64); err == nil {
			c.MaxBW = d
		}
	}

	if s := v.Get("mininputbw"); len(s) != 0 {
		if d, err := strconv.ParseInt(s, 10, 64); err == nil {
			c.MinInputBW = d
		}
	}

	if s := v.Get("messageapi"); len(s) != 0 {
		switch s {
		case "yes", "on", "true", "1":
			c.MessageAPI = true
		case "no", "off", "false", "0":
			c.MessageAPI = false
		}
	}

	// minversion is ignored

	if s := v.Get("mss"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 32); err == nil {
			c.MSS = uint32(d)
		}
	}

	if s := v.Get("nakreport"); len(s) != 0 {
		switch s {
		case "yes", "on", "true", "1":
			c.NAKReport = true
		case "no", "off", "false", "0":
			c.NAKReport = false
		}
	}

	if s := v.Get("oheadbw"); len(s) != 0 {
		if d, err := strconv.ParseInt(s, 10, 64); err == nil {
			c.OverheadBW = d
		}
	}

	if s := v.Get("packetfilter"); len(s) != 0 {
		c.PacketFilter = s
	}

	if s := v.Get("passphrase"); len(s) != 0 {
		c.Passphrase = s
	}

	if s := v.Get("payloadsize"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 32); err == nil {
			c.PayloadSize = uint32(d)
		}
	}

	if s := v.Get("pbkeylen"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.PBKeylen = d
		}
	}

	if s := v.Get("peeridletimeo"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.PeerIdleTimeout = time.Duration(d) * time.Millisecond
		}
	}

	if s := v.Get("peerlatency"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.PeerLatency = time.Duration(d) * time.Millisecond
		}
	}

	if s := v.Get("rcvbuf"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 32); err == nil {
			c.ReceiverBufferSize = uint32(d)
		}
	}

	if s := v.Get("rcvlatency"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.ReceiverLatency = time.Duration(d) * time.Millisecond
		}
	}

	// retransmitalgo not implemented (there's only one)

	if s := v.Get("sndbuf"); len(s) != 0 {
		if d, err := strconv.ParseUint(s, 10, 32); err == nil {
			c.SendBufferSize = uint32(d)
		}
	}

	if s := v.Get("snddropdelay"); len(s) != 0 {
		if d, err := strconv.Atoi(s); err == nil {
			c.SendDropDelay = time.Duration(d) * time.Millisecond
		}
	}

	if s := v.Get("streamid"); len(s) != 0 {
		c.StreamId = s
	}

	if s := v.Get("tlpktdrop"); len(s) != 0 {
		switch s {
		case "yes", "on", "true", "1":
			c.TooLatePacketDrop = true
		case "no", "off", "false", "0":
			c.TooLatePacketDrop = false
		}
	}

	if s := v.Get("transtype"); len(s) != 0 {
		c.TransmissionType = s
	}

	if s := v.Get("tsbpdmode"); len(s) != 0 {
		switch s {
		case "yes", "on", "true", "1":
			c.TSBPDMode = true
		case "no", "off", "false", "0":
			c.TSBPDMode = false
		}
	}

	return nil
}

// MarshalURL returns the SRT URL for this config and the given address (host:port).
func (c *Config) MarshalURL(address string) string {
	return "srt://" + address + "?" + c.MarshalQuery()
}

// MarshalQuery returns the corresponding query string for a configuration.
func (c *Config) MarshalQuery() string {
	q := url.Values{}

	if c.Congestion != defaultConfig.Congestion {
		q.Set("congestion", c.Congestion)
	}

	if c.ConnectionTimeout != defaultConfig.ConnectionTimeout {
		q.Set("conntimeo", strconv.FormatInt(c.ConnectionTimeout.Milliseconds(), 10))
	}

	if c.DriftTracer != defaultConfig.DriftTracer {
		q.Set("drifttracer", strconv.FormatBool(c.DriftTracer))
	}

	if c.EnforcedEncryption != defaultConfig.EnforcedEncryption {
		q.Set("enforcedencryption", strconv.FormatBool(c.EnforcedEncryption))
	}

	if c.FC != defaultConfig.FC {
		q.Set("fc", strconv.FormatUint(uint64(c.FC), 10))
	}

	if c.GroupConnect != defaultConfig.GroupConnect {
		q.Set("groupconnect", strconv.FormatBool(c.GroupConnect))
	}

	if c.GroupStabilityTimeout != defaultConfig.GroupStabilityTimeout {
		q.Set("groupstabtimeo", strconv.FormatInt(c.GroupStabilityTimeout.Milliseconds(), 10))
	}

	if c.InputBW != defaultConfig.InputBW {
		q.Set("inputbw", strconv.FormatInt(c.InputBW, 10))
	}

	if c.IPTOS != defaultConfig.IPTOS {
		q.Set("iptos", strconv.FormatInt(int64(c.IPTOS), 10))
	}

	if c.IPTTL != defaultConfig.IPTTL {
		q.Set("ipttl", strconv.FormatInt(int64(c.IPTTL), 10))
	}

	if c.IPv6Only != defaultConfig.IPv6Only {
		q.Set("ipv6only", strconv.FormatInt(int64(c.IPv6Only), 10))
	}

	if len(c.Passphrase) != 0 {
		if c.KMPreAnnounce != defaultConfig.KMPreAnnounce {
			q.Set("kmpreannounce", strconv.FormatUint(c.KMPreAnnounce, 10))
		}

		if c.KMRefreshRate != defaultConfig.KMRefreshRate {
			q.Set("kmrefreshrate", strconv.FormatUint(c.KMRefreshRate, 10))
		}
	}

	if c.Latency != defaultConfig.Latency {
		q.Set("latency", strconv.FormatInt(c.Latency.Milliseconds(), 10))
	}

	if c.LossMaxTTL != defaultConfig.LossMaxTTL {
		q.Set("lossmaxttl", strconv.FormatInt(int64(c.LossMaxTTL), 10))
	}

	if c.MaxBW != defaultConfig.MaxBW {
		q.Set("maxbw", strconv.FormatInt(c.MaxBW, 10))
	}

	if c.MinInputBW != defaultConfig.InputBW {
		q.Set("mininputbw", strconv.FormatInt(c.MinInputBW, 10))
	}

	if c.MessageAPI != defaultConfig.MessageAPI {
		q.Set("messageapi", strconv.FormatBool(c.MessageAPI))
	}

	if c.MSS != defaultConfig.MSS {
		q.Set("mss", strconv.FormatUint(uint64(c.MSS), 10))
	}

	if c.NAKReport != defaultConfig.NAKReport {
		q.Set("nakreport", strconv.FormatBool(c.NAKReport))
	}

	if c.OverheadBW != defaultConfig.OverheadBW {
		q.Set("oheadbw", strconv.FormatInt(c.OverheadBW, 10))
	}

	if c.PacketFilter != defaultConfig.PacketFilter {
		q.Set("packetfilter", c.PacketFilter)
	}

	if len(c.Passphrase) != 0 {
		q.Set("passphrase", c.Passphrase)
	}

	if c.PayloadSize != defaultConfig.PayloadSize {
		q.Set("payloadsize", strconv.FormatUint(uint64(c.PayloadSize), 10))
	}

	if c.PBKeylen != defaultConfig.PBKeylen {
		q.Set("pbkeylen", strconv.FormatInt(int64(c.PBKeylen), 10))
	}

	if c.PeerIdleTimeout != defaultConfig.PeerIdleTimeout {
		q.Set("peeridletimeo", strconv.FormatInt(c.PeerIdleTimeout.Milliseconds(), 10))
	}

	if c.PeerLatency != defaultConfig.PeerLatency {
		q.Set("peerlatency", strconv.FormatInt(c.PeerLatency.Milliseconds(), 10))
	}

	if c.ReceiverBufferSize != defaultConfig.ReceiverBufferSize {
		q.Set("rcvbuf", strconv.FormatInt(int64(c.ReceiverBufferSize), 10))
	}

	if c.ReceiverLatency != defaultConfig.ReceiverLatency {
		q.Set("rcvlatency", strconv.FormatInt(c.ReceiverLatency.Milliseconds(), 10))
	}

	if c.SendBufferSize != defaultConfig.SendBufferSize {
		q.Set("sndbuf", strconv.FormatInt(int64(c.SendBufferSize), 10))
	}

	if c.SendDropDelay != defaultConfig.SendDropDelay {
		q.Set("snddropdelay", strconv.FormatInt(c.SendDropDelay.Milliseconds(), 10))
	}

	if len(c.StreamId) != 0 {
		q.Set("streamid", c.StreamId)
	}

	if c.TooLatePacketDrop != defaultConfig.TooLatePacketDrop {
		q.Set("tlpktdrop", strconv.FormatBool(c.TooLatePacketDrop))
	}

	if c.TransmissionType != defaultConfig.TransmissionType {
		q.Set("transtype", c.TransmissionType)
	}

	if c.TSBPDMode != defaultConfig.TSBPDMode {
		q.Set("tsbpdmode", strconv.FormatBool(c.TSBPDMode))
	}

	return q.Encode()
}

// Validate validates a configuration, returns an error if a field
// has an invalid value.
func (c *Config) Validate() error {
	if c.TransmissionType != "live" {
		return fmt.Errorf("config: TransmissionType must be 'live'")
	}

	c.Congestion = "live"
	c.TSBPDMode = true

	if c.Congestion != "live" {
		return fmt.Errorf("config: Congestion mode must be 'live'")
	}

	if c.ConnectionTimeout <= 0 {
		return fmt.Errorf("config: ConnectionTimeout must be greater than 0")
	}

	if c.GroupConnect {
		return fmt.Errorf("config: GroupConnect is not supported")
	}

	if c.IPTOS > 0 && c.IPTOS > 255 {
		return fmt.Errorf("config: IPTOS must be lower than 255")
	}

	if c.IPTTL > 0 && c.IPTTL > 255 {
		return fmt.Errorf("config: IPTTL must be between 1 and 255")
	}

	if c.IPv6Only > 0 {
		return fmt.Errorf("config: IPv6Only is not supported")
	}

	if c.KMRefreshRate != 0 {
		if c.KMPreAnnounce < 1 || c.KMPreAnnounce > c.KMRefreshRate/2 {
			return fmt.Errorf("config: KMPreAnnounce must be greater than 1 and smaller than KMRefreshRate/2")
		}
	}

	if c.Latency >= 0 {
		c.PeerLatency = c.Latency
		c.ReceiverLatency = c.Latency
	}

	if c.MinVersion != SRT_VERSION {
		return fmt.Errorf("config: MinVersion must be %#06x", SRT_VERSION)
	}

	if c.MSS < MIN_MSS_SIZE || c.MSS > MAX_MSS_SIZE {
		return fmt.Errorf("config: MSS must be between %d and %d (both inclusive)", MIN_MSS_SIZE, MAX_MSS_SIZE)
	}

	if c.OverheadBW < 10 || c.OverheadBW > 100 {
		return fmt.Errorf("config: OverheadBW must be between 10 and 100")
	}

	if len(c.PacketFilter) != 0 {
		return fmt.Errorf("config: PacketFilter are not supported")
	}

	if len(c.Passphrase) != 0 {
		if len(c.Passphrase) < MIN_PASSPHRASE_SIZE || len(c.Passphrase) > MAX_PASSPHRASE_SIZE {
			return fmt.Errorf("config: Passphrase must be between %d and %d bytes long", MIN_PASSPHRASE_SIZE, MAX_PASSPHRASE_SIZE)
		}
	}

	if c.PayloadSize < MIN_PAYLOAD_SIZE || c.PayloadSize > MAX_PAYLOAD_SIZE {
		return fmt.Errorf("config: PayloadSize must be between %d and %d (both inclusive)", MIN_PAYLOAD_SIZE, MAX_PAYLOAD_SIZE)
	}

	if c.PayloadSize > c.MSS-uint32(SRT_HEADER_SIZE+UDP_HEADER_SIZE) {
		return fmt.Errorf("config: PayloadSize must not be larger than %d (MSS - %d)", c.MSS-uint32(SRT_HEADER_SIZE+UDP_HEADER_SIZE), SRT_HEADER_SIZE-UDP_HEADER_SIZE)
	}

	if c.PBKeylen != 16 && c.PBKeylen != 24 && c.PBKeylen != 32 {
		return fmt.Errorf("config: PBKeylen must be 16, 24, or 32 bytes")
	}

	if c.PeerLatency < 0 {
		return fmt.Errorf("config: PeerLatency must be greater than 0")
	}

	if c.ReceiverLatency < 0 {
		return fmt.Errorf("config: ReceiverLatency must be greater than 0")
	}

	if c.SendDropDelay < 0 {
		return fmt.Errorf("config: SendDropDelay must be greater than 0")
	}

	if len(c.StreamId) > MAX_STREAMID_SIZE {
		return fmt.Errorf("config: StreamId must be shorter than or equal to %d bytes", MAX_STREAMID_SIZE)
	}

	if c.TransmissionType != "live" {
		return fmt.Errorf("config: TransmissionType must be 'live'")
	}

	if !c.TSBPDMode {
		return fmt.Errorf("config: TSBPDMode must be enabled")
	}

	return nil
}
````

## File: congestion/congestion.go
````go
// Package congestions provides interfaces and types congestion control implementations for SRT
package congestion

import (
	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/packet"
)

// Sender is the sending part of the congestion control
type Sender interface {
	// Stats returns sender statistics.
	Stats() SendStats

	// Flush flushes all queued packages.
	Flush()

	// Push pushes a packet to be send on the sender queue.
	Push(p packet.Packet)

	// Tick gets called from a connection in order to proceed with the queued packets. The provided value for
	// now is corresponds to the timestamps in the queued packets. Those timestamps are the microseconds
	// since the start of the connection.
	Tick(now uint64)

	// ACK gets called when a sequence number has been confirmed from a receiver.
	ACK(sequenceNumber circular.Number)

	// NAK get called when packets with the listed sequence number should be resend.
	NAK(sequenceNumbers []circular.Number)

	// SetDropThreshold sets the threshold in microseconds for when to drop too late packages from the queue.
	SetDropThreshold(threshold uint64)
}

// Receiver is the receiving part of the congestion control
type Receiver interface {
	// Stats returns receiver statistics.
	Stats() ReceiveStats

	// PacketRate returns the current packets and bytes per second, and the capacity of the link.
	PacketRate() (pps, bps, capacity float64)

	// Flush flushes all queued packages.
	Flush()

	// Push pushed a recieved packet to the receiver queue.
	Push(pkt packet.Packet)

	// Tick gets called from a connection in order to proceed with queued packets. The provided value for
	// now is corresponds to the timestamps in the queued packets. Those timestamps are the microseconds
	// since the start of the connection.
	Tick(now uint64)

	// SetNAKInterval sets the interval between two periodic NAK messages to the sender in microseconds.
	SetNAKInterval(nakInterval uint64)
}

// SendStats are collected statistics from a sender
type SendStats struct {
	Pkt  uint64 // Sent packets in total
	Byte uint64 // Sent bytes in total

	PktUnique  uint64
	ByteUnique uint64

	PktLoss  uint64
	ByteLoss uint64

	PktRetrans  uint64
	ByteRetrans uint64

	UsSndDuration uint64 // microseconds

	PktDrop  uint64
	ByteDrop uint64

	// instantaneous
	PktBuf  uint64
	ByteBuf uint64
	MsBuf   uint64

	PktFlightSize uint64

	UsPktSndPeriod float64 // microseconds
	BytePayload    uint64

	MbpsEstimatedInputBandwidth float64
	MbpsEstimatedSentBandwidth  float64

	PktLossRate float64
}

// ReceiveStats are collected statistics from a reciever
type ReceiveStats struct {
	Pkt  uint64
	Byte uint64

	PktUnique  uint64
	ByteUnique uint64

	PktLoss  uint64
	ByteLoss uint64

	PktRetrans  uint64
	ByteRetrans uint64

	PktBelated  uint64
	ByteBelated uint64

	PktDrop  uint64
	ByteDrop uint64

	// instantaneous
	PktBuf  uint64
	ByteBuf uint64
	MsBuf   uint64

	BytePayload uint64

	MbpsEstimatedRecvBandwidth float64
	MbpsEstimatedLinkCapacity  float64

	PktLossRate float64
}
````

## File: congestion/live/doc.go
````go
// Package live provides implementations of the Sender and Receiver interfaces for live congestion control
package live
````

## File: congestion/live/fake.go
````go
package live

import (
	"sync"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/congestion"
	"github.com/datarhei/gosrt/packet"
)

type fakeLiveReceive struct {
	maxSeenSequenceNumber       circular.Number
	lastACKSequenceNumber       circular.Number
	lastDeliveredSequenceNumber circular.Number

	nPackets uint

	periodicACKInterval uint64 // config
	periodicNAKInterval uint64 // config

	lastPeriodicACK uint64

	avgPayloadSize float64 // bytes

	rate struct {
		last   time.Time
		period time.Duration

		packets uint64
		bytes   uint64

		pps float64
		bps float64
	}

	sendACK func(seq circular.Number, light bool)
	sendNAK func(list []circular.Number)
	deliver func(p packet.Packet)

	lock sync.RWMutex
}

func NewFakeLiveReceive(config ReceiveConfig) congestion.Receiver {
	r := &fakeLiveReceive{
		maxSeenSequenceNumber:       config.InitialSequenceNumber.Dec(),
		lastACKSequenceNumber:       config.InitialSequenceNumber.Dec(),
		lastDeliveredSequenceNumber: config.InitialSequenceNumber.Dec(),

		periodicACKInterval: config.PeriodicACKInterval,
		periodicNAKInterval: config.PeriodicNAKInterval,

		avgPayloadSize: 1456, //  5.1.2. SRT's Default LiveCC Algorithm

		sendACK: config.OnSendACK,
		sendNAK: config.OnSendNAK,
		deliver: config.OnDeliver,
	}

	if r.sendACK == nil {
		r.sendACK = func(seq circular.Number, light bool) {}
	}

	if r.sendNAK == nil {
		r.sendNAK = func(list []circular.Number) {}
	}

	if r.deliver == nil {
		r.deliver = func(p packet.Packet) {}
	}

	r.rate.last = time.Now()
	r.rate.period = time.Second

	return r
}

func (r *fakeLiveReceive) Stats() congestion.ReceiveStats { return congestion.ReceiveStats{} }
func (r *fakeLiveReceive) PacketRate() (pps, bps, capacity float64) {
	r.lock.Lock()
	defer r.lock.Unlock()

	tdiff := time.Since(r.rate.last)

	if tdiff < r.rate.period {
		pps = r.rate.pps
		bps = r.rate.bps

		return
	}

	r.rate.pps = float64(r.rate.packets) / tdiff.Seconds()
	r.rate.bps = float64(r.rate.bytes) / tdiff.Seconds()

	r.rate.packets, r.rate.bytes = 0, 0
	r.rate.last = time.Now()

	pps = r.rate.pps
	bps = r.rate.bps

	return
}

func (r *fakeLiveReceive) Flush() {}

func (r *fakeLiveReceive) Push(pkt packet.Packet) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if pkt == nil {
		return
	}

	r.nPackets++

	pktLen := pkt.Len()

	r.rate.packets++
	r.rate.bytes += pktLen

	//  5.1.2. SRT's Default LiveCC Algorithm
	r.avgPayloadSize = 0.875*r.avgPayloadSize + 0.125*float64(pktLen)

	if pkt.Header().PacketSequenceNumber.Lte(r.lastDeliveredSequenceNumber) {
		// Too old, because up until r.lastDeliveredSequenceNumber, we already delivered
		return
	}

	if pkt.Header().PacketSequenceNumber.Lt(r.lastACKSequenceNumber) {
		// Already acknowledged, ignoring
		return
	}

	if pkt.Header().PacketSequenceNumber.Lte(r.maxSeenSequenceNumber) {
		return
	}

	r.maxSeenSequenceNumber = pkt.Header().PacketSequenceNumber
}

func (r *fakeLiveReceive) periodicACK(now uint64) (ok bool, sequenceNumber circular.Number, lite bool) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// 4.8.1. Packet Acknowledgement (ACKs, ACKACKs)
	if now-r.lastPeriodicACK < r.periodicACKInterval {
		if r.nPackets >= 64 {
			lite = true // Send light ACK
		} else {
			return
		}
	}

	ok = true
	sequenceNumber = r.maxSeenSequenceNumber.Inc()

	r.lastACKSequenceNumber = r.maxSeenSequenceNumber

	r.lastPeriodicACK = now
	r.nPackets = 0

	return
}

func (r *fakeLiveReceive) Tick(now uint64) {
	if ok, sequenceNumber, lite := r.periodicACK(now); ok {
		r.sendACK(sequenceNumber, lite)
	}

	// Deliver packets whose PktTsbpdTime is ripe
	r.lock.Lock()
	defer r.lock.Unlock()

	r.lastDeliveredSequenceNumber = r.lastACKSequenceNumber
}

func (r *fakeLiveReceive) SetNAKInterval(nakInterval uint64) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.periodicNAKInterval = nakInterval
}
````

## File: congestion/live/receive_test.go
````go
package live

import (
	"net"
	"testing"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/packet"
	"github.com/stretchr/testify/require"
)

func mockLiveRecv(onSendACK func(seq circular.Number, light bool), onSendNAK func(list []circular.Number), onDeliver func(p packet.Packet)) *receiver {
	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnSendACK:             onSendACK,
		OnSendNAK:             onSendNAK,
		OnDeliver:             onDeliver,
		LossMaxTTL:            0,
	})

	return recv.(*receiver)
}

func TestRecvSequence(t *testing.T) {
	nACK := 0
	nNAK := 0
	numbers := []uint32{}
	recv := mockLiveRecv(
		func(seq circular.Number, light bool) {
			nACK++
		},
		func(list []circular.Number) {
			nNAK++
		},
		func(p packet.Packet) {
			numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		},
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	require.Equal(t, 0, nACK)
	require.Equal(t, 0, nNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(0), recv.lastACKSequenceNumber.Inc().Val())

	recv.Tick(1)

	require.Equal(t, 0, nACK)
	require.Equal(t, 0, nNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(0), recv.lastACKSequenceNumber.Inc().Val())

	recv.Tick(10) // ACK period

	require.Equal(t, 1, nACK)
	require.Equal(t, 0, nNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(9), recv.lastACKSequenceNumber.Val())

	require.Exactly(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, numbers)
}

func TestRecvTSBPD(t *testing.T) {
	numbers := []uint32{}
	recv := mockLiveRecv(
		nil,
		nil,
		func(p packet.Packet) {
			numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		},
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 20 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(0), recv.lastACKSequenceNumber.Inc().Val())

	recv.Tick(10) // ACK period

	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(19), recv.lastACKSequenceNumber.Val())

	require.Exactly(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, numbers)
}

func TestRecvNAK(t *testing.T) {
	seqACK := uint32(0)
	seqNAK := []uint32{}
	numbers := []uint32{}
	recv := mockLiveRecv(
		func(seq circular.Number, light bool) {
			seqACK = seq.Val()
		},
		func(list []circular.Number) {
			seqNAK = []uint32{}
			for _, sn := range list {
				seqNAK = append(seqNAK, sn.Val())
			}
		},
		func(p packet.Packet) {
			numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		},
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 5 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(0), seqACK)
	require.Equal(t, []uint32{}, seqNAK)
	require.Equal(t, uint32(4), recv.maxSeenSequenceNumber.Val())

	for i := 7; i < 10; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(0), seqACK)
	require.Equal(t, []uint32{5, 6}, seqNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())

	recv.Tick(10) // ACK period

	require.Equal(t, uint32(10), seqACK)
	require.Equal(t, []uint32{5, 6}, seqNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
}

func TestRecvPeriodicNAK(t *testing.T) {
	seqACK := uint32(0)
	seqNAK := []uint32{}
	numbers := []uint32{}
	recv := mockLiveRecv(
		func(seq circular.Number, light bool) {
			seqACK = seq.Val()
		},
		func(list []circular.Number) {
			seqNAK = []uint32{}
			for _, sn := range list {
				seqNAK = append(seqNAK, sn.Val())
			}
		},
		func(p packet.Packet) {
			numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		},
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 5 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(50 + i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(0), seqACK)
	require.Equal(t, []uint32{}, seqNAK)
	require.Equal(t, uint32(4), recv.maxSeenSequenceNumber.Val())

	for i := 7; i < 10; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(50 + i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(0), seqACK)
	require.Equal(t, []uint32{5, 6}, seqNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())

	recv.Tick(10) // ACK period

	require.Equal(t, uint32(5), seqACK)
	require.Equal(t, []uint32{5, 6}, seqNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())

	recv.Tick(20) // ACK period, NAK period

	require.Equal(t, uint32(5), seqACK)
	require.Equal(t, []uint32{5, 6}, seqNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
}

func TestRecvACK(t *testing.T) {
	seqACK := uint32(0)
	seqNAK := []uint32{}
	numbers := []uint32{}
	recv := mockLiveRecv(
		func(seq circular.Number, light bool) {
			seqACK = seq.Val()
		},
		func(list []circular.Number) {
			seqNAK = []uint32{}
			for _, sn := range list {
				seqNAK = append(seqNAK, sn.Val())
			}
		},
		func(p packet.Packet) {
			numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		},
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 5 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(10 + i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(0), seqACK)
	require.Equal(t, []uint32{}, seqNAK)
	require.Equal(t, uint32(4), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(0), recv.lastACKSequenceNumber.Inc().Val())
	require.Equal(t, uint32(0), recv.lastDeliveredSequenceNumber.Inc().Val())
	require.Exactly(t, []uint32{}, numbers)

	for i := 7; i < 10; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(30 + i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(0), seqACK)
	require.Equal(t, []uint32{5, 6}, seqNAK)
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(0), recv.lastACKSequenceNumber.Inc().Val())
	require.Equal(t, uint32(0), recv.lastDeliveredSequenceNumber.Inc().Val())
	require.Exactly(t, []uint32{}, numbers)

	for i := 15; i < 20; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(30 + i + 1)

		recv.Push(p)
	}

	require.Equal(t, uint32(0), seqACK)
	require.Equal(t, []uint32{10, 14}, seqNAK)
	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(0), recv.lastACKSequenceNumber.Inc().Val())
	require.Equal(t, uint32(0), recv.lastDeliveredSequenceNumber.Inc().Val())
	require.Exactly(t, []uint32{}, numbers)

	recv.Tick(10)

	require.Equal(t, uint32(5), seqACK)
	require.Equal(t, []uint32{10, 14}, seqNAK)
	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(5), recv.lastACKSequenceNumber.Inc().Val())
	require.Equal(t, uint32(0), recv.lastDeliveredSequenceNumber.Inc().Val())
	require.Exactly(t, []uint32{}, numbers)

	recv.Tick(20)

	require.Equal(t, uint32(5), seqACK)
	require.Equal(t, []uint32{5, 6, 10, 14}, seqNAK)
	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(5), recv.lastACKSequenceNumber.Inc().Val())
	require.Equal(t, uint32(5), recv.lastDeliveredSequenceNumber.Inc().Val())
	require.Exactly(t, []uint32{0, 1, 2, 3, 4}, numbers)

	recv.Tick(30)

	require.Equal(t, uint32(5), seqACK)
	require.Equal(t, []uint32{5, 6, 10, 14}, seqNAK)
	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(5), recv.lastACKSequenceNumber.Inc().Val())
	require.Equal(t, uint32(5), recv.lastDeliveredSequenceNumber.Inc().Val())
	require.Exactly(t, []uint32{0, 1, 2, 3, 4}, numbers)

	for i := 5; i < 7; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(30 + i + 1)

		recv.Push(p)
	}

	recv.Tick(40)

	require.Equal(t, uint32(10), seqACK)
	require.Equal(t, []uint32{10, 14}, seqNAK)
	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint32(10), recv.lastACKSequenceNumber.Inc().Val())
	require.Equal(t, uint32(10), recv.lastDeliveredSequenceNumber.Inc().Val())
	require.Exactly(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, numbers)
}

func TestRecvDropTooLate(t *testing.T) {
	recv := mockLiveRecv(
		nil,
		nil,
		nil,
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	recv.Tick(10) // ACK period

	stats := recv.Stats()

	require.Equal(t, uint32(9), recv.lastACKSequenceNumber.Val())
	require.Equal(t, uint32(9), recv.lastDeliveredSequenceNumber.Val())
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint64(0), stats.PktDrop)

	p := packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(uint32(3), packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = uint64(4)

	recv.Push(p)

	stats = recv.Stats()

	require.Equal(t, uint64(1), stats.PktDrop)
}

func TestRecvDropAlreadyACK(t *testing.T) {
	recv := mockLiveRecv(
		nil,
		nil,
		nil,
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 5 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	for i := 5; i < 10; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(10 + i + 1)

		recv.Push(p)
	}

	recv.Tick(10) // ACK period

	stats := recv.Stats()

	require.Equal(t, uint32(9), recv.lastACKSequenceNumber.Val())
	require.Equal(t, uint32(4), recv.lastDeliveredSequenceNumber.Val())
	require.Equal(t, uint32(9), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint64(0), stats.PktDrop)

	p := packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(uint32(6), packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = uint64(7)

	recv.Push(p)

	stats = recv.Stats()

	require.Equal(t, uint64(1), stats.PktDrop)
}

func TestRecvDropAlreadyRecvNoACK(t *testing.T) {
	recv := mockLiveRecv(
		nil,
		nil,
		nil,
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 5 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	for i := 5; i < 10; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(10 + i + 1)

		recv.Push(p)
	}

	recv.Tick(10) // ACK period

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(10+i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(20 + i + 1)

		recv.Push(p)
	}

	stats := recv.Stats()

	require.Equal(t, uint32(9), recv.lastACKSequenceNumber.Val())
	require.Equal(t, uint32(4), recv.lastDeliveredSequenceNumber.Val())
	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())
	require.Equal(t, uint64(0), stats.PktDrop)

	p := packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(uint32(15), packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = uint64(20 + 6)

	recv.Push(p)

	stats = recv.Stats()

	require.Equal(t, uint64(1), stats.PktDrop)
}

func TestRecvFlush(t *testing.T) {
	recv := mockLiveRecv(
		nil,
		nil,
		nil,
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	require.Equal(t, 10, recv.packetList.Len())

	recv.Flush()

	require.Equal(t, 0, recv.packetList.Len())
}

func TestRecvPeriodicACKLite(t *testing.T) {
	liteACK := false
	recv := mockLiveRecv(
		func(seq circular.Number, light bool) {
			liteACK = light
		},
		nil,
		nil,
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 100 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(10 + i + 1)

		recv.Push(p)
	}

	require.Equal(t, false, liteACK)

	recv.Tick(1)

	require.Equal(t, true, liteACK)
}

func TestSkipTooLate(t *testing.T) {
	seqACK := uint32(0)
	numbers := []uint32{}
	recv := mockLiveRecv(
		func(seq circular.Number, light bool) {
			seqACK = seq.Val()
		},
		nil,
		func(p packet.Packet) {
			numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		},
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 5 {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)

		recv.Push(p)
	}

	recv.Tick(10)

	require.Equal(t, uint32(5), seqACK)
	require.Equal(t, []uint32{0, 1, 2, 3, 4}, numbers)

	for i := 5; i < 10; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(3+i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(13 + i + 1)

		recv.Push(p)
	}

	recv.Tick(20)

	require.Equal(t, uint32(13), seqACK)
	require.Equal(t, []uint32{0, 1, 2, 3, 4, 8, 9}, numbers)
}

func TestIssue67(t *testing.T) {
	ackNumbers := []uint32{}
	nakNumbers := [][2]uint32{}
	numbers := []uint32{}
	recv := mockLiveRecv(
		func(seq circular.Number, light bool) {
			ackNumbers = append(ackNumbers, seq.Val())
		},
		func(list []circular.Number) {
			nakNumbers = append(nakNumbers, [2]uint32{list[0].Val(), list[1].Val()})
		},
		func(p packet.Packet) {
			numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		},
	)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	p := packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(0, packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = 1

	recv.Push(p)

	recv.Tick(10)
	recv.Tick(20)
	recv.Tick(30)
	recv.Tick(40)
	recv.Tick(50)
	recv.Tick(60)
	recv.Tick(70)
	recv.Tick(80)
	recv.Tick(90)

	require.Equal(t, []uint32{1, 1, 1, 1, 1, 1, 1, 1, 1}, ackNumbers)

	p = packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(12, packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = 121

	recv.Push(p)

	require.Equal(t, [][2]uint32{
		{1, 11},
	}, nakNumbers)

	p = packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(1, packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = 11

	recv.Push(p)

	p = packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(11, packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = 111

	recv.Push(p)

	recv.Tick(100)

	require.Equal(t, []uint32{1, 1, 1, 1, 1, 1, 1, 1, 1, 2}, ackNumbers)

	recv.Tick(110)

	require.Equal(t, []uint32{1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2}, ackNumbers)

	recv.Tick(120)

	require.Equal(t, []uint32{1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 13}, ackNumbers)

	recv.Tick(130)

	require.Equal(t, []uint32{1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 13, 13}, ackNumbers)
}

func TestRecvLossMaxTTL(t *testing.T) {
	nNAK := 0
	seqNAKFrom := uint32(0)
	seqNAKTo := uint32(0)

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnSendACK:             nil,
		OnSendNAK: func(list []circular.Number) {
			nNAK++
			if len(list) >= 2 {
				seqNAKFrom = list[0].Val()
				seqNAKTo = list[1].Val()
			}
		},
		OnDeliver:             nil,
		LossMaxTTL:            30, // tolerar desorden hasta 30 paquetes
	}).(*receiver)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	// 1. Envía paquetes en orden de 0 a 4
	for i := 0; i < 5; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)
		recv.Push(p)
	}

	require.Equal(t, 0, nNAK)
	require.Equal(t, uint32(4), recv.maxSeenSequenceNumber.Val())

	// 2. Envía un paquete con un salto pequeño (de 4 a 15, brecha = 10 paquetes)
	// Como la brecha (10) es menor o igual a LossMaxTTL (30), NO debe enviar NAK inmediato
	p := packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(15, packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = 16
	recv.Push(p)

	require.Equal(t, 0, nNAK) // No NAK inmediato!
	require.Equal(t, uint32(15), recv.maxSeenSequenceNumber.Val())

	// 3. Envía un paquete con un salto grande (de 15 a 50, brecha = 34 paquetes)
	// Como la brecha (34) es mayor que LossMaxTTL (30), SÍ debe enviar NAK inmediato
	p2 := packet.NewPacket(addr)
	p2.Header().PacketSequenceNumber = circular.New(50, packet.MAX_SEQUENCENUMBER)
	p2.Header().PktTsbpdTime = 51
	recv.Push(p2)

	require.Equal(t, 1, nNAK) // Se envió un NAK inmediato!
	require.Equal(t, uint32(16), seqNAKFrom)
	require.Equal(t, uint32(49), seqNAKTo)
	require.Equal(t, uint32(50), recv.maxSeenSequenceNumber.Val())
}

func TestRecvPeriodicNAKLossMaxTTL(t *testing.T) {
	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnSendACK:             nil,
		OnSendNAK:             nil,
		OnDeliver:             nil,
		LossMaxTTL:            30, // tolerar desorden hasta 30 paquetes
	}).(*receiver)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	// 1. Envía paquetes en orden 0 a 4
	for i := 0; i < 5; i++ {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)
		recv.Push(p)
	}

	// 2. Envía un paquete con un salto pequeño (de 4 a 15, brecha = 10 paquetes)
	// Como la brecha (10) es <= LossMaxTTL (30), no hay NAK inmediato.
	p := packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(15, packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = 16
	recv.Push(p)

	// 3. Forzar tick de periodicNAK
	// El gap es [5, 14]. Todos los paquetes en el gap están a una distancia de 15 (maxSeen) <= 30.
	// Por tanto, periodicNAK no debería reportar nada.
	list := recv.periodicNAK(30000)
	require.Empty(t, list)

	// 4. Envía un paquete con un salto más grande que LossMaxTTL (ej. de 15 a 50, brecha total de 34)
	// Como la brecha (34) > 30, se enviará NAK inmediato para [16, 49].
	// Pero el gap anterior [5, 14] ahora está a distancia > 30 de 50 (maxSeen).
	// El gap [5, 14] ya no está en la ventana de tolerancia.
	// Hacemos el tick de periodic NAK ahora. Debería reportar el gap [5, 14] porque nakLimit es 50 - 30 = 20.
	// Así, el gap [5, 14] está por debajo de 20 y es reportado, mientras que el nuevo gap [16, 49] tiene elementos
	// dentro de la ventana de tolerancia y solo se reportará la parte que exceda la ventana.
	p2 := packet.NewPacket(addr)
	p2.Header().PacketSequenceNumber = circular.New(50, packet.MAX_SEQUENCENUMBER)
	p2.Header().PktTsbpdTime = 51
	recv.Push(p2)

	list2 := recv.periodicNAK(60000)
	require.NotEmpty(t, list2)

	// El primer gap [5, 14] debe ser completamente reportado.
	// El segundo gap [16, 49] tiene su límite en 50 - 30 = 20.
	// Por lo tanto, el segundo gap reportado debe estar limitado de 16 a 19.
	// Verifiquemos si la lista contiene [5, 14] y [16, 19].
	var vals []uint32
	for _, n := range list2 {
		vals = append(vals, n.Val())
	}
	require.Equal(t, []uint32{5, 14, 16, 19}, vals)
}
````

## File: congestion/live/receive.go
````go
package live

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/congestion"
	"github.com/datarhei/gosrt/packet"
)

// ReceiveConfig is the configuration for the liveRecv congestion control
type ReceiveConfig struct {
	InitialSequenceNumber circular.Number
	PeriodicACKInterval   uint64 // microseconds
	PeriodicNAKInterval   uint64 // microseconds
	OnSendACK             func(seq circular.Number, light bool)
	OnSendNAK             func(list []circular.Number)
	OnDeliver             func(p packet.Packet)
	LossMaxTTL            uint32
}

// receiver implements the Receiver interface
type receiver struct {
	maxSeenSequenceNumber       circular.Number
	lastACKSequenceNumber       circular.Number
	lastDeliveredSequenceNumber circular.Number
	packetList                  *list.List
	lock                        sync.RWMutex

	nPackets uint

	periodicACKInterval uint64 // config
	periodicNAKInterval uint64 // config
	lossMaxTTL            uint32 // config: reorder tolerance

	lastPeriodicACK uint64
	lastPeriodicNAK uint64

	avgPayloadSize  float64 // bytes
	avgLinkCapacity float64 // packets per second

	probeTime    time.Time
	probeNextSeq circular.Number

	statistics congestion.ReceiveStats

	rate struct {
		last   uint64 // microseconds
		period uint64

		packets      uint64
		bytes        uint64
		bytesRetrans uint64

		packetsPerSecond float64
		bytesPerSecond   float64

		pktLossRate float64
	}

	sendACK func(seq circular.Number, light bool)
	sendNAK func(list []circular.Number)
	deliver func(p packet.Packet)
}

// NewReceiver takes a ReceiveConfig and returns a new Receiver
func NewReceiver(config ReceiveConfig) congestion.Receiver {
	r := &receiver{
		maxSeenSequenceNumber:       config.InitialSequenceNumber.Dec(),
		lastACKSequenceNumber:       config.InitialSequenceNumber.Dec(),
		lastDeliveredSequenceNumber: config.InitialSequenceNumber.Dec(),
		packetList:                  list.New(),

		periodicACKInterval: config.PeriodicACKInterval,
		periodicNAKInterval: config.PeriodicNAKInterval,
		lossMaxTTL:            config.LossMaxTTL,

		avgPayloadSize: 1456, //  5.1.2. SRT's Default LiveCC Algorithm

		sendACK: config.OnSendACK,
		sendNAK: config.OnSendNAK,
		deliver: config.OnDeliver,
	}

	if r.sendACK == nil {
		r.sendACK = func(seq circular.Number, light bool) {}
	}

	if r.sendNAK == nil {
		r.sendNAK = func(list []circular.Number) {}
	}

	if r.deliver == nil {
		r.deliver = func(p packet.Packet) {}
	}

	r.rate.last = 0
	r.rate.period = uint64(time.Second.Microseconds())

	return r
}

func (r *receiver) Stats() congestion.ReceiveStats {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.statistics.BytePayload = uint64(r.avgPayloadSize)
	r.statistics.MbpsEstimatedRecvBandwidth = r.rate.bytesPerSecond * 8 / 1024 / 1024
	r.statistics.MbpsEstimatedLinkCapacity = r.avgLinkCapacity * packet.MAX_PAYLOAD_SIZE * 8 / 1024 / 1024
	r.statistics.PktLossRate = r.rate.pktLossRate

	return r.statistics
}

func (r *receiver) PacketRate() (pps, bps, capacity float64) {
	r.lock.Lock()
	defer r.lock.Unlock()

	pps = r.rate.packetsPerSecond
	bps = r.rate.bytesPerSecond
	capacity = r.avgLinkCapacity

	return
}

func (r *receiver) Flush() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.packetList = r.packetList.Init()
}

func (r *receiver) Push(pkt packet.Packet) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if pkt == nil {
		return
	}

	// This is not really well (not at all) described in the specs. See core.cpp and window.h
	// and search for PUMASK_SEQNO_PROBE (0xF). Every 16th and 17th packet are
	// sent in pairs. This is used as a probe for the theoretical capacity of the link.
	if !pkt.Header().RetransmittedPacketFlag {
		probe := pkt.Header().PacketSequenceNumber.Val() & 0xF
		if probe == 0 {
			r.probeTime = time.Now()
			r.probeNextSeq = pkt.Header().PacketSequenceNumber.Inc()
		} else if probe == 1 && pkt.Header().PacketSequenceNumber.Equals(r.probeNextSeq) && !r.probeTime.IsZero() && pkt.Len() != 0 {
			// The time between packets scaled to a fully loaded packet
			diff := float64(time.Since(r.probeTime).Microseconds()) * (packet.MAX_PAYLOAD_SIZE / float64(pkt.Len()))
			if diff != 0 {
				// Here we're doing an average of the measurements.
				r.avgLinkCapacity = 0.875*r.avgLinkCapacity + 0.125*1_000_000/diff
			}
		} else {
			r.probeTime = time.Time{}
		}
	} else {
		r.probeTime = time.Time{}
	}

	r.nPackets++

	pktLen := pkt.Len()

	r.rate.packets++
	r.rate.bytes += pktLen

	r.statistics.Pkt++
	r.statistics.Byte += pktLen

	//pkt.PktTsbpdTime = pkt.Timestamp + r.delay
	if pkt.Header().RetransmittedPacketFlag {
		r.statistics.PktRetrans++
		r.statistics.ByteRetrans += pktLen

		r.rate.bytesRetrans += pktLen
	}

	//  5.1.2. SRT's Default LiveCC Algorithm
	r.avgPayloadSize = 0.875*r.avgPayloadSize + 0.125*float64(pktLen)

	if pkt.Header().PacketSequenceNumber.Lte(r.lastDeliveredSequenceNumber) {
		// Too old, because up until r.lastDeliveredSequenceNumber, we already delivered
		r.statistics.PktBelated++
		r.statistics.ByteBelated += pktLen

		r.statistics.PktDrop++
		r.statistics.ByteDrop += pktLen

		return
	}

	if pkt.Header().PacketSequenceNumber.Lt(r.lastACKSequenceNumber) {
		// Already acknowledged, ignoring
		r.statistics.PktDrop++
		r.statistics.ByteDrop += pktLen

		return
	}

	if pkt.Header().PacketSequenceNumber.Equals(r.maxSeenSequenceNumber.Inc()) {
		// In order, the packet we expected
		r.maxSeenSequenceNumber = pkt.Header().PacketSequenceNumber
	} else if pkt.Header().PacketSequenceNumber.Lte(r.maxSeenSequenceNumber) {
		// Out of order, is it a missing piece? put it in the correct position
		inserted := false
		for e := r.packetList.Back(); e != nil; e = e.Prev() {
			p := e.Value.(packet.Packet)

			if p.Header().PacketSequenceNumber.Equals(pkt.Header().PacketSequenceNumber) {
				// Already received (has been sent more than once), ignoring
				r.statistics.PktDrop++
				r.statistics.ByteDrop += pktLen
				inserted = true
				break
			} else if p.Header().PacketSequenceNumber.Lt(pkt.Header().PacketSequenceNumber) {
				// Late arrival, this fills a gap. Insert after the smaller element
				r.statistics.PktBuf++
				r.statistics.PktUnique++

				r.statistics.ByteBuf += pktLen
				r.statistics.ByteUnique += pktLen

				r.packetList.InsertAfter(pkt, e)
				inserted = true
				break
			}
		}

		if !inserted {
			// If not inserted, it means this packet is smaller than all packets in the list.
			// Insert it at the front of the list.
			r.statistics.PktBuf++
			r.statistics.PktUnique++

			r.statistics.ByteBuf += pktLen
			r.statistics.ByteUnique += pktLen

			r.packetList.PushFront(pkt)
		}

		return
	} else {
		// Too far ahead, there are some missing sequence numbers, immediate NAK report
		// here we can prevent a possibly unnecessary NAK with SRTO_LOXXMAXTTL
		if r.lossMaxTTL == 0 || uint64(pkt.Header().PacketSequenceNumber.Distance(r.maxSeenSequenceNumber)) > uint64(r.lossMaxTTL) {
			r.sendNAK([]circular.Number{
				r.maxSeenSequenceNumber.Inc(),
				pkt.Header().PacketSequenceNumber.Dec(),
			})
		}

		len := uint64(pkt.Header().PacketSequenceNumber.Distance(r.maxSeenSequenceNumber))
		r.statistics.PktLoss += len
		r.statistics.ByteLoss += len * uint64(r.avgPayloadSize)

		r.maxSeenSequenceNumber = pkt.Header().PacketSequenceNumber
	}

	r.statistics.PktBuf++
	r.statistics.PktUnique++

	r.statistics.ByteBuf += pktLen
	r.statistics.ByteUnique += pktLen

	r.packetList.PushBack(pkt)
}

func (r *receiver) periodicACK(now uint64) (ok bool, sequenceNumber circular.Number, lite bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 4.8.1. Packet Acknowledgement (ACKs, ACKACKs)
	if now-r.lastPeriodicACK < r.periodicACKInterval {
		if r.nPackets >= 64 {
			lite = true // Send light ACK
		} else {
			return
		}
	}

	minPktTsbpdTime, maxPktTsbpdTime := uint64(0), uint64(0)
	ackSequenceNumber := r.lastACKSequenceNumber

	e := r.packetList.Front()
	if e != nil {
		p := e.Value.(packet.Packet)

		minPktTsbpdTime = p.Header().PktTsbpdTime
		maxPktTsbpdTime = p.Header().PktTsbpdTime
	}

	// Find the sequence number up until we have all in a row.
	// Where the first gap is (or at the end of the list) is where we can ACK to.

	for e := r.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		// Skip packets that we already ACK'd.
		if p.Header().PacketSequenceNumber.Lte(ackSequenceNumber) {
			continue
		}

		// If there are packets that should have been delivered by now, move forward.
		if p.Header().PktTsbpdTime <= now {
			ackSequenceNumber = p.Header().PacketSequenceNumber
			continue
		}

		// Check if the packet is the next in the row.
		if p.Header().PacketSequenceNumber.Equals(ackSequenceNumber.Inc()) {
			ackSequenceNumber = p.Header().PacketSequenceNumber
			maxPktTsbpdTime = p.Header().PktTsbpdTime
			r.statistics.MsBuf = (maxPktTsbpdTime - minPktTsbpdTime) / 1_000
			continue
		}

		break
	}

	ok = true
	sequenceNumber = ackSequenceNumber.Inc()

	// Keep track of the last ACK's sequence number. With this we can faster ignore
	// packets that come in late that have a lower sequence number.
	r.lastACKSequenceNumber = ackSequenceNumber

	r.lastPeriodicACK = now
	r.nPackets = 0

	return
}

func (r *receiver) periodicNAK(now uint64) []circular.Number {
	r.lock.RLock()
	defer r.lock.RUnlock()

	if now-r.lastPeriodicNAK < r.periodicNAKInterval {
		return nil
	}

	list := []circular.Number{}

	// Send a periodic NAK

	ackSequenceNumber := r.lastACKSequenceNumber

	var nakLimit circular.Number
	if r.lossMaxTTL > 0 {
		nakLimit = r.maxSeenSequenceNumber.Sub(r.lossMaxTTL)
	}

	// Send a NAK for all gaps.
	// Not all gaps might get announced because the size of the NAK packet is limited.
	for e := r.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		// Skip packets that we already ACK'd.
		if p.Header().PacketSequenceNumber.Lte(ackSequenceNumber) {
			continue
		}

		// If this packet is not in sequence, we stop here and report that gap.
		if !p.Header().PacketSequenceNumber.Equals(ackSequenceNumber.Inc()) {
			nackStart := ackSequenceNumber.Inc()
			nackEnd := p.Header().PacketSequenceNumber.Dec()

			if r.lossMaxTTL > 0 {
				if nackEnd.Gte(nakLimit) {
					nackEnd = nakLimit.Dec()
				}
			}

			if nackStart.Lte(nackEnd) {
				list = append(list, nackStart)
				list = append(list, nackEnd)
			}
		}

		ackSequenceNumber = p.Header().PacketSequenceNumber
	}

	r.lastPeriodicNAK = now

	return list
}

func (r *receiver) Tick(now uint64) {
	if ok, sequenceNumber, lite := r.periodicACK(now); ok {
		r.sendACK(sequenceNumber, lite)
	}

	if list := r.periodicNAK(now); len(list) != 0 {
		r.sendNAK(list)
	}

	// Deliver packets whose PktTsbpdTime is ripe
	r.lock.Lock()
	removeList := make([]*list.Element, 0, r.packetList.Len())
	for e := r.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		if p.Header().PacketSequenceNumber.Lte(r.lastACKSequenceNumber) && p.Header().PktTsbpdTime <= now {
			r.statistics.PktBuf--
			r.statistics.ByteBuf -= p.Len()

			r.lastDeliveredSequenceNumber = p.Header().PacketSequenceNumber

			r.deliver(p)
			removeList = append(removeList, e)
		} else {
			break
		}
	}

	for _, e := range removeList {
		r.packetList.Remove(e)
	}
	r.lock.Unlock()

	r.lock.Lock()
	tdiff := now - r.rate.last // microseconds

	if tdiff > r.rate.period {
		r.rate.packetsPerSecond = float64(r.rate.packets) / (float64(tdiff) / 1000 / 1000)
		r.rate.bytesPerSecond = float64(r.rate.bytes) / (float64(tdiff) / 1000 / 1000)
		if r.rate.bytes != 0 {
			r.rate.pktLossRate = float64(r.rate.bytesRetrans) / float64(r.rate.bytes) * 100
		} else {
			r.rate.bytes = 0
		}

		r.rate.packets = 0
		r.rate.bytes = 0
		r.rate.bytesRetrans = 0

		r.rate.last = now
	}
	r.lock.Unlock()
}

func (r *receiver) SetNAKInterval(nakInterval uint64) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.periodicNAKInterval = nakInterval
}

func (r *receiver) String(t uint64) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("maxSeen=%d lastACK=%d lastDelivered=%d\n", r.maxSeenSequenceNumber.Val(), r.lastACKSequenceNumber.Val(), r.lastDeliveredSequenceNumber.Val()))

	r.lock.RLock()
	for e := r.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		b.WriteString(fmt.Sprintf("   %d @ %d (in %d)\n", p.Header().PacketSequenceNumber.Val(), p.Header().PktTsbpdTime, int64(p.Header().PktTsbpdTime)-int64(t)))
	}
	r.lock.RUnlock()

	return b.String()
}
````

## File: congestion/live/send_test.go
````go
package live

import (
	"net"
	"testing"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/packet"
	"github.com/stretchr/testify/require"
)

func mockLiveSend(onDeliver func(p packet.Packet)) *sender {
	send := NewSender(SendConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		DropThreshold:         10,
		OnDeliver:             onDeliver,
	})

	return send.(*sender)
}

func TestSendSequence(t *testing.T) {
	numbers := []uint32{}
	send := mockLiveSend(func(p packet.Packet) {
		numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
	})

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PktTsbpdTime = uint64(i + 1)

		send.Push(p)
	}

	send.Tick(5)

	require.Exactly(t, []uint32{0, 1, 2, 3, 4}, numbers)

	send.Tick(10)

	require.Exactly(t, []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, numbers)
}

func TestSendLossListACK(t *testing.T) {
	send := mockLiveSend(nil)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PktTsbpdTime = uint64(i + 1)

		send.Push(p)
	}

	send.Tick(10)

	require.Equal(t, 10, send.lossList.Len())

	for i := range 10 {
		send.ACK(circular.New(uint32(i+1), packet.MAX_SEQUENCENUMBER))
		require.Equal(t, 10-(i+1), send.lossList.Len())
	}
}

func TestSendRetransmit(t *testing.T) {
	numbers := []uint32{}
	nRetransmit := 0
	send := mockLiveSend(func(p packet.Packet) {
		numbers = append(numbers, p.Header().PacketSequenceNumber.Val())
		if p.Header().RetransmittedPacketFlag {
			nRetransmit++
		}
	})

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PktTsbpdTime = uint64(i + 1)

		send.Push(p)
	}

	send.Tick(10)

	require.Equal(t, 0, nRetransmit)

	send.NAK([]circular.Number{
		circular.New(2, packet.MAX_SEQUENCENUMBER),
		circular.New(2, packet.MAX_SEQUENCENUMBER),
	})

	require.Equal(t, 1, nRetransmit)

	send.NAK([]circular.Number{
		circular.New(5, packet.MAX_SEQUENCENUMBER),
		circular.New(7, packet.MAX_SEQUENCENUMBER),
	})

	require.Equal(t, 4, nRetransmit)
}

func TestSendDrop(t *testing.T) {
	send := mockLiveSend(nil)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PktTsbpdTime = uint64(i + 1)

		send.Push(p)
	}

	send.Tick(10)

	require.Equal(t, 10, send.lossList.Len())

	send.Tick(20)

	require.Equal(t, 0, send.lossList.Len())
}

func TestSendFlush(t *testing.T) {
	send := mockLiveSend(nil)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")

	for i := range 10 {
		p := packet.NewPacket(addr)
		p.Header().PktTsbpdTime = uint64(i + 1)

		send.Push(p)
	}

	require.Exactly(t, 10, send.packetList.Len())
	require.Exactly(t, 0, send.lossList.Len())

	send.Tick(5)

	require.Exactly(t, 5, send.packetList.Len())
	require.Exactly(t, 5, send.lossList.Len())

	send.Flush()

	require.Exactly(t, 0, send.packetList.Len())
	require.Exactly(t, 0, send.lossList.Len())
}
````

## File: congestion/live/send.go
````go
package live

import (
	"container/list"
	"sync"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/congestion"
	"github.com/datarhei/gosrt/packet"
)

// SendConfig is the configuration for the liveSend congestion control
type SendConfig struct {
	InitialSequenceNumber circular.Number
	DropThreshold         uint64
	MaxBW                 int64
	InputBW               int64
	MinInputBW            int64
	OverheadBW            int64
	OnDeliver             func(p packet.Packet)
}

// sender implements the Sender interface
type sender struct {
	nextSequenceNumber circular.Number
	dropThreshold      uint64

	packetList *list.List
	lossList   *list.List
	lock       sync.RWMutex

	avgPayloadSize float64 // bytes
	pktSndPeriod   float64 // microseconds
	maxBW          float64 // bytes/s
	inputBW        float64 // bytes/s
	overheadBW     float64 // percent

	statistics congestion.SendStats

	probeTime uint64

	rate struct {
		period uint64 // microseconds
		last   uint64

		bytes        uint64
		bytesSent    uint64
		bytesRetrans uint64

		estimatedInputBW float64 // bytes/s
		estimatedSentBW  float64 // bytes/s

		pktLossRate float64
	}

	deliver func(p packet.Packet)
}

// NewSender takes a SendConfig and returns a new Sender
func NewSender(config SendConfig) congestion.Sender {
	s := &sender{
		nextSequenceNumber: config.InitialSequenceNumber,
		dropThreshold:      config.DropThreshold,
		packetList:         list.New(),
		lossList:           list.New(),

		avgPayloadSize: packet.MAX_PAYLOAD_SIZE, //  5.1.2. SRT's Default LiveCC Algorithm
		maxBW:          float64(config.MaxBW),
		inputBW:        float64(config.InputBW),
		overheadBW:     float64(config.OverheadBW),

		deliver: config.OnDeliver,
	}

	if s.deliver == nil {
		s.deliver = func(p packet.Packet) {}
	}

	s.maxBW = 128 * 1024 * 1024 // 1 Gbit/s
	s.pktSndPeriod = (s.avgPayloadSize + 16) * 1_000_000 / s.maxBW

	s.rate.period = uint64(time.Second.Microseconds())
	s.rate.last = 0

	return s
}

func (s *sender) Stats() congestion.SendStats {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.statistics.UsPktSndPeriod = s.pktSndPeriod
	s.statistics.BytePayload = uint64(s.avgPayloadSize)
	s.statistics.MsBuf = 0

	max := s.lossList.Back()
	min := s.lossList.Front()

	if max != nil && min != nil {
		s.statistics.MsBuf = (max.Value.(packet.Packet).Header().PktTsbpdTime - min.Value.(packet.Packet).Header().PktTsbpdTime) / 1_000
	}

	s.statistics.MbpsEstimatedInputBandwidth = s.rate.estimatedInputBW * 8 / 1024 / 1024
	s.statistics.MbpsEstimatedSentBandwidth = s.rate.estimatedSentBW * 8 / 1024 / 1024

	s.statistics.PktLossRate = s.rate.pktLossRate

	return s.statistics
}

func (s *sender) Flush() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.packetList = s.packetList.Init()
	s.lossList = s.lossList.Init()
}

func (s *sender) Push(p packet.Packet) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if p == nil {
		return
	}

	// Give to the packet a sequence number
	p.Header().PacketSequenceNumber = s.nextSequenceNumber
	p.Header().PacketPositionFlag = packet.SinglePacket
	p.Header().OrderFlag = false
	p.Header().MessageNumber = 1

	s.nextSequenceNumber = s.nextSequenceNumber.Inc()

	pktLen := p.Len()

	s.statistics.PktBuf++
	s.statistics.ByteBuf += pktLen

	// Input bandwidth calculation
	s.rate.bytes += pktLen

	p.Header().Timestamp = uint32(p.Header().PktTsbpdTime & uint64(packet.MAX_TIMESTAMP))

	// Every 16th and 17th packet should be sent at the same time in order
	// for the receiver to determine the link capacity. Not really well
	// documented in the specs.
	// PktTsbpdTime is used for the timing of sending the packets. Here we
	// can modify it because it has already been used to set the packet's
	// timestamp.
	probe := p.Header().PacketSequenceNumber.Val() & 0xF
	if probe == 0 {
		s.probeTime = p.Header().PktTsbpdTime
	} else if probe == 1 {
		p.Header().PktTsbpdTime = s.probeTime
	}

	s.packetList.PushBack(p)

	s.statistics.PktFlightSize = uint64(s.packetList.Len())
}

func (s *sender) Tick(now uint64) {
	// Deliver packets whose PktTsbpdTime is ripe
	s.lock.Lock()
	removeList := make([]*list.Element, 0, s.packetList.Len())
	for e := s.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)
		if p.Header().PktTsbpdTime <= now {
			s.statistics.Pkt++
			s.statistics.PktUnique++

			pktLen := p.Len()

			s.statistics.Byte += pktLen
			s.statistics.ByteUnique += pktLen

			s.statistics.UsSndDuration += uint64(s.pktSndPeriod)

			//  5.1.2. SRT's Default LiveCC Algorithm
			s.avgPayloadSize = 0.875*s.avgPayloadSize + 0.125*float64(pktLen)

			s.rate.bytesSent += pktLen

			s.deliver(p)
			removeList = append(removeList, e)
		} else {
			break
		}
	}

	for _, e := range removeList {
		s.lossList.PushBack(e.Value)
		s.packetList.Remove(e)
	}
	s.lock.Unlock()

	s.lock.Lock()
	removeList = make([]*list.Element, 0, s.lossList.Len())
	for e := s.lossList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		if p.Header().PktTsbpdTime+s.dropThreshold <= now {
			// Dropped packet because too old
			s.statistics.PktDrop++
			s.statistics.PktLoss++
			s.statistics.ByteDrop += p.Len()
			s.statistics.ByteLoss += p.Len()

			removeList = append(removeList, e)
		}
	}

	// These packets are not needed anymore (too late)
	for _, e := range removeList {
		p := e.Value.(packet.Packet)

		s.statistics.PktBuf--
		s.statistics.ByteBuf -= p.Len()

		s.lossList.Remove(e)

		// This packet has been ACK'd and we don't need it anymore
		p.Decommission()
	}
	s.lock.Unlock()

	s.lock.Lock()
	tdiff := now - s.rate.last

	if tdiff > s.rate.period {
		s.rate.estimatedInputBW = float64(s.rate.bytes) / (float64(tdiff) / 1000 / 1000)
		s.rate.estimatedSentBW = float64(s.rate.bytesSent) / (float64(tdiff) / 1000 / 1000)
		if s.rate.bytesSent != 0 {
			s.rate.pktLossRate = float64(s.rate.bytesRetrans) / float64(s.rate.bytesSent) * 100
		} else {
			s.rate.pktLossRate = 0
		}

		s.rate.bytes = 0
		s.rate.bytesSent = 0
		s.rate.bytesRetrans = 0

		s.rate.last = now
	}
	s.lock.Unlock()
}

func (s *sender) ACK(sequenceNumber circular.Number) {
	s.lock.Lock()
	defer s.lock.Unlock()

	removeList := make([]*list.Element, 0, s.lossList.Len())
	for e := s.lossList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)
		if p.Header().PacketSequenceNumber.Lt(sequenceNumber) {
			// Remove packet from buffer because it has been successfully transmitted
			removeList = append(removeList, e)
		} else {
			break
		}
	}

	// These packets are not needed anymore (ACK'd)
	for _, e := range removeList {
		p := e.Value.(packet.Packet)

		s.statistics.PktBuf--
		s.statistics.ByteBuf -= p.Len()

		s.lossList.Remove(e)

		// This packet has been ACK'd and we don't need it anymore
		p.Decommission()
	}

	s.pktSndPeriod = (s.avgPayloadSize + 16) * 1000000 / s.maxBW
}

func (s *sender) NAK(sequenceNumbers []circular.Number) {
	if len(sequenceNumbers) == 0 {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	for e := s.lossList.Back(); e != nil; e = e.Prev() {
		p := e.Value.(packet.Packet)

		for i := 0; i < len(sequenceNumbers); i += 2 {
			if p.Header().PacketSequenceNumber.Gte(sequenceNumbers[i]) && p.Header().PacketSequenceNumber.Lte(sequenceNumbers[i+1]) {
				s.statistics.PktRetrans++
				s.statistics.Pkt++
				s.statistics.PktLoss++

				s.statistics.ByteRetrans += p.Len()
				s.statistics.Byte += p.Len()
				s.statistics.ByteLoss += p.Len()

				//  5.1.2. SRT's Default LiveCC Algorithm
				s.avgPayloadSize = 0.875*s.avgPayloadSize + 0.125*float64(p.Len())

				s.rate.bytesSent += p.Len()
				s.rate.bytesRetrans += p.Len()

				p.Header().RetransmittedPacketFlag = true
				s.deliver(p)
			}
		}
	}
}

func (s *sender) SetDropThreshold(threshold uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.dropThreshold = threshold
}
````

## File: conn_request.go
````go
package srt

import (
	"fmt"
	"net"
	"time"

	"github.com/datarhei/gosrt/crypto"
	"github.com/datarhei/gosrt/packet"
	"github.com/datarhei/gosrt/rand"
)

// ConnRequest is an incoming connection request
type ConnRequest interface {
	// RemoteAddr returns the address of the peer. The returned net.Addr
	// is a copy and can be used at will.
	RemoteAddr() net.Addr

	// Version returns the handshake version of the incoming request. Currently
	// known versions are 4 and 5. With version 4 the StreamId will always be
	// empty and IsEncrypted will always return false. An incoming version 4
	// connection will always be publishing.
	Version() uint32

	// StreamId returns the streamid of the requesting connection. Use this
	// to decide what to do with the connection.
	StreamId() string

	// SocketId return the socketid of the connection.
	SocketId() uint32

	// PeerSocketId returns the socketid of the peer of the connection.
	PeerSocketId() uint32

	// IsEncrypted returns whether the connection is encrypted. If it is
	// encrypted, use SetPassphrase to set the passphrase for decrypting.
	IsEncrypted() bool

	// SetPassphrase sets the passphrase in order to decrypt the incoming
	// data. Returns an error if the passphrase did not work or the connection
	// is not encrypted.
	SetPassphrase(p string) error

	// SetRejectionReason sets the rejection reason for the connection. If
	// no set, REJ_PEER will be used.
	//
	// Deprecated: replaced by Reject().
	SetRejectionReason(r RejectionReason)

	// Accept accepts the request and returns a connection.
	Accept() (Conn, error)

	// Reject rejects the request.
	Reject(r RejectionReason)
}

// connRequest implements the ConnRequest interface
type connRequest struct {
	ln              *listener
	addr            net.Addr
	localAddr       net.Addr
	start           time.Time
	socketId        uint32
	peerSocketId    uint32
	timestamp       uint32
	config          Config
	handshake       *packet.CIFHandshake
	crypto          crypto.Crypto
	passphrase      string
	rejectionReason RejectionReason
}

func newConnRequest(ln *listener, p packet.Packet) *connRequest {
	cif := &packet.CIFHandshake{}

	err := p.UnmarshalCIF(cif)

	ln.log("handshake:recv:dump", func() string { return p.Dump() })
	ln.log("handshake:recv:cif", func() string { return cif.String() })

	if err != nil {
		ln.log("handshake:recv:error", func() string { return err.Error() })
		return nil
	}

	// Assemble the response (4.3.1.  Caller-Listener Handshake)

	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(ln.start).Microseconds())
	p.Header().DestinationSocketId = cif.SRTSocketId

	cif.PeerIP.FromNetAddr(ln.addr)

	// Create a copy of the configuration for the connection
	config := ln.config

	if cif.HandshakeType == packet.HSTYPE_INDUCTION {
		// cif
		cif.Version = 5
		cif.EncryptionField = 0 // Don't advertise any specific encryption method
		cif.ExtensionField = 0x4A17
		//cif.initialPacketSequenceNumber = newCircular(0, MAX_SEQUENCENUMBER)
		//cif.maxTransmissionUnitSize = 0
		//cif.maxFlowWindowSize = 0
		//cif.SRTSocketId = 0
		cif.SynCookie = ln.syncookie.Get(p.Header().Addr.String())

		p.MarshalCIF(cif)

		ln.log("handshake:send:dump", func() string { return p.Dump() })
		ln.log("handshake:send:cif", func() string { return cif.String() })

		ln.send(p)
	} else if cif.HandshakeType == packet.HSTYPE_CONCLUSION {
		// Verify the SYN cookie
		if !ln.syncookie.Verify(cif.SynCookie, p.Header().Addr.String()) {
			cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
			ln.log("handshake:recv:error", func() string { return "invalid SYN cookie" })
			p.MarshalCIF(cif)
			ln.log("handshake:send:dump", func() string { return p.Dump() })
			ln.log("handshake:send:cif", func() string { return cif.String() })
			ln.send(p)

			return nil
		}

		// Peer is advertising a too big MSS
		if cif.MaxTransmissionUnitSize > MAX_MSS_SIZE {
			cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
			ln.log("handshake:recv:error", func() string { return fmt.Sprintf("MTU is too big (%d bytes)", cif.MaxTransmissionUnitSize) })
			p.MarshalCIF(cif)
			ln.log("handshake:send:dump", func() string { return p.Dump() })
			ln.log("handshake:send:cif", func() string { return cif.String() })
			ln.send(p)

			return nil
		}

		// If the peer has a smaller MTU size, adjust to it
		if cif.MaxTransmissionUnitSize < config.MSS {
			config.MSS = cif.MaxTransmissionUnitSize
			config.PayloadSize = config.MSS - SRT_HEADER_SIZE - UDP_HEADER_SIZE

			if config.PayloadSize < MIN_PAYLOAD_SIZE {
				cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
				ln.log("handshake:recv:error", func() string { return fmt.Sprintf("payload size is too small (%d bytes)", config.PayloadSize) })
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)
			}
		}

		// We only support HSv4 and HSv5
		if cif.Version == 4 {
			// Check if the type (encryption field + extension field) has the value 2
			if cif.EncryptionField != 0 || cif.ExtensionField != 2 {
				cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
				ln.log("handshake:recv:error", func() string { return "invalid type, expecting a value of 2 (UDT_DGRAM)" })
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)

				return nil
			}
		} else if cif.Version == 5 {
			if cif.SRTHS == nil {
				cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
				ln.log("handshake:recv:error", func() string { return "missing handshake extension" })
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)

				return nil
			}

			// Check if the peer version is sufficient
			if cif.SRTHS.SRTVersion < config.MinVersion {
				cif.HandshakeType = packet.HandshakeType(REJ_VERSION)
				ln.log("handshake:recv:error", func() string {
					return fmt.Sprintf("peer version insufficient (%#06x), expecting at least %#06x", cif.SRTHS.SRTVersion, config.MinVersion)
				})
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)

				return nil
			}

			// Check the required SRT flags
			if !cif.SRTHS.SRTFlags.TSBPDSND || !cif.SRTHS.SRTFlags.TSBPDRCV || !cif.SRTHS.SRTFlags.TLPKTDROP || !cif.SRTHS.SRTFlags.PERIODICNAK || !cif.SRTHS.SRTFlags.REXMITFLG {
				cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
				ln.log("handshake:recv:error", func() string { return "not all required flags are set" })
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)

				return nil
			}

			// We only support live streaming
			if cif.SRTHS.SRTFlags.STREAM {
				cif.HandshakeType = packet.HandshakeType(REJ_MESSAGEAPI)
				ln.log("handshake:recv:error", func() string { return "only live streaming is supported" })
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)

				return nil
			}

			// We only support live congestion control
			if cif.HasCongestionCtl && cif.CongestionCtl != "live" {
				cif.HandshakeType = packet.HandshakeType(REJ_CONGESTION)
				ln.log("handshake:recv:error", func() string { return "only live congestion control is supported" })
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)

				return nil
			}
		} else {
			cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
			ln.log("handshake:recv:error", func() string { return fmt.Sprintf("only HSv4 and HSv5 are supported (got HSv%d)", cif.Version) })
			p.MarshalCIF(cif)
			ln.log("handshake:send:dump", func() string { return p.Dump() })
			ln.log("handshake:send:cif", func() string { return cif.String() })
			ln.send(p)

			return nil
		}

		req := &connRequest{
			ln:           ln,
			addr:         p.Header().Addr,
			localAddr:    p.Header().LocalAddr,
			start:        time.Now(),
			socketId:     cif.SRTSocketId,
			peerSocketId: cif.SRTSocketId,
			timestamp:    p.Header().Timestamp,
			config:       config,
			handshake:    cif,
		}

		if cif.SRTKM != nil {
			cr, err := crypto.New(int(cif.SRTKM.KLen))
			if err != nil {
				cif.HandshakeType = packet.HandshakeType(REJ_ROGUE)
				ln.log("handshake:recv:error", func() string { return fmt.Sprintf("crypto: %s", err) })
				p.MarshalCIF(cif)
				ln.log("handshake:send:dump", func() string { return p.Dump() })
				ln.log("handshake:send:cif", func() string { return cif.String() })
				ln.send(p)

				return nil
			}

			req.crypto = cr
		}

		ln.lock.Lock()

		// We received a duplicate request: re-send conclusion response if the connection is already active
		conn, exists := ln.connsByPeer[cif.SRTSocketId]
		if exists {
			if conn != nil {
				ln.log("handshake:recv:duplicate", func() string { return fmt.Sprintf("re-sending conclusion response for peer socket %#08x", cif.SRTSocketId) })
				
				respCIF := &packet.CIFHandshake{
					IsRequest:                   false,
					Version:                     cif.Version,
					EncryptionField:             cif.EncryptionField,
					ExtensionField:              cif.ExtensionField,
					InitialPacketSequenceNumber: conn.initialPacketSequenceNumber,
					MaxTransmissionUnitSize:     conn.config.MSS,
					MaxFlowWindowSize:           cif.MaxFlowWindowSize,
					HandshakeType:               packet.HSTYPE_CONCLUSION,
					SRTSocketId:                 conn.socketId,
					SynCookie:                   0,
				}
				respCIF.PeerIP.FromNetAddr(ln.addr)

				if cif.Version == 5 {
					respCIF.HasHS = true
					respCIF.SRTHS = &packet.CIFHandshakeExtension{
						SRTVersion:     SRT_VERSION,
						RecvTSBPDDelay: uint16(conn.tsbpdDelay / 1000),
						SendTSBPDDelay: uint16(conn.peerTsbpdDelay / 1000),
					}
					respCIF.SRTHS.SRTFlags.TSBPDSND = true
					respCIF.SRTHS.SRTFlags.TSBPDRCV = true
					respCIF.SRTHS.SRTFlags.CRYPT = true
					respCIF.SRTHS.SRTFlags.TLPKTDROP = true
					respCIF.SRTHS.SRTFlags.PERIODICNAK = true
					respCIF.SRTHS.SRTFlags.REXMITFLG = true
					respCIF.SRTHS.SRTFlags.STREAM = false
					respCIF.SRTHS.SRTFlags.PACKET_FILTER = false
				}

				respPkt := packet.NewPacket(p.Header().Addr)
				respPkt.Header().IsControlPacket = true
				respPkt.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
				respPkt.Header().SubType = 0
				respPkt.Header().TypeSpecific = 0
				respPkt.Header().Timestamp = uint32(time.Since(ln.start).Microseconds())
				respPkt.Header().DestinationSocketId = conn.socketId
				respPkt.Header().LocalAddr = p.Header().LocalAddr
				respPkt.MarshalCIF(respCIF)

				ln.log("handshake:send:duplicate:dump", func() string { return respPkt.Dump() })
				ln.send(respPkt)
			}
			ln.lock.Unlock()
			return nil
		}

		// Already fill connsByPeer for this connection
		ln.connsByPeer[cif.SRTSocketId] = nil

		// Already reserve a socketId for this connection
		socketId, err := req.generateSocketId()
		if err == nil {
			ln.conns[socketId] = nil
			req.socketId = socketId
		}

		ln.lock.Unlock()

		// We couldn't create a socketId: reject silently
		if err != nil {
			return nil
		}

		return req
	} else {
		if cif.HandshakeType.IsRejection() {
			ln.log("handshake:recv:error", func() string { return fmt.Sprintf("connection rejected: %s", cif.HandshakeType.String()) })
		} else {
			ln.log("handshake:recv:error", func() string { return fmt.Sprintf("unsupported handshake: %s", cif.HandshakeType.String()) })
		}
	}

	return nil
}

func (req *connRequest) RemoteAddr() net.Addr {
	addr, _ := net.ResolveUDPAddr("udp", req.addr.String())
	return addr
}

func (req *connRequest) Version() uint32 {
	return req.handshake.Version
}

func (req *connRequest) StreamId() string {
	return req.handshake.StreamId
}

func (req *connRequest) SocketId() uint32 {
	return req.socketId
}

func (req *connRequest) PeerSocketId() uint32 {
	return req.peerSocketId
}

func (req *connRequest) IsEncrypted() bool {
	return req.crypto != nil
}

func (req *connRequest) SetPassphrase(passphrase string) error {
	if req.handshake.Version == 5 {
		if req.crypto == nil {
			return fmt.Errorf("listen: request without encryption")
		}

		if err := req.crypto.UnmarshalKM(req.handshake.SRTKM, passphrase); err != nil {
			return err
		}
	}

	req.passphrase = passphrase

	return nil
}

func (req *connRequest) SetRejectionReason(reason RejectionReason) {
	req.rejectionReason = reason
}

func (req *connRequest) Reject(reason RejectionReason) {
	req.ln.lock.Lock()
	defer req.ln.lock.Unlock()

	if cr, hasReq := req.ln.connsByPeer[req.peerSocketId]; !hasReq || cr != nil {
		return
	}

	p := packet.NewPacket(req.addr)
	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(req.ln.start).Microseconds())
	p.Header().DestinationSocketId = req.peerSocketId
	p.Header().LocalAddr = req.localAddr
	req.handshake.HandshakeType = packet.HandshakeType(reason)
	p.MarshalCIF(req.handshake)
	req.ln.log("handshake:send:dump", func() string { return p.Dump() })
	req.ln.log("handshake:send:cif", func() string { return req.handshake.String() })
	req.ln.send(p)

	delete(req.ln.conns, req.socketId)
	delete(req.ln.connsByPeer, req.peerSocketId)
}

// generateSocketId generates an SRT SocketID that can be used for this connection
func (req *connRequest) generateSocketId() (uint32, error) {
	for range 10 {
		socketId, err := rand.Uint32()
		if err != nil {
			return 0, fmt.Errorf("could not generate random socket id")
		}

		// check that the socket id is not already in use
		if _, found := req.ln.conns[socketId]; !found {
			return socketId, nil
		}
	}

	return 0, fmt.Errorf("could not generate unused socketid")
}

func (req *connRequest) Accept() (Conn, error) {
	if req.crypto != nil && len(req.passphrase) == 0 {
		req.Reject(REJ_BADSECRET)
		return nil, fmt.Errorf("passphrase is missing")
	}

	req.ln.lock.Lock()
	defer req.ln.lock.Unlock()

	if cr, hasReq := req.ln.connsByPeer[req.peerSocketId]; !hasReq || cr != nil {
		return nil, fmt.Errorf("connection already accepted")
	}

	// Select the largest TSBPD delay advertised by the caller, but at least 120ms
	recvTsbpdDelay := uint16(req.config.ReceiverLatency.Milliseconds())
	sendTsbpdDelay := uint16(req.config.PeerLatency.Milliseconds())

	if req.handshake.Version == 5 {
		if req.handshake.SRTHS.SendTSBPDDelay > recvTsbpdDelay {
			recvTsbpdDelay = req.handshake.SRTHS.SendTSBPDDelay
		}

		if req.handshake.SRTHS.RecvTSBPDDelay > sendTsbpdDelay {
			sendTsbpdDelay = req.handshake.SRTHS.RecvTSBPDDelay
		}

		req.config.StreamId = req.handshake.StreamId
	}

	req.config.Passphrase = req.passphrase

	localAddr := req.localAddr
	if localAddr == nil {
		localAddr = req.ln.addr
	}

	// Create a new connection
	conn := newSRTConn(srtConnConfig{
		version:                     req.handshake.Version,
		localAddr:                   localAddr,
		remoteAddr:                  req.addr,
		config:                      req.config,
		start:                       req.start,
		socketId:                    req.socketId,
		peerSocketId:                req.peerSocketId,
		tsbpdTimeBase:               uint64(req.timestamp),
		tsbpdDelay:                  uint64(recvTsbpdDelay) * 1000,
		peerTsbpdDelay:              uint64(sendTsbpdDelay) * 1000,
		initialPacketSequenceNumber: req.handshake.InitialPacketSequenceNumber,
		crypto:                      req.crypto,
		keyBaseEncryption:           packet.EvenKeyEncrypted,
		onSend:                      req.ln.send,
		onShutdown:                  req.ln.handleShutdown,
		logger:                      req.config.Logger,
	})

	req.ln.log("connection:new", func() string { return fmt.Sprintf("%#08x (%s)", conn.SocketId(), conn.StreamId()) })

	req.handshake.SRTSocketId = req.socketId
	req.handshake.SynCookie = 0

	if req.handshake.Version == 5 {
		//  3.2.1.1.1.  Handshake Extension Message Flags
		req.handshake.SRTHS.SRTVersion = SRT_VERSION
		req.handshake.SRTHS.SRTFlags.TSBPDSND = true
		req.handshake.SRTHS.SRTFlags.TSBPDRCV = true
		req.handshake.SRTHS.SRTFlags.CRYPT = true
		req.handshake.SRTHS.SRTFlags.TLPKTDROP = true
		req.handshake.SRTHS.SRTFlags.PERIODICNAK = true
		req.handshake.SRTHS.SRTFlags.REXMITFLG = true
		req.handshake.SRTHS.SRTFlags.STREAM = false
		req.handshake.SRTHS.SRTFlags.PACKET_FILTER = false
		req.handshake.SRTHS.RecvTSBPDDelay = recvTsbpdDelay
		req.handshake.SRTHS.SendTSBPDDelay = sendTsbpdDelay
	}

	p := packet.NewPacket(req.addr)
	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(req.start).Microseconds())
	p.Header().DestinationSocketId = req.peerSocketId
	p.Header().LocalAddr = req.localAddr
	p.MarshalCIF(req.handshake)
	req.ln.log("handshake:send:dump", func() string { return p.Dump() })
	req.ln.log("handshake:send:cif", func() string { return req.handshake.String() })
	req.ln.send(p)

	req.ln.conns[req.socketId] = conn
	req.ln.connsByPeer[req.peerSocketId] = conn

	return conn, nil
}
````

## File: connection_test.go
````go
package srt

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/datarhei/gosrt/packet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryption(t *testing.T) {
	message := "Hello World!"
	passphrase := "foobarfoobar"
	channel := NewPubSub(PubSubConfig{})

	config := DefaultConfig()
	config.EnforcedEncryption = true

	server := Server{
		Addr:   "127.0.0.1:6003",
		Config: &config,
		HandleConnect: func(req ConnRequest) ConnType {
			if req.IsEncrypted() {
				if err := req.SetPassphrase(passphrase); err != nil {
					return REJECT
				}
			}

			streamid := req.StreamId()

			if streamid == "publish" {
				return PUBLISH
			} else if streamid == "subscribe" {
				return SUBSCRIBE
			}

			return REJECT
		},
		HandlePublish: func(conn Conn) {
			channel.Publish(conn)

			conn.Close()
		},
		HandleSubscribe: func(conn Conn) {
			channel.Subscribe(conn)

			conn.Close()
		},
	}

	err := server.Listen()
	require.NoError(t, err)

	defer server.Shutdown()

	go func() {
		err := server.Serve()
		if err == ErrServerClosed {
			return
		}
		require.NoError(t, err)
	}()

	{
		// Reject connection if wrong password is set
		config := DefaultConfig()
		config.StreamId = "subscribe"
		config.Passphrase = "barfoobarfoo"

		_, err := Dial("srt", "127.0.0.1:6003", config)
		require.Error(t, err)
	}
	// Test transmitting an encrypted message

	readerConnected := make(chan struct{})
	readerDone := make(chan struct{})

	dataReader1 := bytes.Buffer{}

	go func() {
		defer close(readerDone)

		config := DefaultConfig()
		config.StreamId = "subscribe"
		config.Passphrase = "foobarfoobar"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		close(readerConnected)

		buffer := make([]byte, 2048)

		for {
			n, err := conn.Read(buffer)
			if n != 0 {
				dataReader1.Write(buffer[:n])
			}

			if err != nil {
				break
			}
		}

		err = conn.Close()
		require.NoError(t, err)
	}()

	<-readerConnected

	writerDone := make(chan struct{})

	go func() {
		defer close(writerDone)

		config := DefaultConfig()
		config.StreamId = "publish"
		config.Passphrase = "foobarfoobar"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		n, err := conn.Write([]byte(message))
		if !assert.NoError(t, err) {
			panic(err.Error())
		}
		assert.Equal(t, 12, n)

		time.Sleep(3 * time.Second)

		err = conn.Close()
		assert.NoError(t, err)
	}()

	<-writerDone
	<-readerDone

	reader1 := dataReader1.String()

	require.Equal(t, message, reader1)
}

// Test for https://github.com/datarhei/gosrt/pull/94
func TestEncryptionRetransmit(t *testing.T) {
	message := "Hello World!"
	passphrase := "foobarfoobar"
	channel := NewPubSub(PubSubConfig{})

	config := DefaultConfig()
	config.EnforcedEncryption = true

	server := Server{
		Addr:   "127.0.0.1:6003",
		Config: &config,
		HandleConnect: func(req ConnRequest) ConnType {
			if req.IsEncrypted() {
				if err := req.SetPassphrase(passphrase); err != nil {
					return REJECT
				}
			}

			streamid := req.StreamId()

			if streamid == "publish" {
				return PUBLISH
			} else if streamid == "subscribe" {
				return SUBSCRIBE
			}

			return REJECT
		},
		HandlePublish: func(conn Conn) {
			channel.Publish(conn)

			conn.Close()
		},
		HandleSubscribe: func(conn Conn) {
			channel.Subscribe(conn)

			conn.Close()
		},
	}

	err := server.Listen()
	require.NoError(t, err)

	defer server.Shutdown()

	go func() {
		err := server.Serve()
		if err == ErrServerClosed {
			return
		}
		require.NoError(t, err)
	}()

	{
		// Reject connection if wrong password is set
		config := DefaultConfig()
		config.StreamId = "subscribe"
		config.Passphrase = "barfoobarfoo"

		_, err := Dial("srt", "127.0.0.1:6003", config)
		require.Error(t, err)
	}

	// Test transmitting an encrypted message

	readerConnected := make(chan struct{})
	readerDone := make(chan struct{})

	dataReader1 := bytes.Buffer{}

	go func() {
		defer close(readerDone)

		config := DefaultConfig()
		config.StreamId = "subscribe"
		config.Passphrase = "foobarfoobar"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		close(readerConnected)

		buffer := make([]byte, 2048)

		for {
			n, err := conn.Read(buffer)
			if n != 0 {
				dataReader1.Write(buffer[:n])
			}

			if err != nil {
				break
			}
		}

		err = conn.Close()
		require.NoError(t, err)
	}()

	<-readerConnected

	writerDone := make(chan struct{})

	go func() {
		defer close(writerDone)

		config := DefaultConfig()
		config.StreamId = "publish"
		config.Passphrase = "foobarfoobar"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		counter := 0

		dialer, _ := conn.(*dialer)
		originalOnSend := dialer.conn.onSend
		dialer.conn.onSend = func(p packet.Packet) {
			if !p.Header().IsControlPacket {
				// Drop every 2nd original packet
				if !p.Header().RetransmittedPacketFlag {
					counter++
					if counter%2 == 0 {
						return
					}
				}
			}

			originalOnSend(p)
		}

		for range 5 {
			n, err := conn.Write([]byte(message))
			if !assert.NoError(t, err) {
				panic(err.Error())
			}
			assert.Equal(t, 12, n)
		}

		time.Sleep(3 * time.Second)

		err = conn.Close()
		assert.NoError(t, err)
	}()

	<-writerDone
	<-readerDone

	reader1 := dataReader1.String()

	require.Equal(t, message+message+message+message+message, reader1)
}

func TestEncryptionKeySwap(t *testing.T) {
	message := "Hello World!"
	passphrase := "foobarfoobar"
	channel := NewPubSub(PubSubConfig{})

	config := DefaultConfig()
	config.EnforcedEncryption = true

	server := Server{
		Addr:   "127.0.0.1:6003",
		Config: &config,
		HandleConnect: func(req ConnRequest) ConnType {
			if req.IsEncrypted() {
				if err := req.SetPassphrase(passphrase); err != nil {
					return REJECT
				}
			}

			streamid := req.StreamId()

			if streamid == "publish" {
				return PUBLISH
			} else if streamid == "subscribe" {
				return SUBSCRIBE
			}

			return REJECT
		},
		HandlePublish: func(conn Conn) {
			channel.Publish(conn)

			conn.Close()
		},
		HandleSubscribe: func(conn Conn) {
			channel.Subscribe(conn)

			conn.Close()
		},
	}

	err := server.Listen()
	require.NoError(t, err)

	defer server.Shutdown()

	go func() {
		err := server.Serve()
		if err == ErrServerClosed {
			return
		}
		require.NoError(t, err)
	}()

	// Test transmitting encrypted messages with key swap in between

	dataReader1 := bytes.Buffer{}

	readerConnected := make(chan struct{})
	readerDone := make(chan struct{})

	go func() {
		defer close(readerDone)

		config := DefaultConfig()
		config.StreamId = "subscribe"
		config.Passphrase = "foobarfoobar"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		buffer := make([]byte, 2048)

		close(readerConnected)

		for {
			n, err := conn.Read(buffer)
			if n != 0 {
				dataReader1.Write(buffer[:n])
			}

			if err != nil {
				break
			}
		}

		err = conn.Close()
		assert.NoError(t, err)
	}()

	<-readerConnected

	writerDone := make(chan struct{})

	go func() {
		defer close(writerDone)

		config := DefaultConfig()
		config.StreamId = "publish"
		config.Passphrase = "foobarfoobar"
		// Swap encryption key after 50 sent messages
		config.KMPreAnnounce = 10
		config.KMRefreshRate = 30

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		// Send 150 messages
		for range 150 {
			n, err := conn.Write([]byte(message))
			if !assert.NoError(t, err) {
				panic(err.Error())
			}
			assert.Equal(t, 12, n)
		}

		time.Sleep(3 * time.Second)

		err = conn.Close()
		assert.NoError(t, err)
	}()

	<-writerDone
	<-readerDone

	reader1 := dataReader1.String()

	require.Equal(t, strings.Repeat(message, 150), reader1)
}

func TestStats(t *testing.T) {
	message := "Hello World!"
	channel := NewPubSub(PubSubConfig{})

	config := DefaultConfig()

	server := Server{
		Addr:   "127.0.0.1:6003",
		Config: &config,
		HandleConnect: func(req ConnRequest) ConnType {
			streamid := req.StreamId()

			if streamid == "publish" {
				return PUBLISH
			} else if streamid == "subscribe" {
				return SUBSCRIBE
			}

			return REJECT
		},
		HandlePublish: func(conn Conn) {
			channel.Publish(conn)

			conn.Close()
		},
		HandleSubscribe: func(conn Conn) {
			channel.Subscribe(conn)

			conn.Close()
		},
	}

	err := server.Listen()
	require.NoError(t, err)

	defer server.Shutdown()

	go func() {
		err := server.Serve()
		if err == ErrServerClosed {
			return
		}
		require.NoError(t, err)
	}()

	statsReader := Statistics{}
	statsWriter := Statistics{}

	readerConnected := make(chan struct{})
	readerDone := make(chan struct{})

	dataReader1 := bytes.Buffer{}

	go func() {
		defer close(readerDone)

		config := DefaultConfig()
		config.StreamId = "subscribe"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		close(readerConnected)

		buffer := make([]byte, 2048)

		for {
			n, err := conn.Read(buffer)
			if n != 0 {
				dataReader1.Write(buffer[:n])
			}

			if err != nil {
				break
			}
		}

		conn.Stats(&statsReader)

		err = conn.Close()
		require.NoError(t, err)
	}()

	<-readerConnected

	writerDone := make(chan struct{})

	go func() {
		defer close(writerDone)

		config := DefaultConfig()
		config.StreamId = "publish"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		n, err := conn.Write([]byte(message))
		if !assert.NoError(t, err) {
			panic(err.Error())
		}
		assert.Equal(t, 12, n)

		time.Sleep(3 * time.Second)

		conn.Stats(&statsWriter)

		err = conn.Close()
		assert.NoError(t, err)
	}()

	<-writerDone
	<-readerDone

	reader1 := dataReader1.String()

	require.Equal(t, message, reader1)

	require.Equal(t, uint64(len(message)+44), statsReader.Accumulated.ByteRecv)
	require.Equal(t, uint64(1), statsReader.Accumulated.PktRecv)

	require.Equal(t, uint64(len(message)+44), statsWriter.Accumulated.ByteSent)
	require.Equal(t, uint64(1), statsWriter.Accumulated.PktSent)
}
````

## File: connection.go
````go
package srt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/congestion"
	"github.com/datarhei/gosrt/congestion/live"
	"github.com/datarhei/gosrt/crypto"
	"github.com/datarhei/gosrt/packet"
)

// Conn is a SRT network connection.
type Conn interface {
	// Read reads data from the connection.
	// Read can be made to time out and return an error after a fixed
	// time limit; see SetDeadline and SetReadDeadline.
	Read(p []byte) (int, error)

	// ReadPacket reads a packet from the queue of received packets. It blocks
	// if the queue is empty. Only data packets are returned. Using ReadPacket
	// and Read at the same time may lead to data loss.
	ReadPacket() (packet.Packet, error)

	// Write writes data to the connection.
	// Write can be made to time out and return an error after a fixed
	// time limit; see SetDeadline and SetWriteDeadline.
	Write(p []byte) (int, error)

	// WritePacket writes a packet to the write queue. Packets on the write queue
	// will be sent to the peer of the connection. Only data packets will be sent.
	WritePacket(p packet.Packet) error

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// LocalAddr returns the local network address. The returned net.Addr is not shared by other invocations of LocalAddr.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address. The returned net.Addr is not shared by other invocations of RemoteAddr.
	RemoteAddr() net.Addr

	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error

	// SocketId return the socketid of the connection.
	SocketId() uint32

	// PeerSocketId returns the socketid of the peer of the connection.
	PeerSocketId() uint32

	// StreamId returns the streamid use for the connection.
	StreamId() string

	// Stats returns accumulated and instantaneous statistics of the connection.
	Stats(s *Statistics)

	// Version returns the connection version, either 4 or 5. With version 4, the streamid is not available
	Version() uint32
}

type rtt struct {
	rtt    float64 // microseconds
	rttVar float64 // microseconds

	lock sync.RWMutex
}

func (r *rtt) Recalculate(rtt time.Duration) {
	// 4.10.  Round-Trip Time Estimation
	lastRTT := float64(rtt.Microseconds())

	r.lock.Lock()
	defer r.lock.Unlock()

	r.rtt = r.rtt*0.875 + lastRTT*0.125
	r.rttVar = r.rttVar*0.75 + math.Abs(r.rtt-lastRTT)*0.25
}

func (r *rtt) RTT() float64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.rtt
}

func (r *rtt) RTTVar() float64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.rttVar
}

func (r *rtt) NAKInterval() float64 {
	r.lock.RLock()
	defer r.lock.RUnlock()

	// 4.8.2.  Packet Retransmission (NAKs)
	nakInterval := (r.rtt + 4*r.rttVar) / 2
	if nakInterval < 40000 {
		nakInterval = 40000 // 40ms
	}

	return nakInterval
}

type connStats struct {
	headerSize        uint64
	pktSentACK        uint64
	pktRecvACK        uint64
	pktSentACKACK     uint64
	pktRecvACKACK     uint64
	pktSentNAK        uint64
	pktRecvNAK        uint64
	pktSentKM         uint64
	pktRecvKM         uint64
	pktRecvUndecrypt  uint64
	byteRecvUndecrypt uint64
	pktRecvInvalid    uint64
	pktSentKeepalive  uint64
	pktRecvKeepalive  uint64
	pktSentShutdown   uint64
	pktRecvShutdown   uint64
	mbpsLinkCapacity  float64
}

// Check if we implement the net.Conn interface
var _ net.Conn = &srtConn{}

type srtConn struct {
	version  uint32
	isCaller bool // Only relevant if version == 4

	localAddr  net.Addr
	remoteAddr net.Addr

	start time.Time

	shutdownOnce sync.Once

	socketId     uint32
	peerSocketId uint32

	config Config

	crypto                 crypto.Crypto
	keyBaseEncryption      packet.PacketEncryption
	kmPreAnnounceCountdown uint64
	kmRefreshCountdown     uint64
	kmConfirmed            bool
	cryptoLock             sync.Mutex

	peerIdleTimeout *time.Timer

	rtt rtt // microseconds

	ackLock       sync.RWMutex
	ackNumbers    map[uint32]time.Time
	nextACKNumber circular.Number

	initialPacketSequenceNumber circular.Number

	tsbpdTimeBase       uint64 // microseconds
	tsbpdWrapPeriod     bool
	tsbpdTimeBaseOffset uint64 // microseconds
	tsbpdDelay          uint64 // microseconds
	tsbpdDrift          uint64 // microseconds
	peerTsbpdDelay      uint64 // microseconds
	dropThreshold       uint64 // microseconds

	// Queue for packets that are coming from the network
	networkQueue chan packet.Packet

	// Queue for packets that are written with writePacket() and will be send to the network
	writeQueue  chan packet.Packet
	writeBuffer bytes.Buffer
	writeData   []byte

	// Queue for packets that will be read locally with ReadPacket()
	readQueue  chan packet.Packet
	readBuffer bytes.Buffer

	onSend     func(p packet.Packet)
	onShutdown func(*srtConn)

	tick time.Duration

	// Congestion control
	recv congestion.Receiver
	snd  congestion.Sender

	// context of all channels and routines
	ctx       context.Context
	cancelCtx context.CancelFunc

	statistics     connStats
	statisticsLock sync.RWMutex

	logger Logger

	debug struct {
		expectedRcvPacketSequenceNumber  circular.Number
		expectedReadPacketSequenceNumber circular.Number
	}

	// HSv4
	stopHSRequests context.CancelFunc
	stopKMRequests context.CancelFunc
}

type srtConnConfig struct {
	version                     uint32
	isCaller                    bool
	localAddr                   net.Addr
	remoteAddr                  net.Addr
	config                      Config
	start                       time.Time
	socketId                    uint32
	peerSocketId                uint32
	tsbpdTimeBase               uint64 // microseconds
	tsbpdDelay                  uint64 // microseconds
	peerTsbpdDelay              uint64 // microseconds
	initialPacketSequenceNumber circular.Number
	crypto                      crypto.Crypto
	keyBaseEncryption           packet.PacketEncryption
	onSend                      func(p packet.Packet)
	onShutdown                  func(*srtConn)
	logger                      Logger
}

func newSRTConn(config srtConnConfig) *srtConn {
	c := &srtConn{
		version:                     config.version,
		isCaller:                    config.isCaller,
		localAddr:                   config.localAddr,
		remoteAddr:                  config.remoteAddr,
		config:                      config.config,
		start:                       config.start,
		socketId:                    config.socketId,
		peerSocketId:                config.peerSocketId,
		tsbpdTimeBase:               config.tsbpdTimeBase,
		tsbpdDelay:                  config.tsbpdDelay,
		peerTsbpdDelay:              config.peerTsbpdDelay,
		initialPacketSequenceNumber: config.initialPacketSequenceNumber,
		crypto:                      config.crypto,
		keyBaseEncryption:           config.keyBaseEncryption,
		onSend:                      config.onSend,
		onShutdown:                  config.onShutdown,
		logger:                      config.logger,
	}

	if c.onSend == nil {
		c.onSend = func(p packet.Packet) {}
	}

	if c.onShutdown == nil {
		c.onShutdown = func(*srtConn) {}
	}

	c.nextACKNumber = circular.New(1, packet.MAX_TIMESTAMP)
	c.ackNumbers = make(map[uint32]time.Time)

	c.kmPreAnnounceCountdown = c.config.KMRefreshRate - c.config.KMPreAnnounce
	c.kmRefreshCountdown = c.config.KMRefreshRate

	// 4.10.  Round-Trip Time Estimation
	c.rtt = rtt{
		rtt:    float64((100 * time.Millisecond).Microseconds()),
		rttVar: float64((50 * time.Millisecond).Microseconds()),
	}

	c.networkQueue = make(chan packet.Packet, 8192)

	c.writeQueue = make(chan packet.Packet, 8192)
	if c.version == 4 {
		// libsrt-1.2.3 receiver doesn't like it when the payload is larger than 7*188 bytes.
		// Here we just take a multiple of a mpegts chunk size.
		c.writeData = make([]byte, int(c.config.PayloadSize/188*188))
	} else {
		// For v5 we use the max. payload size: https://github.com/Haivision/srt/issues/876
		c.writeData = make([]byte, int(c.config.PayloadSize))
	}

	c.readQueue = make(chan packet.Packet, 8192)

	c.peerIdleTimeout = time.AfterFunc(c.config.PeerIdleTimeout, func() {
		c.log("connection:close", func() string {
			return fmt.Sprintf("no more data received from peer for %s. shutting down", c.config.PeerIdleTimeout)
		})
		go c.close()
	})

	c.tick = 20 * time.Millisecond

	// 4.8.1.  Packet Acknowledgement (ACKs, ACKACKs) -> periodicACK = 20 milliseconds
	// 4.8.2.  Packet Retransmission (NAKs) -> periodicNAK at least 40 milliseconds
	c.recv = live.NewReceiver(live.ReceiveConfig{
		InitialSequenceNumber: c.initialPacketSequenceNumber,
		PeriodicACKInterval:   20_000,
		PeriodicNAKInterval:   40_000,
		OnSendACK:             c.sendACK,
		OnSendNAK:             c.sendNAK,
		OnDeliver:             c.deliver,
		LossMaxTTL:            c.config.LossMaxTTL,
	})

	// 4.6.  Too-Late Packet Drop -> 125% of SRT latency, at least 1 second
	// https://github.com/Haivision/srt/blob/master/docs/API/API-socket-options.md#SRTO_SNDDROPDELAY
	c.dropThreshold = max(uint64(float64(c.peerTsbpdDelay)*1.25)+uint64(c.config.SendDropDelay.Microseconds()), uint64(time.Second.Microseconds()))
	c.dropThreshold += 20_000

	c.snd = live.NewSender(live.SendConfig{
		InitialSequenceNumber: c.initialPacketSequenceNumber,
		DropThreshold:         c.dropThreshold,
		MaxBW:                 c.config.MaxBW,
		InputBW:               c.config.InputBW,
		MinInputBW:            c.config.MinInputBW,
		OverheadBW:            c.config.OverheadBW,
		OnDeliver:             c.pop,
	})

	c.ctx, c.cancelCtx = context.WithCancel(context.Background())

	go c.networkQueueReader(c.ctx)
	go c.writeQueueReader(c.ctx)
	go c.ticker(c.ctx)

	c.debug.expectedRcvPacketSequenceNumber = c.initialPacketSequenceNumber
	c.debug.expectedReadPacketSequenceNumber = c.initialPacketSequenceNumber

	c.statistics.headerSize = 8 + 16 // 8 bytes UDP + 16 bytes SRT
	if strings.Count(c.localAddr.String(), ":") < 2 {
		c.statistics.headerSize += 20 // 20 bytes IPv4 header
	} else {
		c.statistics.headerSize += 40 // 40 bytes IPv6 header
	}

	if c.version == 4 && c.isCaller {
		var hsrequestsCtx context.Context
		hsrequestsCtx, c.stopHSRequests = context.WithCancel(context.Background())
		go c.sendHSRequests(hsrequestsCtx)

		if c.crypto != nil {
			var kmrequestsCtx context.Context
			kmrequestsCtx, c.stopKMRequests = context.WithCancel(context.Background())
			go c.sendKMRequests(kmrequestsCtx)
		}
	}

	return c
}

func (c *srtConn) LocalAddr() net.Addr {
	if c.localAddr == nil {
		return nil
	}

	addr, _ := net.ResolveUDPAddr("udp", c.localAddr.String())
	return addr
}

func (c *srtConn) RemoteAddr() net.Addr {
	if c.remoteAddr == nil {
		return nil
	}

	addr, _ := net.ResolveUDPAddr("udp", c.remoteAddr.String())
	return addr
}

func (c *srtConn) SocketId() uint32 {
	return c.socketId
}

func (c *srtConn) PeerSocketId() uint32 {
	return c.peerSocketId
}

func (c *srtConn) StreamId() string {
	return c.config.StreamId
}

func (c *srtConn) Version() uint32 {
	return c.version
}

// ticker invokes the congestion control in regular intervals with
// the current connection time.
func (c *srtConn) ticker(ctx context.Context) {
	ticker := time.NewTicker(c.tick)
	defer ticker.Stop()
	defer func() {
		c.log("connection:close", func() string { return "left ticker loop" })
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			tickTime := uint64(t.Sub(c.start).Microseconds())

			c.recv.Tick(c.tsbpdTimeBase + tickTime)
			c.snd.Tick(tickTime)
		}
	}
}

func (c *srtConn) ReadPacket() (packet.Packet, error) {
	var p packet.Packet
	select {
	case <-c.ctx.Done():
		return nil, io.EOF
	case p = <-c.readQueue:
	}

	if p.Header().PacketSequenceNumber.Gt(c.debug.expectedReadPacketSequenceNumber) {
		c.log("connection:error", func() string {
			return fmt.Sprintf("lost packets. got: %d, expected: %d (%d)", p.Header().PacketSequenceNumber.Val(), c.debug.expectedReadPacketSequenceNumber.Val(), c.debug.expectedReadPacketSequenceNumber.Distance(p.Header().PacketSequenceNumber))
		})
	} else if p.Header().PacketSequenceNumber.Lt(c.debug.expectedReadPacketSequenceNumber) {
		c.log("connection:error", func() string {
			return fmt.Sprintf("packet out of order. got: %d, expected: %d (%d)", p.Header().PacketSequenceNumber.Val(), c.debug.expectedReadPacketSequenceNumber.Val(), c.debug.expectedReadPacketSequenceNumber.Distance(p.Header().PacketSequenceNumber))
		})
		return nil, io.EOF
	}

	c.debug.expectedReadPacketSequenceNumber = p.Header().PacketSequenceNumber.Inc()

	return p, nil
}

func (c *srtConn) Read(b []byte) (int, error) {
	if c.readBuffer.Len() != 0 {
		return c.readBuffer.Read(b)
	}

	c.readBuffer.Reset()

	p, err := c.ReadPacket()
	if err != nil {
		return 0, err
	}

	c.readBuffer.Write(p.Data())

	// The packet is out of congestion control and written to the read buffer
	p.Decommission()

	return c.readBuffer.Read(b)
}

// WritePacket writes a packet to the write queue. Packets on the write queue
// will be sent to the peer of the connection. Only data packets will be sent.
func (c *srtConn) WritePacket(p packet.Packet) error {
	if p.Header().IsControlPacket {
		// Ignore control packets
		return nil
	}

	_, err := c.Write(p.Data())
	if err != nil {
		return err
	}

	return nil
}

func (c *srtConn) Write(b []byte) (int, error) {
	c.writeBuffer.Write(b)

	for {
		n, err := c.writeBuffer.Read(c.writeData)
		if err != nil {
			return 0, err
		}

		p := packet.NewPacket(nil)

		p.SetData(c.writeData[:n])

		p.Header().IsControlPacket = false
		// Give the packet a deliver timestamp
		p.Header().PktTsbpdTime = c.getTimestamp()

		// Non-blocking write to the write queue
		select {
		case <-c.ctx.Done():
			return 0, io.EOF
		case c.writeQueue <- p:
		default:
			return 0, io.EOF
		}

		if c.writeBuffer.Len() == 0 {
			break
		}
	}

	c.writeBuffer.Reset()

	return len(b), nil
}

// push puts a packet on the network queue. This is where packets go that came in from the network.
func (c *srtConn) push(p packet.Packet) {
	// Non-blocking write to the network queue
	select {
	case <-c.ctx.Done():
	case c.networkQueue <- p:
	default:
		c.log("connection:error", func() string { return "network queue is full" })
	}
}

// getTimestamp returns the elapsed time since the start of the connection in microseconds.
func (c *srtConn) getTimestamp() uint64 {
	return uint64(time.Since(c.start).Microseconds())
}

// getTimestampForPacket returns the elapsed time since the start of the connection in
// microseconds clamped a 32bit value.
func (c *srtConn) getTimestampForPacket() uint32 {
	return uint32(c.getTimestamp() & uint64(packet.MAX_TIMESTAMP))
}

// pop adds the destination address and socketid to the packet and sends it out to the network.
// The packet will be encrypted if required.
func (c *srtConn) pop(p packet.Packet) {
	p.Header().Addr = c.remoteAddr
	p.Header().LocalAddr = c.localAddr
	p.Header().DestinationSocketId = c.peerSocketId

	if !p.Header().IsControlPacket {
		c.cryptoLock.Lock()
		if c.crypto != nil {
			p.Header().KeyBaseEncryptionFlag = c.keyBaseEncryption
			if !p.Header().RetransmittedPacketFlag {
				c.crypto.EncryptOrDecryptPayload(p.Data(), p.Header().KeyBaseEncryptionFlag, p.Header().PacketSequenceNumber.Val())
			}

			c.kmPreAnnounceCountdown--
			c.kmRefreshCountdown--

			if c.kmPreAnnounceCountdown == 0 && !c.kmConfirmed {
				c.sendKMRequest(c.keyBaseEncryption.Opposite())

				// Resend the request until we get a response
				c.kmPreAnnounceCountdown = c.config.KMPreAnnounce/10 + 1
			}

			if c.kmRefreshCountdown == 0 {
				c.kmPreAnnounceCountdown = c.config.KMRefreshRate - c.config.KMPreAnnounce
				c.kmRefreshCountdown = c.config.KMRefreshRate

				// Switch the keys
				c.keyBaseEncryption = c.keyBaseEncryption.Opposite()

				c.kmConfirmed = false
			}

			if c.kmRefreshCountdown == c.config.KMRefreshRate-c.config.KMPreAnnounce {
				// Decommission the previous key, resp. create a new SEK that will
				// be used in the next switch.
				c.crypto.GenerateSEK(c.keyBaseEncryption.Opposite())
			}
		}
		c.cryptoLock.Unlock()

		c.log("data:send:dump", func() string { return p.Dump() })
	}

	// Send the packet on the wire
	c.onSend(p)
}

// networkQueueReader reads the packets from the network queue in order to process them.
func (c *srtConn) networkQueueReader(ctx context.Context) {
	defer func() {
		c.log("connection:close", func() string { return "left network queue reader loop" })
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-c.networkQueue:
			c.handlePacket(p)
		}
	}
}

// writeQueueReader reads the packets from the write queue and puts them into congestion
// control for sending.
func (c *srtConn) writeQueueReader(ctx context.Context) {
	defer func() {
		c.log("connection:close", func() string { return "left write queue reader loop" })
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-c.writeQueue:
			// Put the packet into the send congestion control
			c.snd.Push(p)
		}
	}
}

// deliver writes the packets to the read queue in order to be consumed by the Read function.
func (c *srtConn) deliver(p packet.Packet) {
	// Non-blocking write to the read queue
	select {
	case <-c.ctx.Done():
	case c.readQueue <- p:
	default:
		c.log("connection:error", func() string { return "readQueue was blocking, dropping packet" })
	}
}

// handlePacket checks the packet header. If it is a control packet it will forwarded to the
// respective handler. If it is a data packet it will be put into congestion control for
// receiving. The packet will be decrypted if required.
func (c *srtConn) handlePacket(p packet.Packet) {
	if p == nil {
		return
	}

	c.peerIdleTimeout.Reset(c.config.PeerIdleTimeout)

	header := p.Header()

	if header.IsControlPacket {
		if header.ControlType == packet.CTRLTYPE_KEEPALIVE {
			c.handleKeepAlive(p)
		} else if header.ControlType == packet.CTRLTYPE_SHUTDOWN {
			c.handleShutdown(p)
		} else if header.ControlType == packet.CTRLTYPE_NAK {
			c.handleNAK(p)
		} else if header.ControlType == packet.CTRLTYPE_ACK {
			c.handleACK(p)
		} else if header.ControlType == packet.CTRLTYPE_ACKACK {
			c.handleACKACK(p)
		} else if header.ControlType == packet.CTRLTYPE_USER {
			c.log("connection:recv:ctrl:user", func() string {
				return fmt.Sprintf("got CTRLTYPE_USER packet, subType: %s", header.SubType)
			})

			// HSv4 Extension
			if header.SubType == packet.EXTTYPE_HSREQ {
				c.handleHSRequest(p)
			} else if header.SubType == packet.EXTTYPE_HSRSP {
				c.handleHSResponse(p)
			}

			// 3.2.2.  Key Material
			if header.SubType == packet.EXTTYPE_KMREQ {
				c.handleKMRequest(p)
			} else if header.SubType == packet.EXTTYPE_KMRSP {
				c.handleKMResponse(p)
			}
		}

		return
	}

	if header.PacketSequenceNumber.Gt(c.debug.expectedRcvPacketSequenceNumber) {
		c.log("connection:error", func() string {
			return fmt.Sprintf("recv lost packets. got: %d, expected: %d (%d)\n", header.PacketSequenceNumber.Val(), c.debug.expectedRcvPacketSequenceNumber.Val(), c.debug.expectedRcvPacketSequenceNumber.Distance(header.PacketSequenceNumber))
		})
	}

	c.debug.expectedRcvPacketSequenceNumber = header.PacketSequenceNumber.Inc()

	//fmt.Printf("%s\n", p.String())

	// Ignore FEC filter control packets
	// https://github.com/Haivision/srt/blob/master/docs/features/packet-filtering-and-fec.md
	// "An FEC control packet is distinguished from a regular data packet by having
	// its message number equal to 0. This value isn't normally used in SRT (message
	// numbers start from 1, increment to a maximum, and then roll back to 1)."
	if header.MessageNumber == 0 {
		c.log("connection:filter", func() string { return "dropped FEC filter control packet" })
		return
	}

	// 4.5.1.1.  TSBPD Time Base Calculation
	if !c.tsbpdWrapPeriod {
		if header.Timestamp > packet.MAX_TIMESTAMP-(30*1000000) {
			c.tsbpdWrapPeriod = true
			c.log("connection:tsbpd", func() string { return "TSBPD wrapping period started" })
		}
	} else {
		if header.Timestamp >= (30*1000000) && header.Timestamp <= (60*1000000) {
			c.tsbpdWrapPeriod = false
			c.tsbpdTimeBaseOffset += uint64(packet.MAX_TIMESTAMP) + 1
			c.log("connection:tsbpd", func() string { return "TSBPD wrapping period finished" })
		}
	}

	tsbpdTimeBaseOffset := c.tsbpdTimeBaseOffset
	if c.tsbpdWrapPeriod {
		if header.Timestamp < (30 * 1000000) {
			tsbpdTimeBaseOffset += uint64(packet.MAX_TIMESTAMP) + 1
		}
	}

	header.PktTsbpdTime = c.tsbpdTimeBase + tsbpdTimeBaseOffset + uint64(header.Timestamp) + c.tsbpdDelay + c.tsbpdDrift

	c.log("data:recv:dump", func() string { return p.Dump() })

	c.cryptoLock.Lock()
	if c.crypto != nil {
		if header.KeyBaseEncryptionFlag != 0 {
			if err := c.crypto.EncryptOrDecryptPayload(p.Data(), header.KeyBaseEncryptionFlag, header.PacketSequenceNumber.Val()); err != nil {
				c.statisticsLock.Lock()
				c.statistics.pktRecvUndecrypt++
				c.statistics.byteRecvUndecrypt += p.Len()
				c.statisticsLock.Unlock()
			}
		} else {
			c.statisticsLock.Lock()
			c.statistics.pktRecvUndecrypt++
			c.statistics.byteRecvUndecrypt += p.Len()
			c.statisticsLock.Unlock()
		}
	}
	c.cryptoLock.Unlock()

	// Put the packet into receive congestion control
	c.recv.Push(p)
}

// handleKeepAlive resets the idle timeout and sends a keepalive to the peer.
func (c *srtConn) handleKeepAlive(p packet.Packet) {
	c.log("control:recv:keepalive:dump", func() string { return p.Dump() })

	c.statisticsLock.Lock()
	c.statistics.pktRecvKeepalive++
	c.statistics.pktSentKeepalive++
	c.statisticsLock.Unlock()

	c.peerIdleTimeout.Reset(c.config.PeerIdleTimeout)

	c.log("control:send:keepalive:dump", func() string { return p.Dump() })

	c.pop(p)
}

// handleShutdown closes the connection
func (c *srtConn) handleShutdown(p packet.Packet) {
	c.log("control:recv:shutdown:dump", func() string { return p.Dump() })

	c.statisticsLock.Lock()
	c.statistics.pktRecvShutdown++
	c.statisticsLock.Unlock()

	go c.close()
}

// handleACK forwards the acknowledge sequence number to the congestion control and
// returns a ACKACK (on a full ACK). The RTT is also updated in case of a full ACK.
func (c *srtConn) handleACK(p packet.Packet) {
	c.log("control:recv:ACK:dump", func() string { return p.Dump() })

	c.statisticsLock.Lock()
	c.statistics.pktRecvACK++
	c.statisticsLock.Unlock()

	cif := &packet.CIFACK{}

	if err := p.UnmarshalCIF(cif); err != nil {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:ACK:error", func() string { return fmt.Sprintf("invalid ACK: %s", err) })
		return
	}

	c.log("control:recv:ACK:cif", func() string { return cif.String() })

	c.snd.ACK(cif.LastACKPacketSequenceNumber)

	if !cif.IsLite && !cif.IsSmall {
		// 4.10.  Round-Trip Time Estimation
		c.recalculateRTT(time.Duration(int64(cif.RTT)) * time.Microsecond)

		// Estimated Link Capacity (from packets/s to Mbps)
		c.statisticsLock.Lock()
		c.statistics.mbpsLinkCapacity = float64(cif.EstimatedLinkCapacity) * MAX_PAYLOAD_SIZE * 8 / 1024 / 1024
		c.statisticsLock.Unlock()

		c.sendACKACK(p.Header().TypeSpecific)
	}
}

// handleNAK forwards the lost sequence number to the congestion control.
func (c *srtConn) handleNAK(p packet.Packet) {
	c.log("control:recv:NAK:dump", func() string { return p.Dump() })

	c.statisticsLock.Lock()
	c.statistics.pktRecvNAK++
	c.statisticsLock.Unlock()

	cif := &packet.CIFNAK{}

	if err := p.UnmarshalCIF(cif); err != nil {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:NAK:error", func() string { return fmt.Sprintf("invalid NAK: %s", err) })
		return
	}

	c.log("control:recv:NAK:cif", func() string { return cif.String() })

	// Inform congestion control about lost packets
	c.snd.NAK(cif.LostPacketSequenceNumber)
}

// handleACKACK updates the RTT and NAK interval for the congestion control.
func (c *srtConn) handleACKACK(p packet.Packet) {
	c.ackLock.Lock()

	c.statisticsLock.Lock()
	c.statistics.pktRecvACKACK++
	c.statisticsLock.Unlock()

	c.log("control:recv:ACKACK:dump", func() string { return p.Dump() })

	// p.typeSpecific is the ACKNumber
	if ts, ok := c.ackNumbers[p.Header().TypeSpecific]; ok {
		// 4.10.  Round-Trip Time Estimation
		c.recalculateRTT(time.Since(ts))
		delete(c.ackNumbers, p.Header().TypeSpecific)
	} else {
		c.log("control:recv:ACKACK:error", func() string { return fmt.Sprintf("got unknown ACKACK (%d)", p.Header().TypeSpecific) })
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
	}

	for i := range c.ackNumbers {
		if i < p.Header().TypeSpecific {
			delete(c.ackNumbers, i)
		}
	}

	c.ackLock.Unlock()

	c.recv.SetNAKInterval(uint64(c.rtt.NAKInterval()))
}

// recalculateRTT recalculates the RTT based on a full ACK exchange
func (c *srtConn) recalculateRTT(rtt time.Duration) {
	c.rtt.Recalculate(rtt)

	c.log("connection:rtt", func() string {
		return fmt.Sprintf("RTT=%.0fus RTTVar=%.0fus NAKInterval=%.0fms", c.rtt.RTT(), c.rtt.RTTVar(), c.rtt.NAKInterval()/1000)
	})
}

// handleHSRequest handles the HSv4 handshake extension request and sends the response
func (c *srtConn) handleHSRequest(p packet.Packet) {
	c.log("control:recv:HSReq:dump", func() string { return p.Dump() })

	cif := &packet.CIFHandshakeExtension{}

	if err := p.UnmarshalCIF(cif); err != nil {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:HSReq:error", func() string { return fmt.Sprintf("invalid HSReq: %s", err) })
		return
	}

	c.log("control:recv:HSReq:cif", func() string { return cif.String() })

	// Check for version
	if cif.SRTVersion < 0x010200 || cif.SRTVersion >= 0x010300 {
		c.log("control:recv:HSReq:error", func() string { return fmt.Sprintf("unsupported version: %#08x", cif.SRTVersion) })
		c.close()
		return
	}

	// Check the required SRT flags
	if !cif.SRTFlags.TSBPDSND {
		c.log("control:recv:HSRes:error", func() string { return "TSBPDSND flag must be set" })
		c.close()

		return
	}

	if !cif.SRTFlags.TLPKTDROP {
		c.log("control:recv:HSRes:error", func() string { return "TLPKTDROP flag must be set" })
		c.close()

		return
	}

	if !cif.SRTFlags.CRYPT {
		c.log("control:recv:HSRes:error", func() string { return "CRYPT flag must be set" })
		c.close()

		return
	}

	if !cif.SRTFlags.REXMITFLG {
		c.log("control:recv:HSRes:error", func() string { return "REXMITFLG flag must be set" })
		c.close()

		return
	}

	// we as receiver don't need this
	cif.SRTFlags.TSBPDSND = false

	// we as receiver are supporting these
	cif.SRTFlags.TSBPDRCV = true
	cif.SRTFlags.PERIODICNAK = true

	// These flag was introduced in HSv5 and should not be set in HSv4
	if cif.SRTFlags.STREAM {
		c.log("control:recv:HSReq:error", func() string { return "STREAM flag is set" })
		c.close()
		return
	}

	if cif.SRTFlags.PACKET_FILTER {
		c.log("control:recv:HSReq:error", func() string { return "PACKET_FILTER flag is set" })
		c.close()
		return
	}

	recvTsbpdDelay := max(cif.SendTSBPDDelay, uint16(c.config.ReceiverLatency.Milliseconds()))

	c.tsbpdDelay = uint64(recvTsbpdDelay) * 1000

	cif.RecvTSBPDDelay = 0
	cif.SendTSBPDDelay = recvTsbpdDelay

	p.MarshalCIF(cif)

	// Send HS Response
	p.Header().SubType = packet.EXTTYPE_HSRSP

	c.pop(p)
}

// handleHSResponse handles the HSv4 handshake extension response
func (c *srtConn) handleHSResponse(p packet.Packet) {
	c.log("control:recv:HSRes:dump", func() string { return p.Dump() })

	cif := &packet.CIFHandshakeExtension{}

	if err := p.UnmarshalCIF(cif); err != nil {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:HSRes:error", func() string { return fmt.Sprintf("invalid HSRes: %s", err) })
		return
	}

	c.log("control:recv:HSRes:cif", func() string { return cif.String() })

	if c.version == 4 {
		// Check for version
		if cif.SRTVersion < 0x010200 || cif.SRTVersion >= 0x010300 {
			c.log("control:recv:HSRes:error", func() string { return fmt.Sprintf("unsupported version: %#08x", cif.SRTVersion) })
			c.close()
			return
		}

		// TSBPDSND is not relevant from the receiver
		// PERIODICNAK is the sender's decision, we don't care, but will handle them

		// Check the required SRT flags
		if !cif.SRTFlags.TSBPDRCV {
			c.log("control:recv:HSRes:error", func() string { return "TSBPDRCV flag must be set" })
			c.close()

			return
		}

		if !cif.SRTFlags.TLPKTDROP {
			c.log("control:recv:HSRes:error", func() string { return "TLPKTDROP flag must be set" })
			c.close()

			return
		}

		if !cif.SRTFlags.CRYPT {
			c.log("control:recv:HSRes:error", func() string { return "CRYPT flag must be set" })
			c.close()

			return
		}

		if !cif.SRTFlags.REXMITFLG {
			c.log("control:recv:HSRes:error", func() string { return "REXMITFLG flag must be set" })
			c.close()

			return
		}

		// These flag was introduced in HSv5 and should not be set in HSv4
		if cif.SRTFlags.STREAM {
			c.log("control:recv:HSReq:error", func() string { return "STREAM flag is set" })
			c.close()
			return
		}

		if cif.SRTFlags.PACKET_FILTER {
			c.log("control:recv:HSReq:error", func() string { return "PACKET_FILTER flag is set" })
			c.close()
			return
		}

		sendTsbpdDelay := max(cif.SendTSBPDDelay, uint16(c.config.PeerLatency.Milliseconds()))

		c.dropThreshold = max(uint64(float64(sendTsbpdDelay)*1.25)+uint64(c.config.SendDropDelay.Microseconds()), uint64(time.Second.Microseconds()))
		c.dropThreshold += 20_000

		c.snd.SetDropThreshold(c.dropThreshold)

		c.stopHSRequests()
	}
}

// handleKMRequest checks if the key material is valid and responds with a KM response.
func (c *srtConn) handleKMRequest(p packet.Packet) {
	c.log("control:recv:KMReq:dump", func() string { return p.Dump() })

	c.statisticsLock.Lock()
	c.statistics.pktRecvKM++
	c.statisticsLock.Unlock()

	cif := &packet.CIFKeyMaterialExtension{}

	if err := p.UnmarshalCIF(cif); err != nil {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:KMReq:error", func() string { return fmt.Sprintf("invalid KMReq: %s", err) })
		return
	}

	c.log("control:recv:KMReq:cif", func() string { return cif.String() })

	c.cryptoLock.Lock()

	if c.version == 4 && c.crypto == nil {
		cr, err := crypto.New(int(cif.KLen))
		if err != nil {
			c.log("control:recv:KMReq:error", func() string { return fmt.Sprintf("crypto: %s", err) })
			c.cryptoLock.Unlock()
			c.close()
			return
		}

		c.keyBaseEncryption = cif.KeyBasedEncryption.Opposite()
		c.crypto = cr
	}

	if c.crypto == nil {
		c.log("control:recv:KMReq:error", func() string { return "connection is not encrypted" })
		c.cryptoLock.Unlock()
		return
	}

	if cif.KeyBasedEncryption == c.keyBaseEncryption {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:KMReq:error", func() string {
			return "invalid KM request. wants to reset the key that is already in use"
		})
		c.cryptoLock.Unlock()
		return
	}

	if err := c.crypto.UnmarshalKM(cif, c.config.Passphrase); err != nil {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:KMReq:error", func() string { return fmt.Sprintf("invalid KMReq: %s", err) })
		c.cryptoLock.Unlock()
		return
	}

	// Switch the keys
	c.keyBaseEncryption = c.keyBaseEncryption.Opposite()

	c.cryptoLock.Unlock()

	// Send KM Response
	p.Header().SubType = packet.EXTTYPE_KMRSP

	c.statisticsLock.Lock()
	c.statistics.pktSentKM++
	c.statisticsLock.Unlock()

	c.pop(p)
}

// handleKMResponse confirms the change of encryption keys.
func (c *srtConn) handleKMResponse(p packet.Packet) {
	c.log("control:recv:KMRes:dump", func() string { return p.Dump() })

	c.statisticsLock.Lock()
	c.statistics.pktRecvKM++
	c.statisticsLock.Unlock()

	cif := &packet.CIFKeyMaterialExtension{}

	if err := p.UnmarshalCIF(cif); err != nil {
		c.statisticsLock.Lock()
		c.statistics.pktRecvInvalid++
		c.statisticsLock.Unlock()
		c.log("control:recv:KMRes:error", func() string { return fmt.Sprintf("invalid KMRes: %s", err) })
		return
	}

	c.cryptoLock.Lock()
	defer c.cryptoLock.Unlock()

	if c.crypto == nil {
		c.log("control:recv:KMRes:error", func() string { return "connection is not encrypted" })
		return
	}

	if c.version == 4 {
		c.stopKMRequests()

		if cif.Error != 0 {
			if cif.Error == packet.KM_NOSECRET {
				c.log("control:recv:KMRes:error", func() string { return "peer didn't enabled encryption" })
			} else if cif.Error == packet.KM_BADSECRET {
				c.log("control:recv:KMRes:error", func() string { return "peer has a different passphrase" })
			}
			c.close()
			return
		}
	}

	c.log("control:recv:KMRes:cif", func() string { return cif.String() })

	if c.kmPreAnnounceCountdown >= c.config.KMPreAnnounce {
		c.log("control:recv:KMRes:error", func() string { return "not in pre-announce period, ignored" })
		// Ignore the response, we're not in the pre-announce period
		return
	}

	c.kmConfirmed = true
}

// sendShutdown sends a shutdown packet to the peer.
func (c *srtConn) sendShutdown() {
	p := packet.NewPacket(c.remoteAddr)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_SHUTDOWN
	p.Header().Timestamp = c.getTimestampForPacket()

	cif := packet.CIFShutdown{}

	p.MarshalCIF(&cif)

	c.log("control:send:shutdown:dump", func() string { return p.Dump() })
	c.log("control:send:shutdown:cif", func() string { return cif.String() })

	c.statisticsLock.Lock()
	c.statistics.pktSentShutdown++
	c.statisticsLock.Unlock()

	c.pop(p)
}

// sendNAK sends a NAK to the peer with the given range of sequence numbers.
func (c *srtConn) sendNAK(list []circular.Number) {
	p := packet.NewPacket(c.remoteAddr)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_NAK
	p.Header().Timestamp = c.getTimestampForPacket()

	cif := packet.CIFNAK{}

	cif.LostPacketSequenceNumber = append(cif.LostPacketSequenceNumber, list...)

	p.MarshalCIF(&cif)

	c.log("control:send:NAK:dump", func() string { return p.Dump() })
	c.log("control:send:NAK:cif", func() string { return cif.String() })

	c.statisticsLock.Lock()
	c.statistics.pktSentNAK++
	c.statisticsLock.Unlock()

	c.pop(p)
}

// sendACK sends an ACK to the peer with the given sequence number.
func (c *srtConn) sendACK(seq circular.Number, lite bool) {
	p := packet.NewPacket(c.remoteAddr)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_ACK
	p.Header().Timestamp = c.getTimestampForPacket()

	cif := packet.CIFACK{
		LastACKPacketSequenceNumber: seq,
	}

	c.ackLock.Lock()
	defer c.ackLock.Unlock()

	if lite {
		cif.IsLite = true

		p.Header().TypeSpecific = 0
	} else {
		pps, bps, capacity := c.recv.PacketRate()

		cif.RTT = uint32(c.rtt.RTT())
		cif.RTTVar = uint32(c.rtt.RTTVar())
		cif.AvailableBufferSize = c.config.FC        // TODO: available buffer size (packets)
		cif.PacketsReceivingRate = uint32(pps)       // packets receiving rate (packets/s)
		cif.EstimatedLinkCapacity = uint32(capacity) // estimated link capacity (packets/s), not relevant for live mode
		cif.ReceivingRate = uint32(bps)              // receiving rate (bytes/s), not relevant for live mode

		p.Header().TypeSpecific = c.nextACKNumber.Val()

		c.ackNumbers[p.Header().TypeSpecific] = time.Now()
		c.nextACKNumber = c.nextACKNumber.Inc()
		if c.nextACKNumber.Val() == 0 {
			c.nextACKNumber = c.nextACKNumber.Inc()
		}
	}

	p.MarshalCIF(&cif)

	c.log("control:send:ACK:dump", func() string { return p.Dump() })
	c.log("control:send:ACK:cif", func() string { return cif.String() })

	c.statisticsLock.Lock()
	c.statistics.pktSentACK++
	c.statisticsLock.Unlock()

	c.pop(p)
}

// sendACKACK sends an ACKACK to the peer with the given ACK sequence.
func (c *srtConn) sendACKACK(ackSequence uint32) {
	p := packet.NewPacket(c.remoteAddr)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_ACKACK
	p.Header().Timestamp = c.getTimestampForPacket()

	p.Header().TypeSpecific = ackSequence

	c.log("control:send:ACKACK:dump", func() string { return p.Dump() })

	c.statisticsLock.Lock()
	c.statistics.pktSentACKACK++
	c.statisticsLock.Unlock()

	c.pop(p)
}

func (c *srtConn) sendHSRequests(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		c.sendHSRequest()
	}
}

func (c *srtConn) sendHSRequest() {
	cif := &packet.CIFHandshakeExtension{
		SRTVersion: 0x00010203,
		SRTFlags: packet.CIFHandshakeExtensionFlags{
			TSBPDSND:      true,  // we send in TSBPD mode
			TSBPDRCV:      false, // not relevant for us as sender
			CRYPT:         true,  // must be always set
			TLPKTDROP:     true,  // must be set in live mode
			PERIODICNAK:   false, // not relevant for us as sender
			REXMITFLG:     true,  // must alwasy be set
			STREAM:        false, // has been introducet in HSv5
			PACKET_FILTER: false, // has been introducet in HSv5
		},
		RecvTSBPDDelay: 0,
		SendTSBPDDelay: uint16(c.config.ReceiverLatency.Milliseconds()),
	}

	p := packet.NewPacket(c.remoteAddr)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_USER
	p.Header().SubType = packet.EXTTYPE_HSREQ
	p.Header().Timestamp = c.getTimestampForPacket()

	p.MarshalCIF(cif)

	c.log("control:send:HSReq:dump", func() string { return p.Dump() })
	c.log("control:send:HSReq:cif", func() string { return cif.String() })

	c.pop(p)
}

func (c *srtConn) sendKMRequests(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		c.sendKMRequest(c.keyBaseEncryption)
	}
}

// sendKMRequest sends a KM request to the peer.
func (c *srtConn) sendKMRequest(key packet.PacketEncryption) {
	if c.crypto == nil {
		c.log("control:send:KMReq:error", func() string { return "connection is not encrypted" })
		return
	}

	cif := &packet.CIFKeyMaterialExtension{}

	c.crypto.MarshalKM(cif, c.config.Passphrase, key)

	p := packet.NewPacket(c.remoteAddr)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_USER
	p.Header().SubType = packet.EXTTYPE_KMREQ
	p.Header().Timestamp = c.getTimestampForPacket()

	p.MarshalCIF(cif)

	c.log("control:send:KMReq:dump", func() string { return p.Dump() })
	c.log("control:send:KMReq:cif", func() string { return cif.String() })

	c.statisticsLock.Lock()
	c.statistics.pktSentKM++
	c.statisticsLock.Unlock()

	c.pop(p)
}

// Close closes the connection.
func (c *srtConn) Close() error {
	c.close()

	return nil
}

// close closes the connection.
func (c *srtConn) close() {

	c.shutdownOnce.Do(func() {
		c.log("connection:close", func() string { return "stopping peer idle timeout" })

		c.peerIdleTimeout.Stop()

		c.log("connection:close", func() string { return "sending shutdown message to peer" })

		c.sendShutdown()

		c.log("connection:close", func() string { return "stopping all routines and channels" })

		c.cancelCtx()

		c.log("connection:close", func() string { return "flushing congestion" })

		c.snd.Flush()
		c.recv.Flush()

		c.log("connection:close", func() string { return "shutdown" })

		go func() {
			c.onShutdown(c)
		}()
	})
}

func (c *srtConn) log(topic string, message func() string) {
	c.logger.Print(topic, c.socketId, 2, message)
}

func (c *srtConn) SetDeadline(t time.Time) error      { return nil }
func (c *srtConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *srtConn) SetWriteDeadline(t time.Time) error { return nil }

func (c *srtConn) Stats(s *Statistics) {
	if s == nil {
		return
	}

	now := uint64(time.Since(c.start).Milliseconds())

	send := c.snd.Stats()
	recv := c.recv.Stats()

	previous := s.Accumulated
	interval := now - s.MsTimeStamp

	c.statisticsLock.RLock()
	defer c.statisticsLock.RUnlock()

	// Accumulated
	s.Accumulated = StatisticsAccumulated{
		PktSent:           send.Pkt,
		PktRecv:           recv.Pkt,
		PktSentUnique:     send.PktUnique,
		PktRecvUnique:     recv.PktUnique,
		PktSendLoss:       send.PktLoss,
		PktRecvLoss:       recv.PktLoss,
		PktRetrans:        send.PktRetrans,
		PktRecvRetrans:    recv.PktRetrans,
		PktSentACK:        c.statistics.pktSentACK,
		PktRecvACK:        c.statistics.pktRecvACK,
		PktSentNAK:        c.statistics.pktSentNAK,
		PktRecvNAK:        c.statistics.pktRecvNAK,
		PktSentKM:         c.statistics.pktSentKM,
		PktRecvKM:         c.statistics.pktRecvKM,
		UsSndDuration:     send.UsSndDuration,
		PktSendDrop:       send.PktDrop,
		PktRecvDrop:       recv.PktDrop,
		PktRecvUndecrypt:  c.statistics.pktRecvUndecrypt,
		ByteSent:          send.Byte + (send.Pkt * c.statistics.headerSize),
		ByteRecv:          recv.Byte + (recv.Pkt * c.statistics.headerSize),
		ByteSentUnique:    send.ByteUnique + (send.PktUnique * c.statistics.headerSize),
		ByteRecvUnique:    recv.ByteUnique + (recv.PktUnique * c.statistics.headerSize),
		ByteRecvLoss:      recv.ByteLoss + (recv.PktLoss * c.statistics.headerSize),
		ByteRetrans:       send.ByteRetrans + (send.PktRetrans * c.statistics.headerSize),
		ByteRecvRetrans:   recv.ByteRetrans + (recv.PktRetrans * c.statistics.headerSize),
		ByteSendDrop:      send.ByteDrop + (send.PktDrop * c.statistics.headerSize),
		ByteRecvDrop:      recv.ByteDrop + (recv.PktDrop * c.statistics.headerSize),
		ByteRecvUndecrypt: c.statistics.byteRecvUndecrypt + (c.statistics.pktRecvUndecrypt * c.statistics.headerSize),
	}

	// Interval
	s.Interval = StatisticsInterval{
		MsInterval:         interval,
		PktSent:            s.Accumulated.PktSent - previous.PktSent,
		PktRecv:            s.Accumulated.PktRecv - previous.PktRecv,
		PktSentUnique:      s.Accumulated.PktSentUnique - previous.PktSentUnique,
		PktRecvUnique:      s.Accumulated.PktRecvUnique - previous.PktRecvUnique,
		PktSendLoss:        s.Accumulated.PktSendLoss - previous.PktSendLoss,
		PktRecvLoss:        s.Accumulated.PktRecvLoss - previous.PktRecvLoss,
		PktRetrans:         s.Accumulated.PktRetrans - previous.PktRetrans,
		PktRecvRetrans:     s.Accumulated.PktRecvRetrans - previous.PktRecvRetrans,
		PktSentACK:         s.Accumulated.PktSentACK - previous.PktSentACK,
		PktRecvACK:         s.Accumulated.PktRecvACK - previous.PktRecvACK,
		PktSentNAK:         s.Accumulated.PktSentNAK - previous.PktSentNAK,
		PktRecvNAK:         s.Accumulated.PktRecvNAK - previous.PktRecvNAK,
		MbpsSendRate:       float64(s.Accumulated.ByteSent-previous.ByteSent) * 8 / 1024 / 1024 / (float64(interval) / 1000),
		MbpsRecvRate:       float64(s.Accumulated.ByteRecv-previous.ByteRecv) * 8 / 1024 / 1024 / (float64(interval) / 1000),
		UsSndDuration:      s.Accumulated.UsSndDuration - previous.UsSndDuration,
		PktReorderDistance: 0,
		PktRecvBelated:     s.Accumulated.PktRecvBelated - previous.PktRecvBelated,
		PktSndDrop:         s.Accumulated.PktSendDrop - previous.PktSendDrop,
		PktRecvDrop:        s.Accumulated.PktRecvDrop - previous.PktRecvDrop,
		PktRecvUndecrypt:   s.Accumulated.PktRecvUndecrypt - previous.PktRecvUndecrypt,
		ByteSent:           s.Accumulated.ByteSent - previous.ByteSent,
		ByteRecv:           s.Accumulated.ByteRecv - previous.ByteRecv,
		ByteSentUnique:     s.Accumulated.ByteSentUnique - previous.ByteSentUnique,
		ByteRecvUnique:     s.Accumulated.ByteRecvUnique - previous.ByteRecvUnique,
		ByteRecvLoss:       s.Accumulated.ByteRecvLoss - previous.ByteRecvLoss,
		ByteRetrans:        s.Accumulated.ByteRetrans - previous.ByteRetrans,
		ByteRecvRetrans:    s.Accumulated.ByteRecvRetrans - previous.ByteRecvRetrans,
		ByteRecvBelated:    s.Accumulated.ByteRecvBelated - previous.ByteRecvBelated,
		ByteSendDrop:       s.Accumulated.ByteSendDrop - previous.ByteSendDrop,
		ByteRecvDrop:       s.Accumulated.ByteRecvDrop - previous.ByteRecvDrop,
		ByteRecvUndecrypt:  s.Accumulated.ByteRecvUndecrypt - previous.ByteRecvUndecrypt,
	}

	// Instantaneous
	s.Instantaneous = StatisticsInstantaneous{
		UsPktSendPeriod:       send.UsPktSndPeriod,
		PktFlowWindow:         uint64(c.config.FC),
		PktFlightSize:         send.PktFlightSize,
		MsRTT:                 c.rtt.RTT() / 1000,
		MbpsSentRate:          send.MbpsEstimatedSentBandwidth,
		MbpsRecvRate:          recv.MbpsEstimatedRecvBandwidth,
		MbpsLinkCapacity:      recv.MbpsEstimatedLinkCapacity,
		ByteAvailSendBuf:      0, // unlimited
		ByteAvailRecvBuf:      0, // unlimited
		MbpsMaxBW:             float64(c.config.MaxBW) / 1024 / 1024,
		ByteMSS:               uint64(c.config.MSS),
		PktSendBuf:            send.PktBuf,
		ByteSendBuf:           send.ByteBuf,
		MsSendBuf:             send.MsBuf,
		MsSendTsbPdDelay:      c.peerTsbpdDelay / 1000,
		PktRecvBuf:            recv.PktBuf,
		ByteRecvBuf:           recv.ByteBuf,
		MsRecvBuf:             recv.MsBuf,
		MsRecvTsbPdDelay:      c.tsbpdDelay / 1000,
		PktReorderTolerance:   uint64(c.config.LossMaxTTL),
		PktRecvAvgBelatedTime: 0,
		PktSendLossRate:       send.PktLossRate,
		PktRecvLossRate:       recv.PktLossRate,
	}

	// If we're only sending, the receiver congestion control value for the link capacity is zero,
	// use the value that we got from the receiver via the ACK packets.
	if s.Instantaneous.MbpsLinkCapacity == 0 {
		s.Instantaneous.MbpsLinkCapacity = c.statistics.mbpsLinkCapacity
	}

	if c.config.MaxBW < 0 {
		s.Instantaneous.MbpsMaxBW = -1
	}

	s.MsTimeStamp = now
}
````

## File: contrib/client/main.go
````go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	srt "github.com/datarhei/gosrt"
)

type stats struct {
	bprev  uint64
	btotal uint64
	prev   uint64
	total  uint64

	lock sync.Mutex

	period time.Duration
	last   time.Time
}

func (s *stats) init(period time.Duration) {
	s.bprev = 0
	s.btotal = 0
	s.prev = 0
	s.total = 0

	s.period = period
	s.last = time.Now()

	go s.tick()
}

func (s *stats) tick() {
	ticker := time.NewTicker(s.period)
	defer ticker.Stop()

	for c := range ticker.C {
		s.lock.Lock()
		diff := c.Sub(s.last)

		bavg := float64(s.btotal-s.bprev) * 8 / (1000 * 1000 * diff.Seconds())
		avg := float64(s.total-s.prev) / diff.Seconds()

		s.bprev = s.btotal
		s.prev = s.total
		s.last = c

		s.lock.Unlock()

		fmt.Fprintf(os.Stderr, "\r%-54s: %8.3f kpackets (%8.3f packets/s), %8.3f mbytes (%8.3f Mbps)", c, float64(s.total)/1024, avg, float64(s.btotal)/1024/1024, bavg)
	}
}

func (s *stats) update(n uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.btotal += n
	s.total++
}

func main() {
	var from string
	var to string
	var logtopics string

	flag.StringVar(&from, "from", "", "Address to read from, sources: srt://, udp://, - (stdin)")
	flag.StringVar(&to, "to", "", "Address to write to, targets: srt://, udp://, file://, - (stdout)")
	flag.StringVar(&logtopics, "logtopics", "", "topics for the log output")

	flag.Parse()

	var logger srt.Logger

	if len(logtopics) != 0 {
		logger = srt.NewLogger(strings.Split(logtopics, ","))
	}

	go func() {
		if logger == nil {
			return
		}

		for m := range logger.Listen() {
			fmt.Fprintf(os.Stderr, "%#08x %s (in %s:%d)\n%s \n", m.SocketId, m.Topic, m.File, m.Line, m.Message)
		}
	}()

	r, err := openReader(from, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: from: %v\n", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	w, err := openWriter(to, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: to: %v\n", err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	doneChan := make(chan error)

	go func() {
		buffer := make([]byte, 2048)

		s := stats{}
		s.init(200 * time.Millisecond)

		for {
			n, err := r.Read(buffer)
			if err != nil {
				doneChan <- fmt.Errorf("read: %w", err)
				return
			}

			s.update(uint64(n))

			if _, err := w.Write(buffer[:n]); err != nil {
				doneChan <- fmt.Errorf("write: %w", err)
				return
			}
		}
	}()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		doneChan <- nil
	}()

	if err := <-doneChan; err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		fmt.Fprint(os.Stderr, "\n")
	}

	w.Close()

	if srtconn, ok := w.(srt.Conn); ok {
		stats := &srt.Statistics{}
		srtconn.Stats(stats)

		data, err := json.MarshalIndent(stats, "", "   ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "writer: %+v\n", stats)
		} else {
			fmt.Fprintf(os.Stderr, "writer: %s\n", string(data))
		}
	}

	r.Close()

	if srtconn, ok := r.(srt.Conn); ok {
		stats := &srt.Statistics{}
		srtconn.Stats(stats)

		data, err := json.MarshalIndent(stats, "", "   ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "reader: %+v\n", stats)
		} else {
			fmt.Fprintf(os.Stderr, "reader: %s\n", string(data))
		}
	}

	if logger != nil {
		logger.Close()
	}
}

func openReader(addr string, logger srt.Logger) (io.ReadCloser, error) {
	if len(addr) == 0 {
		return nil, fmt.Errorf("the address must not be empty")
	}

	if addr == "-" {
		if os.Stdin == nil {
			return nil, fmt.Errorf("stdin is not defined")
		}

		return os.Stdin, nil
	}

	if strings.HasPrefix(addr, "debug://") {
		readerOptions := DebugReaderOptions{}
		parts := strings.SplitN(strings.TrimPrefix(addr, "debug://"), "?", 2)
		if len(parts) > 1 {
			options, err := url.ParseQuery(parts[1])
			if err != nil {
				return nil, err
			}

			if x, err := strconv.ParseUint(options.Get("bitrate"), 10, 64); err == nil {
				readerOptions.Bitrate = x
			}
		}

		r, err := NewDebugReader(readerOptions)

		return r, err
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "srt" {
		config := srt.DefaultConfig()
		if err := config.UnmarshalQuery(u.RawQuery); err != nil {
			return nil, err
		}
		config.Logger = logger

		mode := u.Query().Get("mode")

		if mode == "listener" {
			ln, err := srt.Listen("srt", u.Host, config)
			if err != nil {
				return nil, err
			}

			conn, _, err := ln.Accept(func(req srt.ConnRequest) srt.ConnType {
				if config.StreamId != req.StreamId() {
					return srt.REJECT
				}

				req.SetPassphrase(config.Passphrase)

				return srt.PUBLISH
			})
			if err != nil {
				return nil, err
			}

			if conn == nil {
				return nil, fmt.Errorf("incoming connection rejected")
			}

			return conn, nil
		} else if mode == "caller" {
			conn, err := srt.Dial("srt", u.Host, config)
			if err != nil {
				return nil, err
			}

			return conn, nil
		} else {
			return nil, fmt.Errorf("unsupported mode")
		}
	} else if u.Scheme == "udp" {
		laddr, err := net.ResolveUDPAddr("udp", u.Host)
		if err != nil {
			return nil, err
		}

		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	return nil, fmt.Errorf("unsupported reader")
}

func openWriter(addr string, logger srt.Logger) (io.WriteCloser, error) {
	if len(addr) == 0 {
		return nil, fmt.Errorf("the address must not be empty")
	}

	if addr == "-" {
		if os.Stdout == nil {
			return nil, fmt.Errorf("stdout is not defined")
		}

		return NewNonblockingWriter(os.Stdout, 2048), nil
	}

	if after, ok := strings.CutPrefix(addr, "file://"); ok {
		path := after
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		return NewNonblockingWriter(file, 2048), nil
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "srt" {
		config := srt.DefaultConfig()
		if err := config.UnmarshalQuery(u.RawQuery); err != nil {
			return nil, err
		}
		config.Logger = logger

		mode := u.Query().Get("mode")

		if mode == "listener" {
			ln, err := srt.Listen("srt", u.Host, config)
			if err != nil {
				return nil, err
			}

			conn, _, err := ln.Accept(func(req srt.ConnRequest) srt.ConnType {
				if config.StreamId != req.StreamId() {
					return srt.REJECT
				}

				req.SetPassphrase(config.Passphrase)

				return srt.SUBSCRIBE
			})
			if err != nil {
				return nil, err
			}

			if conn == nil {
				return nil, fmt.Errorf("incoming connection rejected")
			}

			return conn, nil
		} else if mode == "caller" {
			conn, err := srt.Dial("srt", u.Host, config)
			if err != nil {
				return nil, err
			}

			return conn, nil
		} else {
			return nil, fmt.Errorf("unsupported mode")
		}
	} else if u.Scheme == "udp" {
		raddr, err := net.ResolveUDPAddr("udp", u.Host)
		if err != nil {
			return nil, err
		}

		conn, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	return nil, fmt.Errorf("unsupported writer")
}
````

## File: contrib/client/reader.go
````go
package main

import (
	"context"
	"io"
	"time"
)

type Reader interface {
	io.ReadCloser
}

type debugReader struct {
	bytesPerSec uint64
	cancel      context.CancelFunc
	data        chan byte
}

type DebugReaderOptions struct {
	Bitrate uint64
}

func NewDebugReader(options DebugReaderOptions) (Reader, error) {
	r := &debugReader{
		bytesPerSec: options.Bitrate / 8,
	}

	if r.bytesPerSec == 0 {
		r.bytesPerSec = 262_144 // 2Mbit/s
	}

	r.data = make(chan byte, r.bytesPerSec)

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	go r.generator(ctx)

	return r, nil
}

func (r *debugReader) Read(p []byte) (int, error) {
	len := len(p)

	if len == 0 {
		return 0, nil
	}

	var i int = 0

	for b := range r.data {
		p[i] = b

		i += 1
		if i == len {
			break
		}
	}

	return i, nil
}

func (r *debugReader) Close() error {
	r.cancel()

	return nil
}

func (r *debugReader) generator(ctx context.Context) {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	s := "abcdefghijklmnopqrstuvwxyz*"
	pivot := 0

	defer func() { close(r.data) }()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for i := uint64(0); i < r.bytesPerSec/10; i += 1 {
				r.data <- s[pivot]
				pivot += 1
				if pivot >= len(s) {
					pivot = 0
				}
			}
		}
	}
}
````

## File: contrib/client/writer.go
````go
package main

import (
	"bytes"
	"io"
	"sync"
	"time"
)

// NonblockingWriter is a io.Writer and io.Closer that won't block
// any writes. If the underlying writer is blocking the data will be
// buffered until it's available again.
type Writer interface {
	io.WriteCloser
}

// nonblockingWriter implements the NonblockingWriter interface
type nonblockingWriter struct {
	dst  io.WriteCloser
	buf  *bytes.Buffer
	lock sync.RWMutex
	size int
	done bool
}

// NewNonblockingWriter return a new NonBlockingWriter with writer as the
// underlying writer. The size is the number of bytes to write to the
// underlying writer in one iteration. It written as fast as possible to
// the underlying writer. If there's no more data available to write
// a pause of 10 milliseconds will be done. There's currently no limit
// for the amount of data to be buffered. A call of the Close function
// will close this writer. The underlying writer will not be closed. In
// case there's an error while writing to the underlying writer, this
// will close itself.
func NewNonblockingWriter(writer io.WriteCloser, size int) Writer {
	u := &nonblockingWriter{
		dst:  writer,
		buf:  new(bytes.Buffer),
		size: size,
		done: false,
	}

	if u.size <= 0 {
		u.size = 2048
	}

	go u.writer()

	return u
}

func (u *nonblockingWriter) Write(p []byte) (int, error) {
	if u.done {
		return 0, io.EOF
	}

	u.lock.Lock()
	defer u.lock.Unlock()

	return u.buf.Write(p)
}

func (u *nonblockingWriter) Close() error {
	u.done = true

	u.dst.Close()

	return nil
}

// writer writes to the underlying writer in chunks read from
// the buffer. If the buffer is empty, a short pause will be made.
func (u *nonblockingWriter) writer() {
	p := make([]byte, u.size)

	for {
		u.lock.RLock()
		n, err := u.buf.Read(p)
		u.lock.RUnlock()

		if n == 0 || err == io.EOF {
			if u.done {
				break
			}

			time.Sleep(10 * time.Millisecond)
			continue
		}

		if _, err := u.dst.Write(p[:n]); err != nil {
			break
		}
	}

	u.done = true
}
````

## File: contrib/server/main.go
````go
package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"

	srt "github.com/datarhei/gosrt"
	"github.com/pkg/profile"
)

// server is an implementation of the Server framework
type server struct {
	// Configuration parameter taken from the Config
	addr       string
	app        string
	token      string
	passphrase string
	logtopics  string
	profile    string

	server *srt.Server

	// Map of publishing channels and a lock to serialize
	// access to the map.
	channels map[string]srt.PubSub
	lock     sync.RWMutex
}

func (s *server) ListenAndServe() error {
	if len(s.app) == 0 {
		s.app = "/"
	}

	return s.server.ListenAndServe()
}

func (s *server) Shutdown() {
	s.server.Shutdown()
}

func main() {
	s := server{
		channels: make(map[string]srt.PubSub),
	}

	flag.StringVar(&s.addr, "addr", "", "address to listen on")
	flag.StringVar(&s.app, "app", "", "path prefix for streamid")
	flag.StringVar(&s.token, "token", "", "token query param for streamid")
	flag.StringVar(&s.passphrase, "passphrase", "", "passphrase for de- and enrcypting the data")
	flag.StringVar(&s.logtopics, "logtopics", "", "topics for the log output")
	flag.StringVar(&s.profile, "profile", "", "enable profiling (cpu, mem, allocs, heap, rate, mutex, block, thread, trace)")

	flag.Parse()

	if len(s.addr) == 0 {
		fmt.Fprintf(os.Stderr, "Provide a listen address with -addr\n")
		os.Exit(1)
	}

	var p func(*profile.Profile)
	switch s.profile {
	case "cpu":
		p = profile.CPUProfile
	case "mem":
		p = profile.MemProfile
	case "allocs":
		p = profile.MemProfileAllocs
	case "heap":
		p = profile.MemProfileHeap
	case "rate":
		p = profile.MemProfileRate(2048)
	case "mutex":
		p = profile.MutexProfile
	case "block":
		p = profile.BlockProfile
	case "thread":
		p = profile.ThreadcreationProfile
	case "trace":
		p = profile.TraceProfile
	default:
	}

	if p != nil {
		defer profile.Start(profile.ProfilePath("."), profile.NoShutdownHook, p).Stop()
	}

	config := srt.DefaultConfig()

	if len(s.logtopics) != 0 {
		config.Logger = srt.NewLogger(strings.Split(s.logtopics, ","))
	}

	config.KMPreAnnounce = 200
	config.KMRefreshRate = 10000

	s.server = &srt.Server{
		Addr:            s.addr,
		HandleConnect:   s.handleConnect,
		HandlePublish:   s.handlePublish,
		HandleSubscribe: s.handleSubscribe,
		Config:          &config,
	}

	fmt.Fprintf(os.Stderr, "Listening on %s\n", s.addr)

	go func() {
		if config.Logger == nil {
			return
		}

		for m := range config.Logger.Listen() {
			fmt.Fprintf(os.Stderr, "%#08x %s (in %s:%d)\n%s \n", m.SocketId, m.Topic, m.File, m.Line, m.Message)
		}
	}()

	go func() {
		if err := s.ListenAndServe(); err != nil && err != srt.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "SRT Server: %s\n", err)
			os.Exit(2)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	s.Shutdown()

	if config.Logger != nil {
		config.Logger.Close()
	}
}

func (s *server) log(who, action, path, message string, client net.Addr) {
	fmt.Fprintf(os.Stderr, "%-10s %10s %s (%s) %s\n", who, action, path, client, message)
}

func (s *server) handleConnect(req srt.ConnRequest) srt.ConnType {
	var mode srt.ConnType = srt.SUBSCRIBE
	client := req.RemoteAddr()

	channel := ""

	if req.Version() == 4 {
		mode = srt.PUBLISH
		channel = "/" + client.String()

		req.SetPassphrase(s.passphrase)
	} else if req.Version() == 5 {
		streamId := req.StreamId()
		path := streamId

		if strings.HasPrefix(streamId, "publish:") {
			mode = srt.PUBLISH
			path = strings.TrimPrefix(streamId, "publish:")
		} else if after, ok := strings.CutPrefix(streamId, "subscribe:"); ok {
			path = after
		}

		u, err := url.Parse(path)
		if err != nil {
			return srt.REJECT
		}

		if req.IsEncrypted() {
			if err := req.SetPassphrase(s.passphrase); err != nil {
				s.log("CONNECT", "FORBIDDEN", u.Path, err.Error(), client)
				return srt.REJECT
			}
		}

		// Check the token
		token := u.Query().Get("token")
		if len(s.token) != 0 && s.token != token {
			s.log("CONNECT", "FORBIDDEN", u.Path, "invalid token ("+token+")", client)
			return srt.REJECT
		}

		// Check the app patch
		if !strings.HasPrefix(u.Path, s.app) {
			s.log("CONNECT", "FORBIDDEN", u.Path, "invalid app", client)
			return srt.REJECT
		}

		if len(strings.TrimPrefix(u.Path, s.app)) == 0 {
			s.log("CONNECT", "INVALID", u.Path, "stream name not provided", client)
			return srt.REJECT
		}

		channel = u.Path
	} else {
		return srt.REJECT
	}

	s.lock.RLock()
	pubsub := s.channels[channel]
	s.lock.RUnlock()

	if mode == srt.PUBLISH && pubsub != nil {
		s.log("CONNECT", "CONFLICT", channel, "already publishing", client)
		return srt.REJECT
	}

	if mode == srt.SUBSCRIBE && pubsub == nil {
		s.log("CONNECT", "NOTFOUND", channel, "not publishing", client)
		return srt.REJECT
	}

	return mode
}

func (s *server) handlePublish(conn srt.Conn) {
	channel := ""
	client := conn.RemoteAddr()
	if client == nil {
		conn.Close()
		return
	}

	if conn.Version() == 4 {
		channel = "/" + client.String()
	} else if conn.Version() == 5 {
		streamId := conn.StreamId()
		path := strings.TrimPrefix(streamId, "publish:")

		channel = path
	} else {
		s.log("PUBLISH", "INVALID", channel, "unknown connection version", client)
		conn.Close()
		return
	}

	// Look for the stream
	s.lock.Lock()
	pubsub := s.channels[channel]
	if pubsub == nil {
		pubsub = srt.NewPubSub(srt.PubSubConfig{
			Logger: s.server.Config.Logger,
		})
		s.channels[channel] = pubsub
	} else {
		pubsub = nil
	}
	s.lock.Unlock()

	if pubsub == nil {
		s.log("PUBLISH", "CONFLICT", channel, "already publishing", client)
		conn.Close()
		return
	}

	s.log("PUBLISH", "START", channel, "publishing", client)

	pubsub.Publish(conn)

	s.lock.Lock()
	delete(s.channels, channel)
	s.lock.Unlock()

	s.log("PUBLISH", "STOP", channel, "", client)

	stats := &srt.Statistics{}
	conn.Stats(stats)

	fmt.Fprintf(os.Stderr, "%+v\n", stats)

	conn.Close()
}

func (s *server) handleSubscribe(conn srt.Conn) {
	channel := ""
	client := conn.RemoteAddr()
	if client == nil {
		conn.Close()
		return
	}

	if conn.Version() == 4 {
		channel = client.String()
	} else if conn.Version() == 5 {
		streamId := conn.StreamId()
		path := strings.TrimPrefix(streamId, "subscribe:")

		channel = path
	} else {
		s.log("SUBSCRIBE", "INVALID", channel, "unknown connection version", client)
		conn.Close()
		return
	}

	s.log("SUBSCRIBE", "START", channel, "", client)

	// Look for the stream
	s.lock.RLock()
	pubsub := s.channels[channel]
	s.lock.RUnlock()

	if pubsub == nil {
		s.log("SUBSCRIBE", "NOTFOUND", channel, "not publishing", client)
		conn.Close()
		return
	}

	pubsub.Subscribe(conn)

	s.log("SUBSCRIBE", "STOP", channel, "", client)

	stats := &srt.Statistics{}
	conn.Stats(stats)

	fmt.Fprintf(os.Stderr, "%+v\n", stats)

	conn.Close()
}
````

## File: crypto/crypto_test.go
````go
package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/datarhei/gosrt/packet"

	"github.com/stretchr/testify/require"
)

func mustDecodeString(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return b
}

func TestInvalidKeylength(t *testing.T) {
	_, err := New(42)
	require.Error(t, err, "succeeded to create crypto with invalid keylength")
}

func TestInvalidKM(t *testing.T) {
	c, err := New(16)
	require.NoError(t, err)

	km := &packet.CIFKeyMaterialExtension{}

	km.KeyBasedEncryption = packet.UnencryptedPacket
	km.Salt = mustDecodeString("6c438852715a4d26e0e810b3132ca61f")
	km.Wrap = mustDecodeString("699ab4eac6b7c66c3a9fa0d6836326c2b294a10764233356")

	err = c.UnmarshalKM(km, "foobarfoobar")
	require.ErrorIs(t, err, ErrInvalidKey)

	km = &packet.CIFKeyMaterialExtension{}

	km.KeyBasedEncryption = packet.EvenKeyEncrypted
	km.Salt = mustDecodeString("6c438852715a4d26e0e810b3132ca61f")
	km.Wrap = mustDecodeString("5b901889bd106609ca8a83264b12ed1bfab3f02812bad65784ac396b1f57eb16c53e1020d3a3250b")

	err = c.UnmarshalKM(km, "foobarfoobar")
	require.ErrorIs(t, err, ErrInvalidWrap)
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		keylength   int
		salt        string
		passphrase  string
		evenWrap    string
		oddWrap     string
		evenOddWrap string
	}{
		{
			keylength:   16,
			salt:        "6c438852715a4d26e0e810b3132ca61f",
			passphrase:  "foobarfoobar",
			evenWrap:    "699ab4eac6b7c66c3a9fa0d6836326c2b294a10764233356",
			oddWrap:     "ca4decaaf8d7b5c38288e84c8796929c84b7c139f1f769d5",
			evenOddWrap: "5b901889bd106609ca8a83264b12ed1bfab3f02812bad65784ac396b1f57eb16c53e1020d3a3250b",
		},
		{
			keylength:   24,
			salt:        "e636259ccc41e73611b9363bb58586b1",
			passphrase:  "foobarfoobar",
			evenWrap:    "8c6502d6a83e0ab894a43cb5b37b71c2755afc64a682bed9d46912138b60f384",
			oddWrap:     "1fe56a4475759636674a7c5e44f1cdfb365a9f11d8fe74536e8df6b97eecf1c9",
			evenOddWrap: "7360357d363ebec384885b10c8120528889d1be05624bfc381c5fa090f00f9ecef5d6427f7542a58be144f4aeb07452beca546874a68197d",
		},
		{
			keylength:   32,
			salt:        "3825bb4163f7d5cf2804ec0b31a7370f",
			passphrase:  "foobarfoobar",
			evenWrap:    "7d1578458e41680dd997d1a185c75753f3344c6711542b35833f881f7c480304cbe9bdbe76035914",
			oddWrap:     "cc1af097af558fa25b925417c4e6e9e1adacd8b96916b4ac4fac8e6ecdc3b5c48c01134e92e9e5f6",
			evenOddWrap: "f7373def4e9f61f6cd6a22e78916aa07cac8e5f07669d556ec8a15b7631fa9c631e9d98a3f92dbe187f434569ec71b9e2a53171feafd909a5560233fe02ed0301e576d4992b10c86",
		},
	}

	for _, test := range tests {
		c, err := New(test.keylength)
		require.NoError(t, err)

		km := &packet.CIFKeyMaterialExtension{}

		km.KeyBasedEncryption = packet.EvenKeyEncrypted
		km.Salt = mustDecodeString(test.salt)
		km.Wrap = mustDecodeString(test.evenWrap)

		err = c.UnmarshalKM(km, test.passphrase)
		require.NoError(t, err)

		km.KeyBasedEncryption = packet.OddKeyEncrypted
		km.Salt = mustDecodeString(test.salt)
		km.Wrap = mustDecodeString(test.oddWrap)

		err = c.UnmarshalKM(km, test.passphrase)
		require.NoError(t, err)

		km.KeyBasedEncryption = packet.EvenAndOddKey
		km.Salt = mustDecodeString(test.salt)
		km.Wrap = mustDecodeString(test.evenOddWrap)

		err = c.UnmarshalKM(km, test.passphrase)
		require.NoError(t, err)
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		keylength   int
		salt        string
		evenSEK     string
		oddSEK      string
		passphrase  string
		evenWrap    string
		oddWrap     string
		evenOddWrap string
	}{
		{
			keylength:   16,
			salt:        "6c438852715a4d26e0e810b3132ca61f",
			evenSEK:     "047dc22e7f000be55a25ba56ae2e9180",
			oddSEK:      "240c8e76ccf3637641af473edaf15aaf",
			passphrase:  "foobarfoobar",
			evenWrap:    "699ab4eac6b7c66c3a9fa0d6836326c2b294a10764233356",
			oddWrap:     "ca4decaaf8d7b5c38288e84c8796929c84b7c139f1f769d5",
			evenOddWrap: "5b901889bd106609ca8a83264b12ed1bfab3f02812bad65784ac396b1f57eb16c53e1020d3a3250b",
		},
		{
			keylength:   24,
			salt:        "e636259ccc41e73611b9363bb58586b1",
			evenSEK:     "4dca0ad088da64fdc8e98002d141bc46fed4fa0167b931c8",
			oddSEK:      "2b2bbb64ee3942cfa31bfe58efd1d2102c40b7bc028f8946",
			passphrase:  "foobarfoobar",
			evenWrap:    "8c6502d6a83e0ab894a43cb5b37b71c2755afc64a682bed9d46912138b60f384",
			oddWrap:     "1fe56a4475759636674a7c5e44f1cdfb365a9f11d8fe74536e8df6b97eecf1c9",
			evenOddWrap: "7360357d363ebec384885b10c8120528889d1be05624bfc381c5fa090f00f9ecef5d6427f7542a58be144f4aeb07452beca546874a68197d",
		},
		{
			keylength:   32,
			salt:        "3825bb4163f7d5cf2804ec0b31a7370f",
			evenSEK:     "53a088d93431181075f8a9bc4876359afe48967308120c93f97bbd823d8de62a",
			oddSEK:      "7893e88b6296ffcc5a2eab5f53d48efd7adaeced8cb3a851d4f8e2dbda8db17a",
			passphrase:  "foobarfoobar",
			evenWrap:    "7d1578458e41680dd997d1a185c75753f3344c6711542b35833f881f7c480304cbe9bdbe76035914",
			oddWrap:     "cc1af097af558fa25b925417c4e6e9e1adacd8b96916b4ac4fac8e6ecdc3b5c48c01134e92e9e5f6",
			evenOddWrap: "f7373def4e9f61f6cd6a22e78916aa07cac8e5f07669d556ec8a15b7631fa9c631e9d98a3f92dbe187f434569ec71b9e2a53171feafd909a5560233fe02ed0301e576d4992b10c86",
		},
	}

	for _, test := range tests {
		c, err := New(test.keylength)
		require.NoError(t, err)

		cr := c.(*crypto)

		cr.salt = mustDecodeString(test.salt)
		cr.evenSEK = mustDecodeString(test.evenSEK)
		cr.oddSEK = mustDecodeString(test.oddSEK)

		km := &packet.CIFKeyMaterialExtension{}

		err = c.MarshalKM(km, test.passphrase, packet.EvenKeyEncrypted)
		require.NoError(t, err, "keylength: %d", test.keylength)

		wrap := mustDecodeString(test.evenWrap)

		x := bytes.Compare(km.Wrap, wrap)
		require.Equal(t, 0, x, "keylength: %d", test.keylength)

		km = &packet.CIFKeyMaterialExtension{}

		err = c.MarshalKM(km, test.passphrase, packet.OddKeyEncrypted)
		require.NoError(t, err, "keylength: %d", test.keylength)

		wrap = mustDecodeString(test.oddWrap)

		x = bytes.Compare(km.Wrap, wrap)
		require.Equal(t, 0, x, "keylength: %d", test.keylength)

		km = &packet.CIFKeyMaterialExtension{}

		err = c.MarshalKM(km, test.passphrase, packet.EvenAndOddKey)
		require.NoError(t, err, "keylength: %d", test.keylength)

		wrap = mustDecodeString(test.evenOddWrap)

		x = bytes.Compare(km.Wrap, wrap)
		require.Equal(t, 0, x, "keylength: %d", test.keylength)
	}
}

func TestDecode(t *testing.T) {
	packetSequenceNumber := uint32(0x79ee189e)
	data, _ := hex.DecodeString("47410011000001e00000808005211651aa410000000109f000000001419a2088dc6f8003fd0320993a9d4f33dd3a34628addfddfeffd03bf8f8677d149a74e8b1e56fd8208df79622c98c2edd137b6ff2331378c2c1299a2643d4048dd5e7062159a71fceacc10e61138f1d4e8788051c2f0f000fd1213d43ad36bd05682297ec6071bbfa83864a3436e813118a71e44797f3e12c22fc9f77e03455d5bc1405a001cae497adbff3f36004d122e5b77bfba7d391a6a830ef2235cf40747010012fe062c21e108d840187080300f6197be0c04ecb77bc455f686e23867fc2f802fe36677abf008cc43760db57d600afafd3fdad2f523ce844cf7ba3668f3e2ffff085b9c20baf084b810031e0400e7a094d9f114806e003ef32a9ef7684ddbef9b0007277e9524458b687c0801ae379e360c06f9c73130e90da9fff0201e3e81f7f5f0f35ff81016b8435c12e11d52f8342465d3a0de082f87381e0ccc83c1999040918a100908fc0e0428897c0e010e225e18060083010cc847010013cab57e091b4ff7f20697809cc00441b4ec99a00906ee73f795577dd2fef700111bbb5b8002f73beeb9d8a5debd5dd821020066420423209d2222ad5f8499e05ac0a6736b460c5d3c38312fb98cbcbc3e08023f0ffeabc09c285104a5cf89d002fe006c4dfa933a12aeabe60e0016689962dabb48ffefc3011e19c3f8200c7003b226030ae2036d41f4d00c23c8ad501344ee4ad41c4cd808210406da830301a7011c007977125ea5e4b3ef78783881e52841bfaf80065a6b47010014e72d60801017eb5c10020573e36020a4947bfd7c5d66fa85882e06ac5602bb987c92f0446c1c21883013b98ea1db881fa00bd7a5326b80b68c14c43856cafff000cebf7bc7ee3f000f38a4ccbb0ab620f78630085f575d4c85907f20276aa3df30835a0fc1673d74bba557609fabafdefe1f82dac64556e4155409b77275bfbff072105b9c100478679de1c82102070400c0bddf0400c382085eb5ae0084c88020e4915a80f4cc1c0026999a12652ab37e11863f00800e264701001546021dc486dab21021d062c81e8fff8c9461b27a997aa6d9a88187003e3aacd93a67bdeb5f249fc38a002d3607c66dd9b7d7c200132200133c811f612017e2004ce2004cf030001898220002010870e0104112c300192418a81f80885b482ed470a4907da0380108225f60a44452123200491a1b9486026ed55e484c42180bbb5e1983d25dfe03b5f7ffffff6830fbf17b3d9b33030f3f9fcef0cc200c3821020028e6c59b05e62e080180b91517eef77c210af0038ed823470100161881b6af0045260ce232b57f7104d0ca0007cd483569fe4cbc44b20992fff750c300186bebfb0677fb4fdfabffc37e0074a6606339e1b6a0972434060014f49331dcc8ee301df808bfd3e77fbbdeb9eb0f0c347f043cefe0c782085604e5a6825cfc0554d4cbe411286bb3f2cc06ab7fdfbf813ec7e45fe648fb8a216b6a26bfdf0e7f3f9de17820061c10030e1085f0042d20215acad5eef733834006419dd5def7e0032daeeb5deff867e0085a6085636b57e6418fdd354701001767fbafd3063f3a74cff7fb2c188406c5d5555545d45c998b8e32e3b1de39355fe10866eef5f849efbd00fc210d600a740661595abe01244c1800622d32398baa43c11fe769cf0843540c075b7dd0170089bebe7ff80073446b69b22df7f08433a0c07a4bba03ad7ef5f88ffe05d9a71348ce7dff9d73ac290400c3820060b5e0c017f03820416cb06009597c3189040f89003613152b3b7da0033850ab2322fb41c02f7353c002a9089cc4111f683c210ce008a6c188c5d6")
	tests := []struct {
		keylength     int
		salt          string
		evenSEK       string
		oddSEK        string
		passphrase    string
		evenEncrypted string
		oddEncrypted  string
	}{
		{
			keylength:     16,
			salt:          "6c438852715a4d26e0e810b3132ca61f",
			evenSEK:       "047dc22e7f000be55a25ba56ae2e9180",
			oddSEK:        "240c8e76ccf3637641af473edaf15aaf",
			passphrase:    "foobarfoobar",
			evenEncrypted: "6053ce41905e451494d156b53c53fc6372c7e0eea81b10c4f21d4624915a93a368605cd958b14bf7274f1f75254ca606c464ea8a3f34e766fdee34dc6886512bf9b3916a3a24a3febe559cb9a07cfa5a0ec56ffee9b11605a97aa684736afcdad091fb467ca2a6ccce7356e8b2cd931f03f6c0ee4bccdd9fc6f635b080274f32f6b64a14b745efcfd7daaf05fc75369c1e37892d7f97dca5526eef2eba914cb729689b879c5a3f9a8fff061dda8f930eb20e92f241524c3743e12dd8e9b50937d73212307c2c47cb14834a5c2ab6fe1f68b46b2e1ca85cbecf00cf176a426d819ec9d41ed525f9ee7c2e7ee775cebc98357c55de1991ee23a979a4072672c4add4e215966ccef2bd04ad78f8822fcb3ed3a75d945a6c0d3dfa84b3f80c107cc94a0fb5b0b97ff71b8c8ab816ade945fb0ac70228f30e17565cd09b849a06746383193e55105904543b6079eecc53cf51d6d47690871f3f569501f0bcc790fc4c4ab344aedb53a5ef4ed19c8f5c43179f3c1de7065c32182e1ebe291809064c3cbc9ff69f92a61f3642e7ec5d711ef6a07556ea375550d6bf4221093a85b1b6bb8fd650ab3acd40f409497fdf08b88b718b408574ff9f6e003fe64803a34d5e4d86d361c5754741b60024ae24035f01f1a1e2f6dbb0aa93794fac2afa7fc824933260954b05845e856db7b7720ed0dd08504f50a4469d383b9d552e048bd674a98dfaa137bf15b740388ce0f8bcde80046d4ddc0e5987315183950477f069796201e97c990eb98370cd8f5f3bb99704d0806b59d31d02f5d0233875c27f9b132d6a7fdc3508a264b2d33a570f56aa4388e512a4181e9b0caeae8b368224d4b4ba499ef0cc7daf659d3bd38aa238577dde5d19ce1eb5257198096c00f90ab02a8c7320c6b1d10297fabf2e5f62af421a89cbaaa1e4a4e120d0e888354610c47d52a32e78f4ae21475d2201c74460d088e7686d00e7333a7c259a73944cd7981522205bec96b87157ae5fd5473e77d900f57c7edb170bc69d1e736760d63bf31c845f002259ab495c069cb39b074a066a9ddc718f3c6b94bde6fe0c82ace2d9c748f060435a1d4719651858a2f3c3ddf974035cbf1fe248fc3ee479ab37f064f019174df80f79b4bd7beb691f995358779fd124886158e669b118bbae7fac87a13558dc94f0f2dee89abe6e860082bca2fddd3822c193708b37cd53dfbeaab437a7654490bd55a29ca79e8de4f78f6c9c24aee40944d6c0e4f3cc4a5530683c087741b3ef36b464d11e207bc9b1e2ebddce33f3a706d7b7d83603b604a054d02ac0fac6d4194887b30c542e4f44a8df70d4ed53e47f019ab83dde4f251011b8fbe381ca6145547a6a9a340855d1020c078b8152a71f199653744c5c73e9910d601ee74ea6533393261b6a94ab356b3d028d05b47ce1a31cdf5acf3088d4a835285b42b7ea06cff52c98abc099fc0df611588a8941380e5ff65938aa9212ddc0d5a0b3f229615868a0654a1fc2bb0ac91f32ee802f2b316db20b758b6c7373f978f6efe7a6b8e1ffae4061e2a43f108965fdc7e60509d6896f299ad6bd3bfa5adc4a5c968878aecb8ce49ae278d613defcd5bc462c311f04093571af0f6f01484833c8977d3900e4b42085882b66d783edfeb3efd782cd6e44ef38bd1c1ce9dc427fa071caaa26924978530b3002d7e33722cbc87f1b2d670835960071763b712229740a7613da5236e10caa2466384306537d8253417a7f5bfee629129a8e52a52926a2b65f7c0c903c3762988350777a99abbeb34eeb49ebdb5a00c2bef02387651773e8f888b4f672c9f926aff5df12a899599fc40e5f95fc5394d615cafc8ef0831e21c7",
			oddEncrypted:  "8f91686a120b4461f0e078ce80805cb51994b15d31f126ece327469f326c337f25d06e67fd9209308c933379aa796d6c3c7d3dee6410e413829ddb6d6327026c00bb4000992c54f185abf9d724c772d32b8a38ac094b39c948e46c1b06e4263fbf7c7318ef3667937e17a63579decfe667c8478284fab330c500f5f9266922a35a30f612cce0af0cb8f5ec1d3b393339955f7c6470380b9cf7e5ed29b83d296895f0308805e3e5acd0c5e8f54535223e61845692a7e52763a751e0b14633ebe73fe7ec215ffb3e45bcc777661d74efdb1631e5b8fa7ab3c7ec59f2c1f2399e90479b42557b73a14288792609eab84929152216c8969e3ff7f5a17c810ce31f0387448b9a3fc5901302f295d43c98efaceb2f3a0df2f0c6ccd328da26e6c836527ac19641dbe36f4a47a05f62ea0dd411a34f5f73a86cc9daf6e58eed9d630313687e5dd06d51a92668149cb0395a3e42bdd8174396f576ae88a574f933d7ccedd152a9981f1dbd2395be9647f9b0291de1a2a1dffbf3aa8739906dd47ddeaa4978ef50762bee7df61594da9c34109da985a611aa2c9a8a3e6d388a27d1378a85496b3ba54334b559477cecbc844319768c1ded16949dd4298a0b04e6ae478953bf1a18d98fb1db3128ca2997402b8d2f5110e4c238291fbed9e01babfcaa3117f1dbd0113ae5eaff9e0c75db87d2166cd3d67db5e4bb01badd69f3d6464df2029e51438882ca99dc038bf0f0811f3773490edf9ff75f63ac97dc41656035c19c943c094cf25c1c07a5db82de71f79f94075aab064bfb8dc1a4f434e0f8d24ae775f0c773098d5bef537dcc87f8e2a91e71b87ac6320c3619c612709ee7c8077d67a5f5768503445b08aa9feef79bf6318d7abab7feb9382ab0fcfc9835ea0f94ae3c554ba9aeb447b4c05028672a89fac1370b2ce1c438d2311474f91ee97f386d06369210b52eb10237b5ac692b944ff637920fb9aa258680ebfba9df52588f2edf086925403d9bcc3e92ea513131e42654dbe42a092825d5e5b96474080e24b9fb51624427a8d2fae30dba7f249520080b80ef8709aa50285c0d5cbf144df419d85018bc458c6a3f60d44c94503461db6e53de0723da92e2e89eb2e6c98dfac7b7d9f2e542f414bae4234683ce8ca33761d5e149380e417c779c63c157c98850d37ec4711bdb1e4e2a741c4195e4c5dd18274450d340755cde38a48d3004bcd12f9af936d916fa2ee855c4a4fc590d849460e690cd76bb0bc823dae428e831cd98b8cb405d9d2427951e37886855dc3dec94a9c5d32abb4517b56950af31f89b4e29f9e6edb7bb6f24a12af7bad41f012e79da1cec774dcc241dd110f810d6c9263e1bf5711eb4af1c5ab5f0bfb15fadc442124aa55db8d83454fab030f727bd4512bb0aa6d128637025442e0f0cec76fdf97c467f2b5072bbec7c44376becce9bf425725f433e13b4a335a549b1cbf05c51cb87907749b67f46be67339870ed1b3e5e5a356e4c1b36baf9bd28a58aa9bf1f66839947f1ffbe3331263b705798d7e3d1ad86f4a39169c874fe793af83680146cbef8d604e1f707ff16f90c2b707acc246d40bedce5b1a86023ed046258d3715104795edca4e1476d59bafd9e25e8b4dd087b2748dcbb0c4965b6a310e0639f6e7d3f1a0917fc4902bca1e908526c663dd259c3ea32f3abc3959746dd171179c931586a3725bfd0e4b814388afea817501113eab47555db0a470e724020cd57b158a49ba9ed340e6d8f1a0b67666dc4c31d02485b97354816a5c3c401a676005b603af3c632a2370b46b80b28ecebe3437e8dc57cdd299c702f7aff2b536eb920f402504590236d515fec68a17d02f6c3",
		},
		{
			keylength:     24,
			salt:          "e636259ccc41e73611b9363bb58586b1",
			evenSEK:       "4dca0ad088da64fdc8e98002d141bc46fed4fa0167b931c8",
			oddSEK:        "2b2bbb64ee3942cfa31bfe58efd1d2102c40b7bc028f8946",
			passphrase:    "foobarfoobar",
			evenEncrypted: "97888ba6bd9d4015a6f62c4d3ad9c2674077218fb83ba213a634bef143186587debe76466eef8218f2b3aeae2cf00d5c46b28ffa0d325fb2ac4409fb97921e1f4c4a652d731ec6f608c92006a0ef8abe203dcf1c4f81795b126b2aff3a46d6c69a047a050a385417ecf9bea11530346607d8db1d821baf49b36b772160a1633257333b8c544fb0f4d25da28dcdb1e24a8611968053b4e636440756c7a8bf6fb7b9cb3bbc34a54e0b88c6651ac23192d18327d46e7aaf6c01a038346cb65f59d4f6be2d3a5f92c1bd043f236c0c9dbb1dad4994f426ee8ebd05a62c7558c384de5be20018ac50c73efd4ddc1255419c766f3cf53be6ad5fe35682208538e447abf0a68b1ba7d7030ace2a6b20d73bc71dd3ac7b026d60220f1eff05a66c5e06102e2801c8bd9d60091489f282eab072670f1e667bdbc2a067ddd62ffcdc82d9d78cd491277beed96f9fa87de05e6ddbbc7a79557fddd610a58c20187b71bf46f0e2017065b48689efd2862ed3bce11e0792ccfcf64494e34277be21fee874f15dd5377695b69ae1a1cf5413502ea182eea4cd4bbd8d29923aa357e05c7b974ab9df120dbaae26f282f2cc14bcb6a7075f678cf2ad3f39ce8cb01bd44bbeb60269ada55735c88d2355839267e7f9f7596e4baedd2f425d05cac9df2d3b8d6105fb9826a55c78438be8579c02664ec8bac3bde88513cb4019c2a57ad0044730d7f48568bb47ef15aae7949a8be5950421d4de73a6ad27397c0f19da4739acd7527648a8b643ecbf2451a2ac61d6e7666764f657f1003a981fa3f7ebd8db91f286b3b33a764b3542c99017f56b48db5e5f1061eb9327d137b5d6e8e86203de8501103b0fccb0cb97d0266778ea439e1432326b719a7f9396f1c773e6543b391c926920282faadf37983e2fa0dc368e2d384466981ec165df62df4386a618e5cd2a8bdd28d03a5688c38bcb00b2449c39e35f8113253c339e7e4b97048886a8046e686b97591e9b3507a9f10b9c6ca9b6b8546dbade8e42a5af607a35e99f31c06838099c6f5df0907b0f09a33216318118c03bb11374cf2a3cf2ff03978dd66b7b95dff36bfc0bf6c10e7de490e9ed6f9ec226275c59da836f405bb70f504e6b4d1140365d714b84b18a735a4f0de2b5ed7e4fc5ddaf8812d52dc6b49740b7673f408a2b64e4a05f05349706fdfea71914ae29f614fa58d7ad494f8e72de1f9ce579a0fa595a4cbe112effb769bb0e146c20f9aef08a72aff74f9a7d0404d802f9705428776bf660abce214ace695310ec32d5caa95d84e15c102032f40f7d51f409b285d272864e8e8f075e5e0d0034b0405a54c1b0946e1cbff5438fd7ed54544ab3b5e1382749b1949daac5921bee6cee31e377a331400dd6d8baefe84598d2d09159614b4a3e89de69d2b951470d8ab77534e152e04297b54418d73989e1c459c88c38c07bf881bf3e9421b1b22cf1572e969251f63613b3b696f391756ac48457748112ca41882fd6b21ec87bd0027d13a158ab224adab8e66375ba044239dec6dc6a67566a03da021fec02cf295f85eec4adec4b6583eb013b2a3505c37dfcf60de3d9376661760499b03ca718e150dc43825d93ed82c575ecc50c9d5353f66b5f37ea50b202777c7e3f8bbb7b776c98688f2505eceb735e1879d149982306c20ebb1fab4d81a1382f2d42277a7357a9a4bd4d6aaa60adb7fbbc6026710a6162f553938d4a42b2213d626f393c8b8cb7b0583e29d53b02d93f2ba47cdc1af9adc497c5b4285d73120965c44a05a1553cdaedeb615aa506e6d509aace22c2d285c0edb41c93e5041e55d65ee1b74fbcd735794fe0678c98e9d3b66ae92ef2768dee78b5",
			oddEncrypted:  "232f7ccc827d309606b5d879ae55f3a7cce1bbb6a50d5d3fc4fdfc84a57687a2783ccf101ed666809707f57d611b8754af176edc336e4535047de6d22d97c094d126805d9ce17704cbebc84596047e86aaea51e0e268fddc76c2b78e2e3dc0cde4dd11ee383ae2d24fc8151de48b05360612d9f68c864bd4fe1933af1b979ed44b721cbe84f20f1e415219a88406e4110ed20a441feb045d72bf7684a4c7b91fdde8f39e9d92aec62c292f588de5b6e29162f0e9c916cd316f6ac1bb67d781eef73ff6c34e73325900558692b5a8fad10fc25b55835fa1f0a188d1b310f7ac146567abd8ec4ff34c9e421a1d8ca7202d5b086b9c6bc91401b018e25fa27b79810b91e41e923a8e34aac3a50bc79ec4f680284fcc7334a43225a187a44eab0795b09179d9acdb5a307b043d14b30fb2956bdb39d1a1acace0c9f6f1efbab1b3b6cb5eb5b6955daaae972aaddc69a23cbe0d427598319be0b2434fed8fc680e1f185ec953afe59c83c197d2637d1646e3b9ee81b262995e4a87fc901a0200cc109add2d0f9d3fdf2c28b0e08f2e7e1b777e9e65198726f0cc32d49d6505da3d2f520c24d6ae25238ae8f546cf36c60d2f531b501828b3b2472b922e3eb35e49862a5afb9ee16fdfa0c8f550ba843e9ffd459ec95f0773143db6124e7614ee92c52420fef4afeb38f1889a00f272d744d640f8299da75691353df80d902581cb4f5e8a6a49346fb5479a8591788b1e29b251b661a1581253aaa009874e61f4a90c332828c9698ce8b8b5a844464fb024236adeeedcf2495794fba71faaa91f8495d61f5d94bd4ea77f8449317d59684d738331c513b0c673db453ed6febc4c2b22ffaea1bbc605b9b18cd97bf8b13d23dbe0c1eac7b783d27f31c280ff6a4ed0abd365ffd9094b6b70d8664242b0a56980aa43b861611c7cdc53fd655d4ce7fd85d2b694b4ccbfadc20705bf647989772d13368cd6d8564d65e2efad39b2c3892ee1056c658dc1fba3cb25e9f3754bb19f9c62c84899ae222aa72d331a58c7099c0b962b4086979492b71079808e55b3a1428e363ef07657dfeb380000ef375090d68fa96eed42a2666d9d674b5e89bb1b46ae20f2f34f1f73a9498c862438159be04342eb965bf2320db95103857e23b34179f05802b831c61423763117feb7259bd79f462639e69b3d94d6c232989cb7c51c1a76d99f784505aef17ddbc15ab54dc18d263d978696d120e4c0cb316bfe351e9ea40c2124091d741bf143e05a8e1d2079ae17d17195ad59796e58fb2c7bf6ea38ff4d6b10611918e70be6fa813b2f0c6168efe9e978f0531c79339f8d1299af0406ea410c632209378ec7cb64f59df64ec59432d5804914c2e02f880ff7808f16dedd9e5f9ff65d0b82356ded08dff783fff05451830c9fcc9ed4d29bb31fe05bb7a26b13a2eaf844599434a7e5d66f4c937151d03c84cd948d73550988a3857365879d610322df76e60af66a82fc0edfd014d427dff1b07da87c1f95cbaf328c2d5bcfad68e9f21f39908c70276fb651efc8595891987219216ff56fc537c91dd6691b946efeba5c033ea240ef807c7ed3a004eb6eafd5c5fd6f3c4130a21d598eeb02347ca27d21491e9f8596b4dc1fd68cd7c4a299aaff22aefbc785fbbab7b5cfd31c3b3569cdd37a36f551e1ef5c200ed834342946f52a7c8dc9257369cc4eb6bfd834aa522658018cb72c220f9af613e6b176b71706c083c9159edae377db5419006ab6a74911e985b90a701cd076d6d3a61aff3c7ca97226c4449c14005b7ea0e935616269a92cc93a8ed25b89c3ee4cb995ff21a6ed8d193560cb25be2af1c9d36836cde6bf208ed7a903e672b0a",
		},
		{
			keylength:     32,
			salt:          "3825bb4163f7d5cf2804ec0b31a7370f",
			evenSEK:       "53a088d93431181075f8a9bc4876359afe48967308120c93f97bbd823d8de62a",
			oddSEK:        "7893e88b6296ffcc5a2eab5f53d48efd7adaeced8cb3a851d4f8e2dbda8db17a",
			passphrase:    "foobarfoobar",
			evenEncrypted: "78614028a1a14bb323b5c618fc1bacdb7790af4ab63f271620841be94a272f2bce578a71ee17bc0be6bd0ce8ca8399ad70dbecedd034140a37ac208e6a4f662dac1bb1d9dbf621a7dd225520893376e846b7113dd6655c523f7a170b4628e555fc0f1a9da4db2fb6b43e1c27864573f5b3a4216891df7a44fc6afe4ea0b51e8c6661aadaccb5bb8646715bb186af78f25a15f4788926a7e2e3d3a02d44e6abf73c988399a81a715ccf9939584ac747871336583604e05c6ee40678656030081594fb1d5b875b10e7e9e23a4806087fa31119358e322387c4b68c3076c1f8798cb75aed16e9abe38d762a218834ef31d0224fdcbce0f3b5d1ab3de4a562051554425a9ed62467342cd97118ac51ebd968b17bd484bd91b2041040f27affb047a6a364f85784707d908846f44575aee763c00b4f252796a3893e38fb4cb0315b12f49f112b187b0a14eea5bc73c2855dd0925d79528fcacfa7fbf51d61c884dbf03dd3670ff25a7b8172f6008dcba7d5e0755cec03d537196e268c96a96eef051d8dc076af0b41b44cf016c8fa9c8a8bc8c05ba1baaf0ba5ef3707c055a3ed5dde97ae3a3fe01d80f5d86209b525b7cc427e9f5e7ccab993b8f96f5bde3831143249d0d87aff6f1098a0042d4d97ef363a4010c9a16c77aeb012656ee6dc14c955841a0d96bff509bd8ec4ba09b08d881681d25823666154791e0c6ea518fbd477bebc95e1540b45e2e5fa20fc5a55f44cb42e482f1681cf9f25b8b23e3763466e7599f2e9baac13d69df1fde7f20bda805c7b7a6b0eeb831638536a0d2b87722d93dd84bcb7c0416ec7f847d7043562f73e0cc61b8614fbfcf996d31f3c91e09515b81f8a940bd0d2c1df76d12bef9aca5ad3bbbc3184b117768c3c845e07b407fe38308c5186fbbec71b804243da8f3dc7d687eaab4614566426ac3471d0d8985d9e03dd49d69ad89d482e83c56dda25f3c993ff230343945064e98be1be142392ed8601141de324619e117fcecec781d64ccc6ad8a26c8220e118269b2e86a1965e289c3c7359a8de4496c3678211e6270ab5e351bf82516ca546c3725ff1f8be3d70a15d976405bda6af7ec6a9c2ace9651444df58bc3cf8f443af5e77e6e7d1a224129538f539c264718e835b19032d1e90d4a68a8ab86be5fb2cd8a411f23464afaa570e24b9b6a1e9293f868386f7fe62bee6505ed328a9648b5c7add16b2282d6d72b6ec744158039d135aa456b7e02d2143f7d7f6549e3a184bd6ba2205d98eacca8ce8b835473ae3049c742c6cd225f05f278a7c61923519984895671d5a852a8c0b1cea2c58b2e68076809da32cf2bbbd1efabab63b5b1744cc1d46c1bb9636e20c3ac9083fa150b8b7ccbb7c02acd945069e4b634cef0dd966895510e85bc77263dd759cc010e7b6c101ce9a7a174e86eae96bf170f326b74199fb80dd22ada008b94b65b8ff70df0c0ecb5cf6264467534e2fd284bdb5649325e8a8426ede70562ced020e7d01e8bff5ff5329c26ca316628bbb79cec4076b331c4fdc548420520e2fe659e32e15910c9b92194ab046af9016c10e4f3140f5d77cb3c3db495d1d2a2890ab5279f38c29b718adea946440c41f80ecb64c3aff989366321c4c4d036b35eff3ef0d32574d121abe1388bab5d2129ffe4742b8669f4657355c5d3be9aa4d83fa25c09d0c40ddff40d359f1d52450daceb1f2ab7555d52a1d4a95b96058e737b1e9ba46055eded6721b11cc23e68682a188e591c2f12334667433d189ebb43deb559bfaf3f0d942d815df2d61ebe7249b25b663fe8ba0d7457a49e6c351dc4ea83bbdd3b510dd4f0fcef25394773a2567369f15abbb654cb7123a",
			oddEncrypted:  "1702c977482690c91205b502100c582fafe4d64445c157d405c12eca7c754137323e3262f1df2a493bbc801c81939daab7a0e5d8daaf2f16b57529253fbc867872c79dcc5693e5755a8b003b800366141e90260dd252c483117b109c1c98c6be916388c61cc50fd89d38a14ef9528d7b637dfb1bac7ea841854f3fb1ae08a0198c810bb0e71ccc8de9400876044517d0ace69a9fb7661757f70e15691ef5316174d750cde660952999ca884a39d013ef4d6742a39ac04c8186e950c70ef8c54ba9c95413951e623d4031727a5999fc4afbac5a431cff7af97deefe1726420cdf80f6fa466e15ddec9c5032e692a74235fd3c4daa39bba62a8fee67f7a274f59dd5df1f088422a3355dfdd039edeaa292cf73c3bfa9232f89bcf9152e9c8ab6765196f4d85c3c4953aa23c2b4112eecffa956180e7b656b9ef60a39c802bc53054464bcde89d75a39adc07517c0d953602d93a6faf94b898d59a28dae483dff6059dc6aa60dbc5df49296b9588b10b2aa5951838b76ad9e6f8acbc12821f7d325a56e4cbbe3fb28d247a3ec65dedd8eeae5e57acec2629f12c0d60bfa7313f857d1d651c98bd03e53566f55833eacd80a1fad586012882a3c72ab6b69153628d3daeb94d371a9f3057d3c12e4456740cf3739401383b0ffafb1066160949c0ba01ac361acec06b7f401a761c99d6a70612266ae93431f6818acdace7f53e7012f17bb8af6ee6bbf21ede564c87946311a496bb6946de81a542dee39e4eb3b5e0b20fcd03dd9e03ac4c7e2d991abf80510acfc0f9ff1754a1cb14c71dc06744114c557767653e6f224e64e45dbcd1d56c0fb02b09122eda9baff3f9a638e0ebc586383e8d4236f9f941db18e27f3c9ac50d3247002604dfecbe8a314ec491e9a27662bcecc60af269b768e1a3cae0b1df3c395b89dd193fb7551dfa5f2e3cf6b9f4d20f9c1388d7921981184fc2bad66aeebc64066398930f3d5a9b0131f2363ddae7620dd3fa5733c0cd1a4c10b1f362b868d24c0fe1b5af164560d99c268e9379389c7b4520f91d273f3746c616ce8cb204466167d9fb7a5a9053f9e3ce3f9db75edce1c711819439e639c4b34d1e8eca923521c6124d2b8d77f9f64df8b0cace079d0212c874245c38ab5d82406322d862094f3a63ac593928ce51a71a91c5ed6443c215528779a98c49702a6b589fc2862b03bf3da2ed6edc9a06980b2d3d19401c5c6f45ef429d658c3170a5427f935e5bdbbea4bed65b3a1a39c208a6cc3d0028f67c8e8a517b151579cf33b16bd1e2c62a1a5ef5fd25c17b3d19c0caa6d1e0ad41fdbcacc54721e0fb39fc1eaa5dbd8b7ba12ca5562e606376850eaf608f0a118f9d9c3179e3042852f108393a31ebbd6b7842119f991873c1a345eb0922fd3d0580512e880557e1748d3dbaafc17e1d520e29db951d7cd92427d050fb76141a50800b717f5d8e3aeac9489ffd40fd30d5b224d17d03be3d2c23779658cedf233003a85fbe06a34241d5cfa816e67bbf1c6b121a7e92af4bb693c4830b7b0b8c4d4b912209dce55a1d48e4d276e61f75b104b1052fbba25c2f2ab2bc2542337bd1e01e7c060ba4ff7cbc06f6531737de25f2869aebb91cdcd0177b04ce12271c65445ea59bb36bce7c8fc00139f753f972afc23576215bf8cb171c78d13025d0ed11ba9ca1c3062e1d02d5f01f05628d6d76c38a4235cc3cd05eea0e543b5ca5194197cc92eb903281d16ce58fb67116f9d12b6b976b5c6b1997f08ec15de40aeaaf80f87763ace3de08c5c9aae3b8e18d5c9bf799ea7b8a2219410f23fbe5c769c048d59f23dea27647a07952af3297c074aa6645c6d5c8358f1dbdbf09847f27b",
		},
	}

	for _, test := range tests {
		c, err := New(test.keylength)
		require.NoError(t, err)

		cr := c.(*crypto)

		cr.salt = mustDecodeString(test.salt)
		cr.evenSEK = mustDecodeString(test.evenSEK)
		cr.oddSEK = mustDecodeString(test.oddSEK)

		encrypted := mustDecodeString(test.evenEncrypted)

		err = c.EncryptOrDecryptPayload(encrypted, packet.EvenKeyEncrypted, packetSequenceNumber)
		require.NoError(t, err, "keylength: %d", test.keylength)

		x := bytes.Compare(data, encrypted)
		require.Equal(t, 0, x, "keylength: %d", test.keylength)

		encrypted = mustDecodeString(test.oddEncrypted)

		err = c.EncryptOrDecryptPayload(encrypted, packet.OddKeyEncrypted, packetSequenceNumber)
		require.NoError(t, err, "keylength: %d", test.keylength)

		x = bytes.Compare(data, encrypted)
		require.Equal(t, 0, x, "keylength: %d", test.keylength)
	}
}

func TestEncode(t *testing.T) {
	packetSequenceNumber := uint32(0x79ee189e)
	originalData := "47410011000001e00000808005211651aa410000000109f000000001419a2088dc6f8003fd0320993a9d4f33dd3a34628addfddfeffd03bf8f8677d149a74e8b1e56fd8208df79622c98c2edd137b6ff2331378c2c1299a2643d4048dd5e7062159a71fceacc10e61138f1d4e8788051c2f0f000fd1213d43ad36bd05682297ec6071bbfa83864a3436e813118a71e44797f3e12c22fc9f77e03455d5bc1405a001cae497adbff3f36004d122e5b77bfba7d391a6a830ef2235cf40747010012fe062c21e108d840187080300f6197be0c04ecb77bc455f686e23867fc2f802fe36677abf008cc43760db57d600afafd3fdad2f523ce844cf7ba3668f3e2ffff085b9c20baf084b810031e0400e7a094d9f114806e003ef32a9ef7684ddbef9b0007277e9524458b687c0801ae379e360c06f9c73130e90da9fff0201e3e81f7f5f0f35ff81016b8435c12e11d52f8342465d3a0de082f87381e0ccc83c1999040918a100908fc0e0428897c0e010e225e18060083010cc847010013cab57e091b4ff7f20697809cc00441b4ec99a00906ee73f795577dd2fef700111bbb5b8002f73beeb9d8a5debd5dd821020066420423209d2222ad5f8499e05ac0a6736b460c5d3c38312fb98cbcbc3e08023f0ffeabc09c285104a5cf89d002fe006c4dfa933a12aeabe60e0016689962dabb48ffefc3011e19c3f8200c7003b226030ae2036d41f4d00c23c8ad501344ee4ad41c4cd808210406da830301a7011c007977125ea5e4b3ef78783881e52841bfaf80065a6b47010014e72d60801017eb5c10020573e36020a4947bfd7c5d66fa85882e06ac5602bb987c92f0446c1c21883013b98ea1db881fa00bd7a5326b80b68c14c43856cafff000cebf7bc7ee3f000f38a4ccbb0ab620f78630085f575d4c85907f20276aa3df30835a0fc1673d74bba557609fabafdefe1f82dac64556e4155409b77275bfbff072105b9c100478679de1c82102070400c0bddf0400c382085eb5ae0084c88020e4915a80f4cc1c0026999a12652ab37e11863f00800e264701001546021dc486dab21021d062c81e8fff8c9461b27a997aa6d9a88187003e3aacd93a67bdeb5f249fc38a002d3607c66dd9b7d7c200132200133c811f612017e2004ce2004cf030001898220002010870e0104112c300192418a81f80885b482ed470a4907da0380108225f60a44452123200491a1b9486026ed55e484c42180bbb5e1983d25dfe03b5f7ffffff6830fbf17b3d9b33030f3f9fcef0cc200c3821020028e6c59b05e62e080180b91517eef77c210af0038ed823470100161881b6af0045260ce232b57f7104d0ca0007cd483569fe4cbc44b20992fff750c300186bebfb0677fb4fdfabffc37e0074a6606339e1b6a0972434060014f49331dcc8ee301df808bfd3e77fbbdeb9eb0f0c347f043cefe0c782085604e5a6825cfc0554d4cbe411286bb3f2cc06ab7fdfbf813ec7e45fe648fb8a216b6a26bfdf0e7f3f9de17820061c10030e1085f0042d20215acad5eef733834006419dd5def7e0032daeeb5deff867e0085a6085636b57e6418fdd354701001767fbafd3063f3a74cff7fb2c188406c5d5555545d45c998b8e32e3b1de39355fe10866eef5f849efbd00fc210d600a740661595abe01244c1800622d32398baa43c11fe769cf0843540c075b7dd0170089bebe7ff80073446b69b22df7f08433a0c07a4bba03ad7ef5f88ffe05d9a71348ce7dff9d73ac290400c3820060b5e0c017f03820416cb06009597c3189040f89003613152b3b7da0033850ab2322fb41c02f7353c002a9089cc4111f683c210ce008a6c188c5d6"
	tests := []struct {
		keylength     int
		salt          string
		evenSEK       string
		oddSEK        string
		passphrase    string
		evenEncrypted string
		oddEncrypted  string
	}{
		{
			keylength:     16,
			salt:          "6c438852715a4d26e0e810b3132ca61f",
			evenSEK:       "047dc22e7f000be55a25ba56ae2e9180",
			oddSEK:        "240c8e76ccf3637641af473edaf15aaf",
			passphrase:    "foobarfoobar",
			evenEncrypted: "6053ce41905e451494d156b53c53fc6372c7e0eea81b10c4f21d4624915a93a368605cd958b14bf7274f1f75254ca606c464ea8a3f34e766fdee34dc6886512bf9b3916a3a24a3febe559cb9a07cfa5a0ec56ffee9b11605a97aa684736afcdad091fb467ca2a6ccce7356e8b2cd931f03f6c0ee4bccdd9fc6f635b080274f32f6b64a14b745efcfd7daaf05fc75369c1e37892d7f97dca5526eef2eba914cb729689b879c5a3f9a8fff061dda8f930eb20e92f241524c3743e12dd8e9b50937d73212307c2c47cb14834a5c2ab6fe1f68b46b2e1ca85cbecf00cf176a426d819ec9d41ed525f9ee7c2e7ee775cebc98357c55de1991ee23a979a4072672c4add4e215966ccef2bd04ad78f8822fcb3ed3a75d945a6c0d3dfa84b3f80c107cc94a0fb5b0b97ff71b8c8ab816ade945fb0ac70228f30e17565cd09b849a06746383193e55105904543b6079eecc53cf51d6d47690871f3f569501f0bcc790fc4c4ab344aedb53a5ef4ed19c8f5c43179f3c1de7065c32182e1ebe291809064c3cbc9ff69f92a61f3642e7ec5d711ef6a07556ea375550d6bf4221093a85b1b6bb8fd650ab3acd40f409497fdf08b88b718b408574ff9f6e003fe64803a34d5e4d86d361c5754741b60024ae24035f01f1a1e2f6dbb0aa93794fac2afa7fc824933260954b05845e856db7b7720ed0dd08504f50a4469d383b9d552e048bd674a98dfaa137bf15b740388ce0f8bcde80046d4ddc0e5987315183950477f069796201e97c990eb98370cd8f5f3bb99704d0806b59d31d02f5d0233875c27f9b132d6a7fdc3508a264b2d33a570f56aa4388e512a4181e9b0caeae8b368224d4b4ba499ef0cc7daf659d3bd38aa238577dde5d19ce1eb5257198096c00f90ab02a8c7320c6b1d10297fabf2e5f62af421a89cbaaa1e4a4e120d0e888354610c47d52a32e78f4ae21475d2201c74460d088e7686d00e7333a7c259a73944cd7981522205bec96b87157ae5fd5473e77d900f57c7edb170bc69d1e736760d63bf31c845f002259ab495c069cb39b074a066a9ddc718f3c6b94bde6fe0c82ace2d9c748f060435a1d4719651858a2f3c3ddf974035cbf1fe248fc3ee479ab37f064f019174df80f79b4bd7beb691f995358779fd124886158e669b118bbae7fac87a13558dc94f0f2dee89abe6e860082bca2fddd3822c193708b37cd53dfbeaab437a7654490bd55a29ca79e8de4f78f6c9c24aee40944d6c0e4f3cc4a5530683c087741b3ef36b464d11e207bc9b1e2ebddce33f3a706d7b7d83603b604a054d02ac0fac6d4194887b30c542e4f44a8df70d4ed53e47f019ab83dde4f251011b8fbe381ca6145547a6a9a340855d1020c078b8152a71f199653744c5c73e9910d601ee74ea6533393261b6a94ab356b3d028d05b47ce1a31cdf5acf3088d4a835285b42b7ea06cff52c98abc099fc0df611588a8941380e5ff65938aa9212ddc0d5a0b3f229615868a0654a1fc2bb0ac91f32ee802f2b316db20b758b6c7373f978f6efe7a6b8e1ffae4061e2a43f108965fdc7e60509d6896f299ad6bd3bfa5adc4a5c968878aecb8ce49ae278d613defcd5bc462c311f04093571af0f6f01484833c8977d3900e4b42085882b66d783edfeb3efd782cd6e44ef38bd1c1ce9dc427fa071caaa26924978530b3002d7e33722cbc87f1b2d670835960071763b712229740a7613da5236e10caa2466384306537d8253417a7f5bfee629129a8e52a52926a2b65f7c0c903c3762988350777a99abbeb34eeb49ebdb5a00c2bef02387651773e8f888b4f672c9f926aff5df12a899599fc40e5f95fc5394d615cafc8ef0831e21c7",
			oddEncrypted:  "8f91686a120b4461f0e078ce80805cb51994b15d31f126ece327469f326c337f25d06e67fd9209308c933379aa796d6c3c7d3dee6410e413829ddb6d6327026c00bb4000992c54f185abf9d724c772d32b8a38ac094b39c948e46c1b06e4263fbf7c7318ef3667937e17a63579decfe667c8478284fab330c500f5f9266922a35a30f612cce0af0cb8f5ec1d3b393339955f7c6470380b9cf7e5ed29b83d296895f0308805e3e5acd0c5e8f54535223e61845692a7e52763a751e0b14633ebe73fe7ec215ffb3e45bcc777661d74efdb1631e5b8fa7ab3c7ec59f2c1f2399e90479b42557b73a14288792609eab84929152216c8969e3ff7f5a17c810ce31f0387448b9a3fc5901302f295d43c98efaceb2f3a0df2f0c6ccd328da26e6c836527ac19641dbe36f4a47a05f62ea0dd411a34f5f73a86cc9daf6e58eed9d630313687e5dd06d51a92668149cb0395a3e42bdd8174396f576ae88a574f933d7ccedd152a9981f1dbd2395be9647f9b0291de1a2a1dffbf3aa8739906dd47ddeaa4978ef50762bee7df61594da9c34109da985a611aa2c9a8a3e6d388a27d1378a85496b3ba54334b559477cecbc844319768c1ded16949dd4298a0b04e6ae478953bf1a18d98fb1db3128ca2997402b8d2f5110e4c238291fbed9e01babfcaa3117f1dbd0113ae5eaff9e0c75db87d2166cd3d67db5e4bb01badd69f3d6464df2029e51438882ca99dc038bf0f0811f3773490edf9ff75f63ac97dc41656035c19c943c094cf25c1c07a5db82de71f79f94075aab064bfb8dc1a4f434e0f8d24ae775f0c773098d5bef537dcc87f8e2a91e71b87ac6320c3619c612709ee7c8077d67a5f5768503445b08aa9feef79bf6318d7abab7feb9382ab0fcfc9835ea0f94ae3c554ba9aeb447b4c05028672a89fac1370b2ce1c438d2311474f91ee97f386d06369210b52eb10237b5ac692b944ff637920fb9aa258680ebfba9df52588f2edf086925403d9bcc3e92ea513131e42654dbe42a092825d5e5b96474080e24b9fb51624427a8d2fae30dba7f249520080b80ef8709aa50285c0d5cbf144df419d85018bc458c6a3f60d44c94503461db6e53de0723da92e2e89eb2e6c98dfac7b7d9f2e542f414bae4234683ce8ca33761d5e149380e417c779c63c157c98850d37ec4711bdb1e4e2a741c4195e4c5dd18274450d340755cde38a48d3004bcd12f9af936d916fa2ee855c4a4fc590d849460e690cd76bb0bc823dae428e831cd98b8cb405d9d2427951e37886855dc3dec94a9c5d32abb4517b56950af31f89b4e29f9e6edb7bb6f24a12af7bad41f012e79da1cec774dcc241dd110f810d6c9263e1bf5711eb4af1c5ab5f0bfb15fadc442124aa55db8d83454fab030f727bd4512bb0aa6d128637025442e0f0cec76fdf97c467f2b5072bbec7c44376becce9bf425725f433e13b4a335a549b1cbf05c51cb87907749b67f46be67339870ed1b3e5e5a356e4c1b36baf9bd28a58aa9bf1f66839947f1ffbe3331263b705798d7e3d1ad86f4a39169c874fe793af83680146cbef8d604e1f707ff16f90c2b707acc246d40bedce5b1a86023ed046258d3715104795edca4e1476d59bafd9e25e8b4dd087b2748dcbb0c4965b6a310e0639f6e7d3f1a0917fc4902bca1e908526c663dd259c3ea32f3abc3959746dd171179c931586a3725bfd0e4b814388afea817501113eab47555db0a470e724020cd57b158a49ba9ed340e6d8f1a0b67666dc4c31d02485b97354816a5c3c401a676005b603af3c632a2370b46b80b28ecebe3437e8dc57cdd299c702f7aff2b536eb920f402504590236d515fec68a17d02f6c3",
		},
		{
			keylength:     24,
			salt:          "e636259ccc41e73611b9363bb58586b1",
			evenSEK:       "4dca0ad088da64fdc8e98002d141bc46fed4fa0167b931c8",
			oddSEK:        "2b2bbb64ee3942cfa31bfe58efd1d2102c40b7bc028f8946",
			passphrase:    "foobarfoobar",
			evenEncrypted: "97888ba6bd9d4015a6f62c4d3ad9c2674077218fb83ba213a634bef143186587debe76466eef8218f2b3aeae2cf00d5c46b28ffa0d325fb2ac4409fb97921e1f4c4a652d731ec6f608c92006a0ef8abe203dcf1c4f81795b126b2aff3a46d6c69a047a050a385417ecf9bea11530346607d8db1d821baf49b36b772160a1633257333b8c544fb0f4d25da28dcdb1e24a8611968053b4e636440756c7a8bf6fb7b9cb3bbc34a54e0b88c6651ac23192d18327d46e7aaf6c01a038346cb65f59d4f6be2d3a5f92c1bd043f236c0c9dbb1dad4994f426ee8ebd05a62c7558c384de5be20018ac50c73efd4ddc1255419c766f3cf53be6ad5fe35682208538e447abf0a68b1ba7d7030ace2a6b20d73bc71dd3ac7b026d60220f1eff05a66c5e06102e2801c8bd9d60091489f282eab072670f1e667bdbc2a067ddd62ffcdc82d9d78cd491277beed96f9fa87de05e6ddbbc7a79557fddd610a58c20187b71bf46f0e2017065b48689efd2862ed3bce11e0792ccfcf64494e34277be21fee874f15dd5377695b69ae1a1cf5413502ea182eea4cd4bbd8d29923aa357e05c7b974ab9df120dbaae26f282f2cc14bcb6a7075f678cf2ad3f39ce8cb01bd44bbeb60269ada55735c88d2355839267e7f9f7596e4baedd2f425d05cac9df2d3b8d6105fb9826a55c78438be8579c02664ec8bac3bde88513cb4019c2a57ad0044730d7f48568bb47ef15aae7949a8be5950421d4de73a6ad27397c0f19da4739acd7527648a8b643ecbf2451a2ac61d6e7666764f657f1003a981fa3f7ebd8db91f286b3b33a764b3542c99017f56b48db5e5f1061eb9327d137b5d6e8e86203de8501103b0fccb0cb97d0266778ea439e1432326b719a7f9396f1c773e6543b391c926920282faadf37983e2fa0dc368e2d384466981ec165df62df4386a618e5cd2a8bdd28d03a5688c38bcb00b2449c39e35f8113253c339e7e4b97048886a8046e686b97591e9b3507a9f10b9c6ca9b6b8546dbade8e42a5af607a35e99f31c06838099c6f5df0907b0f09a33216318118c03bb11374cf2a3cf2ff03978dd66b7b95dff36bfc0bf6c10e7de490e9ed6f9ec226275c59da836f405bb70f504e6b4d1140365d714b84b18a735a4f0de2b5ed7e4fc5ddaf8812d52dc6b49740b7673f408a2b64e4a05f05349706fdfea71914ae29f614fa58d7ad494f8e72de1f9ce579a0fa595a4cbe112effb769bb0e146c20f9aef08a72aff74f9a7d0404d802f9705428776bf660abce214ace695310ec32d5caa95d84e15c102032f40f7d51f409b285d272864e8e8f075e5e0d0034b0405a54c1b0946e1cbff5438fd7ed54544ab3b5e1382749b1949daac5921bee6cee31e377a331400dd6d8baefe84598d2d09159614b4a3e89de69d2b951470d8ab77534e152e04297b54418d73989e1c459c88c38c07bf881bf3e9421b1b22cf1572e969251f63613b3b696f391756ac48457748112ca41882fd6b21ec87bd0027d13a158ab224adab8e66375ba044239dec6dc6a67566a03da021fec02cf295f85eec4adec4b6583eb013b2a3505c37dfcf60de3d9376661760499b03ca718e150dc43825d93ed82c575ecc50c9d5353f66b5f37ea50b202777c7e3f8bbb7b776c98688f2505eceb735e1879d149982306c20ebb1fab4d81a1382f2d42277a7357a9a4bd4d6aaa60adb7fbbc6026710a6162f553938d4a42b2213d626f393c8b8cb7b0583e29d53b02d93f2ba47cdc1af9adc497c5b4285d73120965c44a05a1553cdaedeb615aa506e6d509aace22c2d285c0edb41c93e5041e55d65ee1b74fbcd735794fe0678c98e9d3b66ae92ef2768dee78b5",
			oddEncrypted:  "232f7ccc827d309606b5d879ae55f3a7cce1bbb6a50d5d3fc4fdfc84a57687a2783ccf101ed666809707f57d611b8754af176edc336e4535047de6d22d97c094d126805d9ce17704cbebc84596047e86aaea51e0e268fddc76c2b78e2e3dc0cde4dd11ee383ae2d24fc8151de48b05360612d9f68c864bd4fe1933af1b979ed44b721cbe84f20f1e415219a88406e4110ed20a441feb045d72bf7684a4c7b91fdde8f39e9d92aec62c292f588de5b6e29162f0e9c916cd316f6ac1bb67d781eef73ff6c34e73325900558692b5a8fad10fc25b55835fa1f0a188d1b310f7ac146567abd8ec4ff34c9e421a1d8ca7202d5b086b9c6bc91401b018e25fa27b79810b91e41e923a8e34aac3a50bc79ec4f680284fcc7334a43225a187a44eab0795b09179d9acdb5a307b043d14b30fb2956bdb39d1a1acace0c9f6f1efbab1b3b6cb5eb5b6955daaae972aaddc69a23cbe0d427598319be0b2434fed8fc680e1f185ec953afe59c83c197d2637d1646e3b9ee81b262995e4a87fc901a0200cc109add2d0f9d3fdf2c28b0e08f2e7e1b777e9e65198726f0cc32d49d6505da3d2f520c24d6ae25238ae8f546cf36c60d2f531b501828b3b2472b922e3eb35e49862a5afb9ee16fdfa0c8f550ba843e9ffd459ec95f0773143db6124e7614ee92c52420fef4afeb38f1889a00f272d744d640f8299da75691353df80d902581cb4f5e8a6a49346fb5479a8591788b1e29b251b661a1581253aaa009874e61f4a90c332828c9698ce8b8b5a844464fb024236adeeedcf2495794fba71faaa91f8495d61f5d94bd4ea77f8449317d59684d738331c513b0c673db453ed6febc4c2b22ffaea1bbc605b9b18cd97bf8b13d23dbe0c1eac7b783d27f31c280ff6a4ed0abd365ffd9094b6b70d8664242b0a56980aa43b861611c7cdc53fd655d4ce7fd85d2b694b4ccbfadc20705bf647989772d13368cd6d8564d65e2efad39b2c3892ee1056c658dc1fba3cb25e9f3754bb19f9c62c84899ae222aa72d331a58c7099c0b962b4086979492b71079808e55b3a1428e363ef07657dfeb380000ef375090d68fa96eed42a2666d9d674b5e89bb1b46ae20f2f34f1f73a9498c862438159be04342eb965bf2320db95103857e23b34179f05802b831c61423763117feb7259bd79f462639e69b3d94d6c232989cb7c51c1a76d99f784505aef17ddbc15ab54dc18d263d978696d120e4c0cb316bfe351e9ea40c2124091d741bf143e05a8e1d2079ae17d17195ad59796e58fb2c7bf6ea38ff4d6b10611918e70be6fa813b2f0c6168efe9e978f0531c79339f8d1299af0406ea410c632209378ec7cb64f59df64ec59432d5804914c2e02f880ff7808f16dedd9e5f9ff65d0b82356ded08dff783fff05451830c9fcc9ed4d29bb31fe05bb7a26b13a2eaf844599434a7e5d66f4c937151d03c84cd948d73550988a3857365879d610322df76e60af66a82fc0edfd014d427dff1b07da87c1f95cbaf328c2d5bcfad68e9f21f39908c70276fb651efc8595891987219216ff56fc537c91dd6691b946efeba5c033ea240ef807c7ed3a004eb6eafd5c5fd6f3c4130a21d598eeb02347ca27d21491e9f8596b4dc1fd68cd7c4a299aaff22aefbc785fbbab7b5cfd31c3b3569cdd37a36f551e1ef5c200ed834342946f52a7c8dc9257369cc4eb6bfd834aa522658018cb72c220f9af613e6b176b71706c083c9159edae377db5419006ab6a74911e985b90a701cd076d6d3a61aff3c7ca97226c4449c14005b7ea0e935616269a92cc93a8ed25b89c3ee4cb995ff21a6ed8d193560cb25be2af1c9d36836cde6bf208ed7a903e672b0a",
		},
		{
			keylength:     32,
			salt:          "3825bb4163f7d5cf2804ec0b31a7370f",
			evenSEK:       "53a088d93431181075f8a9bc4876359afe48967308120c93f97bbd823d8de62a",
			oddSEK:        "7893e88b6296ffcc5a2eab5f53d48efd7adaeced8cb3a851d4f8e2dbda8db17a",
			passphrase:    "foobarfoobar",
			evenEncrypted: "78614028a1a14bb323b5c618fc1bacdb7790af4ab63f271620841be94a272f2bce578a71ee17bc0be6bd0ce8ca8399ad70dbecedd034140a37ac208e6a4f662dac1bb1d9dbf621a7dd225520893376e846b7113dd6655c523f7a170b4628e555fc0f1a9da4db2fb6b43e1c27864573f5b3a4216891df7a44fc6afe4ea0b51e8c6661aadaccb5bb8646715bb186af78f25a15f4788926a7e2e3d3a02d44e6abf73c988399a81a715ccf9939584ac747871336583604e05c6ee40678656030081594fb1d5b875b10e7e9e23a4806087fa31119358e322387c4b68c3076c1f8798cb75aed16e9abe38d762a218834ef31d0224fdcbce0f3b5d1ab3de4a562051554425a9ed62467342cd97118ac51ebd968b17bd484bd91b2041040f27affb047a6a364f85784707d908846f44575aee763c00b4f252796a3893e38fb4cb0315b12f49f112b187b0a14eea5bc73c2855dd0925d79528fcacfa7fbf51d61c884dbf03dd3670ff25a7b8172f6008dcba7d5e0755cec03d537196e268c96a96eef051d8dc076af0b41b44cf016c8fa9c8a8bc8c05ba1baaf0ba5ef3707c055a3ed5dde97ae3a3fe01d80f5d86209b525b7cc427e9f5e7ccab993b8f96f5bde3831143249d0d87aff6f1098a0042d4d97ef363a4010c9a16c77aeb012656ee6dc14c955841a0d96bff509bd8ec4ba09b08d881681d25823666154791e0c6ea518fbd477bebc95e1540b45e2e5fa20fc5a55f44cb42e482f1681cf9f25b8b23e3763466e7599f2e9baac13d69df1fde7f20bda805c7b7a6b0eeb831638536a0d2b87722d93dd84bcb7c0416ec7f847d7043562f73e0cc61b8614fbfcf996d31f3c91e09515b81f8a940bd0d2c1df76d12bef9aca5ad3bbbc3184b117768c3c845e07b407fe38308c5186fbbec71b804243da8f3dc7d687eaab4614566426ac3471d0d8985d9e03dd49d69ad89d482e83c56dda25f3c993ff230343945064e98be1be142392ed8601141de324619e117fcecec781d64ccc6ad8a26c8220e118269b2e86a1965e289c3c7359a8de4496c3678211e6270ab5e351bf82516ca546c3725ff1f8be3d70a15d976405bda6af7ec6a9c2ace9651444df58bc3cf8f443af5e77e6e7d1a224129538f539c264718e835b19032d1e90d4a68a8ab86be5fb2cd8a411f23464afaa570e24b9b6a1e9293f868386f7fe62bee6505ed328a9648b5c7add16b2282d6d72b6ec744158039d135aa456b7e02d2143f7d7f6549e3a184bd6ba2205d98eacca8ce8b835473ae3049c742c6cd225f05f278a7c61923519984895671d5a852a8c0b1cea2c58b2e68076809da32cf2bbbd1efabab63b5b1744cc1d46c1bb9636e20c3ac9083fa150b8b7ccbb7c02acd945069e4b634cef0dd966895510e85bc77263dd759cc010e7b6c101ce9a7a174e86eae96bf170f326b74199fb80dd22ada008b94b65b8ff70df0c0ecb5cf6264467534e2fd284bdb5649325e8a8426ede70562ced020e7d01e8bff5ff5329c26ca316628bbb79cec4076b331c4fdc548420520e2fe659e32e15910c9b92194ab046af9016c10e4f3140f5d77cb3c3db495d1d2a2890ab5279f38c29b718adea946440c41f80ecb64c3aff989366321c4c4d036b35eff3ef0d32574d121abe1388bab5d2129ffe4742b8669f4657355c5d3be9aa4d83fa25c09d0c40ddff40d359f1d52450daceb1f2ab7555d52a1d4a95b96058e737b1e9ba46055eded6721b11cc23e68682a188e591c2f12334667433d189ebb43deb559bfaf3f0d942d815df2d61ebe7249b25b663fe8ba0d7457a49e6c351dc4ea83bbdd3b510dd4f0fcef25394773a2567369f15abbb654cb7123a",
			oddEncrypted:  "1702c977482690c91205b502100c582fafe4d64445c157d405c12eca7c754137323e3262f1df2a493bbc801c81939daab7a0e5d8daaf2f16b57529253fbc867872c79dcc5693e5755a8b003b800366141e90260dd252c483117b109c1c98c6be916388c61cc50fd89d38a14ef9528d7b637dfb1bac7ea841854f3fb1ae08a0198c810bb0e71ccc8de9400876044517d0ace69a9fb7661757f70e15691ef5316174d750cde660952999ca884a39d013ef4d6742a39ac04c8186e950c70ef8c54ba9c95413951e623d4031727a5999fc4afbac5a431cff7af97deefe1726420cdf80f6fa466e15ddec9c5032e692a74235fd3c4daa39bba62a8fee67f7a274f59dd5df1f088422a3355dfdd039edeaa292cf73c3bfa9232f89bcf9152e9c8ab6765196f4d85c3c4953aa23c2b4112eecffa956180e7b656b9ef60a39c802bc53054464bcde89d75a39adc07517c0d953602d93a6faf94b898d59a28dae483dff6059dc6aa60dbc5df49296b9588b10b2aa5951838b76ad9e6f8acbc12821f7d325a56e4cbbe3fb28d247a3ec65dedd8eeae5e57acec2629f12c0d60bfa7313f857d1d651c98bd03e53566f55833eacd80a1fad586012882a3c72ab6b69153628d3daeb94d371a9f3057d3c12e4456740cf3739401383b0ffafb1066160949c0ba01ac361acec06b7f401a761c99d6a70612266ae93431f6818acdace7f53e7012f17bb8af6ee6bbf21ede564c87946311a496bb6946de81a542dee39e4eb3b5e0b20fcd03dd9e03ac4c7e2d991abf80510acfc0f9ff1754a1cb14c71dc06744114c557767653e6f224e64e45dbcd1d56c0fb02b09122eda9baff3f9a638e0ebc586383e8d4236f9f941db18e27f3c9ac50d3247002604dfecbe8a314ec491e9a27662bcecc60af269b768e1a3cae0b1df3c395b89dd193fb7551dfa5f2e3cf6b9f4d20f9c1388d7921981184fc2bad66aeebc64066398930f3d5a9b0131f2363ddae7620dd3fa5733c0cd1a4c10b1f362b868d24c0fe1b5af164560d99c268e9379389c7b4520f91d273f3746c616ce8cb204466167d9fb7a5a9053f9e3ce3f9db75edce1c711819439e639c4b34d1e8eca923521c6124d2b8d77f9f64df8b0cace079d0212c874245c38ab5d82406322d862094f3a63ac593928ce51a71a91c5ed6443c215528779a98c49702a6b589fc2862b03bf3da2ed6edc9a06980b2d3d19401c5c6f45ef429d658c3170a5427f935e5bdbbea4bed65b3a1a39c208a6cc3d0028f67c8e8a517b151579cf33b16bd1e2c62a1a5ef5fd25c17b3d19c0caa6d1e0ad41fdbcacc54721e0fb39fc1eaa5dbd8b7ba12ca5562e606376850eaf608f0a118f9d9c3179e3042852f108393a31ebbd6b7842119f991873c1a345eb0922fd3d0580512e880557e1748d3dbaafc17e1d520e29db951d7cd92427d050fb76141a50800b717f5d8e3aeac9489ffd40fd30d5b224d17d03be3d2c23779658cedf233003a85fbe06a34241d5cfa816e67bbf1c6b121a7e92af4bb693c4830b7b0b8c4d4b912209dce55a1d48e4d276e61f75b104b1052fbba25c2f2ab2bc2542337bd1e01e7c060ba4ff7cbc06f6531737de25f2869aebb91cdcd0177b04ce12271c65445ea59bb36bce7c8fc00139f753f972afc23576215bf8cb171c78d13025d0ed11ba9ca1c3062e1d02d5f01f05628d6d76c38a4235cc3cd05eea0e543b5ca5194197cc92eb903281d16ce58fb67116f9d12b6b976b5c6b1997f08ec15de40aeaaf80f87763ace3de08c5c9aae3b8e18d5c9bf799ea7b8a2219410f23fbe5c769c048d59f23dea27647a07952af3297c074aa6645c6d5c8358f1dbdbf09847f27b",
		},
	}

	for _, test := range tests {
		c, err := New(test.keylength)
		require.NoError(t, err)

		cr := c.(*crypto)

		cr.salt = mustDecodeString(test.salt)
		cr.evenSEK = mustDecodeString(test.evenSEK)
		cr.oddSEK = mustDecodeString(test.oddSEK)

		data, _ := hex.DecodeString(originalData)

		c.EncryptOrDecryptPayload(data, packet.EvenKeyEncrypted, packetSequenceNumber)

		encrypted := mustDecodeString(test.evenEncrypted)

		x := bytes.Compare(data, encrypted)
		require.Equal(t, 0, x, "keylength: %d", test.keylength)

		data = mustDecodeString(originalData)

		c.EncryptOrDecryptPayload(data, packet.OddKeyEncrypted, packetSequenceNumber)

		encrypted = mustDecodeString(test.oddEncrypted)

		x = bytes.Compare(data, encrypted)
		require.Equal(t, 0, x, "keylength: %d", test.keylength)
	}
}
````

## File: crypto/crypto.go
````go
// Package crypto provides SRT cryptography
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/datarhei/gosrt/packet"
	"github.com/datarhei/gosrt/rand"

	"github.com/benburkert/openpgp/aes/keywrap"
)

// Crypto implements the SRT data encryption and decryption.
type Crypto interface {
	// Generate generates an even or odd SEK.
	GenerateSEK(key packet.PacketEncryption) error

	// UnmarshalMK unwraps the key with the passphrase in a Key Material Extension Message. If the passphrase
	// is wrong an error is returned.
	UnmarshalKM(km *packet.CIFKeyMaterialExtension, passphrase string) error

	// MarshalKM wraps the key with the passphrase and the odd/even SEK for a Key Material Extension Message.
	MarshalKM(km *packet.CIFKeyMaterialExtension, passphrase string, key packet.PacketEncryption) error

	// EncryptOrDecryptPayload encrypts or decrypts the data of a packet with an even or odd SEK and
	// the sequence number.
	EncryptOrDecryptPayload(data []byte, key packet.PacketEncryption, packetSequenceNumber uint32) error
}

// crypto implements the Crypto interface
type crypto struct {
	salt      []byte
	keyLength int

	evenSEK []byte
	oddSEK  []byte
}

// New returns a new SRT data encryption and decryption for the keyLength. On failure
// error is non-nil.
func New(keyLength int) (Crypto, error) {
	// 3.2.2.  Key Material
	switch keyLength {
	case 16:
	case 24:
	case 32:
	default:
		return nil, fmt.Errorf("crypto: invalid key size, must be either 16, 24, or 32")
	}

	c := &crypto{
		keyLength: keyLength,
	}

	// 3.2.2.  Key Material: "The only valid length of salt defined is 128 bits."
	c.salt = make([]byte, 16)
	if err := c.prng(c.salt); err != nil {
		return nil, fmt.Errorf("crypto: can't generate salt: %w", err)
	}

	sek, err := c.generateSEK(c.keyLength)
	if err != nil {
		return nil, err
	}
	c.evenSEK = sek

	sek, err = c.generateSEK(c.keyLength)
	if err != nil {
		return nil, err
	}
	c.oddSEK = sek

	return c, nil
}

func (c *crypto) GenerateSEK(key packet.PacketEncryption) error {
	if !key.IsValid() {
		return fmt.Errorf("crypto: unknown key type")
	}

	sek, err := c.generateSEK(c.keyLength)
	if err != nil {
		return err
	}

	switch key {
	case packet.EvenKeyEncrypted:
		c.evenSEK = sek
	case packet.OddKeyEncrypted:
		c.oddSEK = sek
	}

	return nil
}

func (c *crypto) generateSEK(keyLength int) ([]byte, error) {
	sek := make([]byte, keyLength)

	err := c.prng(sek)
	if err != nil {
		return nil, fmt.Errorf("crypto: can't generate SEK: %w", err)
	}

	return sek, nil
}

// ErrInvalidKey is returned when the packet encryption is invalid
var ErrInvalidKey = errors.New("crypto: invalid key for encryption. Must be even, odd, or both")

// ErrInvalidWrap is returned when the packet encryption indicates a different length of the wrapped key
var ErrInvalidWrap = errors.New("crypto: the un/wrapped key has the wrong length")

func (c *crypto) UnmarshalKM(km *packet.CIFKeyMaterialExtension, passphrase string) error {
	if km.KeyBasedEncryption == packet.UnencryptedPacket || !km.KeyBasedEncryption.IsValid() {
		return ErrInvalidKey
	}

	n := 1
	if km.KeyBasedEncryption == packet.EvenAndOddKey {
		n = 2
	}

	wrapLength := n * c.keyLength

	if len(km.Wrap)-8 != wrapLength {
		return ErrInvalidWrap
	}

	if len(km.Salt) != 0 {
		copy(c.salt, km.Salt)
	}

	kek, err := c.calculateKEK(passphrase, c.salt, c.keyLength)
	if err != nil {
		return err
	}

	unwrap, err := keywrap.Unwrap(kek, km.Wrap)
	if err != nil {
		return err
	}

	if len(unwrap) != wrapLength {
		return ErrInvalidWrap
	}

	switch km.KeyBasedEncryption {
	case packet.EvenKeyEncrypted:
		copy(c.evenSEK, unwrap)
	case packet.OddKeyEncrypted:
		copy(c.oddSEK, unwrap)
	default:
		copy(c.evenSEK, unwrap[:c.keyLength])
		copy(c.oddSEK, unwrap[c.keyLength:])
	}

	return nil
}

func (c *crypto) MarshalKM(km *packet.CIFKeyMaterialExtension, passphrase string, key packet.PacketEncryption) error {
	if key == packet.UnencryptedPacket || !key.IsValid() {
		return ErrInvalidKey
	}

	km.S = 0
	km.Version = 1
	km.PacketType = 2
	km.Sign = 0x2029
	km.KeyBasedEncryption = key // even or odd key
	km.KeyEncryptionKeyIndex = 0
	km.Cipher = 2
	km.Authentication = 0
	km.StreamEncapsulation = 2
	km.SLen = 16
	km.KLen = uint16(c.keyLength)

	if len(km.Salt) != 16 {
		km.Salt = make([]byte, 16)
	}
	copy(km.Salt, c.salt)

	n := 1
	if key == packet.EvenAndOddKey {
		n = 2
	}

	w := make([]byte, n*c.keyLength)

	switch key {
	case packet.EvenKeyEncrypted:
		copy(w, c.evenSEK)
	case packet.OddKeyEncrypted:
		copy(w, c.oddSEK)
	default:
		copy(w[:c.keyLength], c.evenSEK)
		copy(w[c.keyLength:], c.oddSEK)
	}

	kek, err := c.calculateKEK(passphrase, c.salt, c.keyLength)
	if err != nil {
		return err
	}

	wrap, err := keywrap.Wrap(kek, w)
	if err != nil {
		return err
	}

	if len(km.Wrap) != len(wrap) {
		km.Wrap = make([]byte, len(wrap))
	}

	copy(km.Wrap, wrap)

	return nil
}

func (c *crypto) EncryptOrDecryptPayload(data []byte, key packet.PacketEncryption, packetSequenceNumber uint32) error {
	// 6.1.2.  AES Counter
	//    0   1   2   3   4   5  6   7   8   9   10  11  12  13  14  15
	// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
	// |                   0s                  |      psn      |  0   0|
	// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
	//                            XOR
	// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+
	// |                    MSB(112, Salt)                     |
	// +---+---+---+---+---+---+---+---+---+---+---+---+---+---+
	//
	// psn    (32 bit): packet sequence number
	// ctr    (16 bit): block counter, all zeros
	// nonce (112 bit): 14 most significant bytes of the salt
	//
	// CTR = (MSB(112, Salt) XOR psn) << 16

	if len(c.salt) != 16 {
		return fmt.Errorf("crypto: invalid salt. Must be of length 16 bytes")
	}

	ctr := make([]byte, 16)

	binary.BigEndian.PutUint32(ctr[10:], packetSequenceNumber)

	for i := range ctr[:14] {
		ctr[i] ^= c.salt[i]
	}

	var sek []byte
	switch key {
	case packet.EvenKeyEncrypted:
		sek = c.evenSEK
	case packet.OddKeyEncrypted:
		sek = c.oddSEK
	default:
		return fmt.Errorf("crypto: invalid SEK selected. Must be either even or odd")
	}

	// 6.2.2.  Encrypting the Payload
	// 6.3.2.  Decrypting the Payload
	block, err := aes.NewCipher(sek)
	if err != nil {
		return err
	}

	stream := cipher.NewCTR(block, ctr)
	stream.XORKeyStream(data, data)

	return nil
}

// calculateKEK calculates a KEK based on the passphrase.
func (c *crypto) calculateKEK(passphrase string, salt []byte, keyLength int) ([]byte, error) {
	// 6.1.4.  Key Encrypting Key (KEK)
	return pbkdf2.Key(sha1.New, passphrase, salt[8:], 2048, keyLength)
}

// prng generates a random sequence of byte into the given slice p.
func (c *crypto) prng(p []byte) error {
	_, err := rand.Read(p)
	return err
}
````

## File: dial_test.go
````go
package srt

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/packet"

	"github.com/stretchr/testify/require"
)

func TestDialReject(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenWg := sync.WaitGroup{}
	listenWg.Add(1)

	go func(ln Listener) {
		listenWg.Done()
		for {
			_, _, err := ln.Accept(func(req ConnRequest) ConnType {
				return REJECT
			})

			if err == ErrListenerClosed {
				return
			}

			require.NoError(t, err)
		}
	}(ln)

	listenWg.Wait()

	conn, err := Dial("srt", "127.0.0.1:6003", DefaultConfig())
	require.Error(t, err)
	require.Nil(t, conn)

	ln.Close()
}

func TestDialOK(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenWg := sync.WaitGroup{}
	listenWg.Add(1)

	go func(ln Listener) {
		listenWg.Done()
		for {
			_, _, err := ln.Accept(func(req ConnRequest) ConnType {
				return SUBSCRIBE
			})

			if err == ErrListenerClosed {
				return
			}

			require.NoError(t, err)
		}
	}(ln)

	listenWg.Wait()

	conn, err := Dial("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)

	ln.Close()
}

func TestDialV4(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenWg := sync.WaitGroup{}
	listenWg.Add(1)

	go func(ln Listener) {
		listenWg.Done()
		for {
			_, _, err := ln.Accept(func(req ConnRequest) ConnType {
				return SUBSCRIBE
			})

			if err == ErrListenerClosed {
				return
			}

			require.NoError(t, err)
		}
	}(ln)

	listenWg.Wait()

	start := time.Now()

	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6003")
	require.NoError(t, err)

	pc, err := net.DialUDP("udp", nil, raddr)
	require.NoError(t, err)

	packets := make(chan packet.Packet, 16)

	listenWg.Add(1)

	go func() {
		buffer := make([]byte, MAX_MSS_SIZE)
		listenWg.Done()
		for {
			n, _, err := pc.ReadFrom(buffer)
			if err != nil {
				return
			}

			p, err := packet.NewPacketFromData(pc.RemoteAddr(), buffer[:n])
			require.NoError(t, err)

			packets <- p
		}
	}()

	p := packet.NewPacket(pc.RemoteAddr())

	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = 0

	sendcif := &packet.CIFHandshake{
		IsRequest:                   true,
		Version:                     4,
		EncryptionField:             0,
		ExtensionField:              2,
		InitialPacketSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     1500, // MTU size
		MaxFlowWindowSize:           25600,
		HandshakeType:               packet.HSTYPE_INDUCTION,
		SRTSocketId:                 1234,
		SynCookie:                   0,
	}

	sendcif.PeerIP.FromNetAddr(pc.LocalAddr())

	p.MarshalCIF(sendcif)

	var data bytes.Buffer

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.Write(data.Bytes())

	p = <-packets

	recvcif := &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, false, recvcif.IsRequest)
	require.Equal(t, uint32(5), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(0x4A17), recvcif.ExtensionField)
	require.Equal(t, sendcif.InitialPacketSequenceNumber, recvcif.InitialPacketSequenceNumber)
	require.Equal(t, sendcif.MaxTransmissionUnitSize, recvcif.MaxTransmissionUnitSize)
	require.Equal(t, sendcif.MaxFlowWindowSize, recvcif.MaxFlowWindowSize)
	require.Equal(t, sendcif.HandshakeType, recvcif.HandshakeType)
	require.NotEmpty(t, recvcif.SynCookie)

	sendcif.HandshakeType = packet.HSTYPE_CONCLUSION
	sendcif.SynCookie = recvcif.SynCookie

	p.MarshalCIF(sendcif)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0

	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = 0

	data.Reset()

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.Write(data.Bytes())

	p = <-packets

	recvcif = &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, false, recvcif.IsRequest)
	require.Equal(t, uint32(4), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(2), recvcif.ExtensionField)
	require.Equal(t, sendcif.InitialPacketSequenceNumber, recvcif.InitialPacketSequenceNumber)
	require.Equal(t, sendcif.MaxTransmissionUnitSize, recvcif.MaxTransmissionUnitSize)
	require.Equal(t, sendcif.MaxFlowWindowSize, recvcif.MaxFlowWindowSize)
	require.Equal(t, sendcif.HandshakeType, recvcif.HandshakeType)
	require.Empty(t, recvcif.SynCookie)

	require.False(t, recvcif.HasHS)
	require.False(t, recvcif.HasKM)
	require.False(t, recvcif.HasSID)

	pc.Close()
	ln.Close()
}

func TestDialV5(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenWg := sync.WaitGroup{}
	listenWg.Add(1)

	go func(ln Listener) {
		listenWg.Done()
		for {
			_, _, err := ln.Accept(func(req ConnRequest) ConnType {
				return SUBSCRIBE
			})

			if err == ErrListenerClosed {
				return
			}

			require.NoError(t, err)
		}
	}(ln)

	listenWg.Wait()

	start := time.Now()

	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6003")
	require.NoError(t, err)

	pc, err := net.DialUDP("udp", nil, raddr)
	require.NoError(t, err)

	packets := make(chan packet.Packet, 16)

	listenWg.Add(1)

	go func() {
		buffer := make([]byte, MAX_MSS_SIZE)
		listenWg.Done()
		for {
			n, _, err := pc.ReadFrom(buffer)
			if err != nil {
				return
			}

			p, err := packet.NewPacketFromData(pc.RemoteAddr(), buffer[:n])
			require.NoError(t, err)

			packets <- p
		}
	}()

	p := packet.NewPacket(pc.RemoteAddr())

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0

	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = 0

	sendcif := &packet.CIFHandshake{
		IsRequest:                   true,
		Version:                     4,
		EncryptionField:             0,
		ExtensionField:              2,
		InitialPacketSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     1500, // MTU size
		MaxFlowWindowSize:           25600,
		HandshakeType:               packet.HSTYPE_INDUCTION,
		SRTSocketId:                 1234,
		SynCookie:                   0,
	}

	sendcif.PeerIP.FromNetAddr(pc.LocalAddr())

	p.MarshalCIF(sendcif)

	var data bytes.Buffer

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.Write(data.Bytes())

	p = <-packets

	recvcif := &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, false, recvcif.IsRequest)
	require.Equal(t, uint32(5), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(0x4A17), recvcif.ExtensionField)
	require.Equal(t, sendcif.InitialPacketSequenceNumber, recvcif.InitialPacketSequenceNumber)
	require.Equal(t, sendcif.MaxTransmissionUnitSize, recvcif.MaxTransmissionUnitSize)
	require.Equal(t, sendcif.MaxFlowWindowSize, recvcif.MaxFlowWindowSize)
	require.Equal(t, sendcif.HandshakeType, recvcif.HandshakeType)
	require.NotEmpty(t, recvcif.SynCookie)

	sendcif.Version = 5
	sendcif.ExtensionField = recvcif.ExtensionField
	sendcif.HandshakeType = packet.HSTYPE_CONCLUSION
	sendcif.SynCookie = recvcif.SynCookie

	sendcif.HasHS = true
	sendcif.SRTHS = &packet.CIFHandshakeExtension{
		SRTVersion: SRT_VERSION,
		SRTFlags: packet.CIFHandshakeExtensionFlags{
			TSBPDSND:      true,
			TSBPDRCV:      true,
			CRYPT:         true, // must always set to true
			TLPKTDROP:     true,
			PERIODICNAK:   true,
			REXMITFLG:     true,
			STREAM:        false,
			PACKET_FILTER: false,
		},
		RecvTSBPDDelay: uint16(120),
		SendTSBPDDelay: uint16(120),
	}

	sendcif.HasSID = true
	sendcif.StreamId = "foobar"

	p.MarshalCIF(sendcif)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0

	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = 0

	data.Reset()

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.Write(data.Bytes())

	p = <-packets

	recvcif = &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, false, recvcif.IsRequest)
	require.Equal(t, uint32(5), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(5), recvcif.ExtensionField)
	require.Equal(t, sendcif.InitialPacketSequenceNumber, recvcif.InitialPacketSequenceNumber)
	require.Equal(t, sendcif.MaxTransmissionUnitSize, recvcif.MaxTransmissionUnitSize)
	require.Equal(t, sendcif.MaxFlowWindowSize, recvcif.MaxFlowWindowSize)
	require.Equal(t, sendcif.HandshakeType, recvcif.HandshakeType)
	require.Empty(t, recvcif.SynCookie)

	require.True(t, recvcif.HasHS)
	require.Equal(t, recvcif.SRTHS, sendcif.SRTHS)
	require.False(t, recvcif.HasKM)
	require.True(t, recvcif.HasSID)
	require.Equal(t, recvcif.StreamId, sendcif.StreamId)

	pc.Close()
	ln.Close()
}

// test support for servers based on libsrt <= 1.3.0
// in which DestinationSocketId of the CONCLUSION response is always zero.
func TestDialV5Pre130(t *testing.T) {
	ln, err := net.ListenPacket("udp", "127.0.0.1:6003")
	require.NoError(t, err)
	defer ln.Close()

	serverDone := make(chan error, 1)

	go func() {
		buf := make([]byte, MAX_MSS_SIZE)

		// Receive INDUCTION request.
		n, addr, err := ln.ReadFrom(buf)
		if err != nil {
			serverDone <- err
			return
		}
		p, err := packet.NewPacketFromData(addr, buf[:n])
		if err != nil {
			serverDone <- err
			return
		}
		recvcif := &packet.CIFHandshake{}
		if err = p.UnmarshalCIF(recvcif); err != nil {
			serverDone <- err
			return
		}
		callerSocketId := recvcif.SRTSocketId

		p.Header().IsControlPacket = true
		p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
		p.Header().SubType = 0
		p.Header().TypeSpecific = 0
		p.Header().Timestamp = 0
		p.Header().DestinationSocketId = callerSocketId
		inductionResp := &packet.CIFHandshake{
			IsRequest:                   false,
			Version:                     5,
			EncryptionField:             0,
			ExtensionField:              0x4A17,
			InitialPacketSequenceNumber: recvcif.InitialPacketSequenceNumber,
			MaxTransmissionUnitSize:     recvcif.MaxTransmissionUnitSize,
			MaxFlowWindowSize:           recvcif.MaxFlowWindowSize,
			HandshakeType:               packet.HSTYPE_INDUCTION,
			SRTSocketId:                 9876,
			SynCookie:                   0xdeadbeef,
		}
		inductionResp.PeerIP.FromNetAddr(ln.LocalAddr())
		p.MarshalCIF(inductionResp)
		var outbuf bytes.Buffer
		if err = p.Marshal(&outbuf); err != nil {
			serverDone <- err
			return
		}
		ln.WriteTo(outbuf.Bytes(), p.Header().Addr)

		// Receive CONCLUSION request.
		n, addr, err = ln.ReadFrom(buf)
		if err != nil {
			serverDone <- err
			return
		}
		p, err = packet.NewPacketFromData(addr, buf[:n])
		if err != nil {
			serverDone <- err
			return
		}
		recvcif = &packet.CIFHandshake{}
		if err = p.UnmarshalCIF(recvcif); err != nil {
			serverDone <- err
			return
		}

		// Send CONCLUSION response with DestinationSocketId = 0
		p.Header().IsControlPacket = true
		p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
		p.Header().SubType = 0
		p.Header().TypeSpecific = 0
		p.Header().Timestamp = 0
		p.Header().DestinationSocketId = 0
		conclusionResp := &packet.CIFHandshake{
			IsRequest:                   false,
			Version:                     5,
			EncryptionField:             0,
			ExtensionField:              1,
			InitialPacketSequenceNumber: recvcif.InitialPacketSequenceNumber,
			MaxTransmissionUnitSize:     recvcif.MaxTransmissionUnitSize,
			MaxFlowWindowSize:           recvcif.MaxFlowWindowSize,
			HandshakeType:               packet.HSTYPE_CONCLUSION,
			SRTSocketId:                 9876,
			SynCookie:                   0,
			HasHS:                       true,
			SRTHS: &packet.CIFHandshakeExtension{
				SRTVersion: SRT_VERSION,
				SRTFlags: packet.CIFHandshakeExtensionFlags{
					TSBPDSND:    true,
					TSBPDRCV:    true,
					CRYPT:       true,
					TLPKTDROP:   true,
					PERIODICNAK: true,
					REXMITFLG:   true,
				},
				RecvTSBPDDelay: uint16(DefaultConfig().ReceiverLatency.Milliseconds()),
				SendTSBPDDelay: uint16(DefaultConfig().PeerLatency.Milliseconds()),
			},
		}
		conclusionResp.PeerIP.FromNetAddr(ln.LocalAddr())
		p.MarshalCIF(conclusionResp)
		outbuf.Reset()
		if err = p.Marshal(&outbuf); err != nil {
			serverDone <- err
			return
		}
		ln.WriteTo(outbuf.Bytes(), p.Header().Addr)
		serverDone <- nil
	}()

	cfg := DefaultConfig()
	cfg.ConnectionTimeout = 3 * time.Second
	conn, err := Dial("srt", "127.0.0.1:6003", cfg)
	require.NoError(t, err)
	conn.Close()

	require.NoError(t, <-serverDone)
}

func TestDialV5MissingExtension(t *testing.T) {
	ln, err := net.ListenPacket("udp", "127.0.0.1:6003")
	require.NoError(t, err)
	defer ln.Close()

	go func() {
		// read induction request
		buf := make([]byte, MAX_MSS_SIZE)
		n, addr, err := ln.ReadFrom(buf)
		require.NoError(t, err)
		p, err := packet.NewPacketFromData(addr, buf[:n])
		require.NoError(t, err)
		recvcif := &packet.CIFHandshake{}
		err = p.UnmarshalCIF(recvcif)
		require.NoError(t, err)
		require.Equal(t, packet.HSTYPE_INDUCTION, recvcif.HandshakeType)

		// write induction response
		p.Header().IsControlPacket = true
		p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
		p.Header().SubType = 0
		p.Header().TypeSpecific = 0
		p.Header().Timestamp = 0
		p.Header().DestinationSocketId = recvcif.SRTSocketId
		sendcif := &packet.CIFHandshake{
			IsRequest:                   false,
			Version:                     5,
			EncryptionField:             0,
			ExtensionField:              0x4A17,
			InitialPacketSequenceNumber: recvcif.InitialPacketSequenceNumber,
			MaxTransmissionUnitSize:     recvcif.MaxTransmissionUnitSize,
			MaxFlowWindowSize:           recvcif.MaxFlowWindowSize,
			HandshakeType:               packet.HSTYPE_INDUCTION,
			SRTSocketId:                 recvcif.SRTSocketId,
			SynCookie:                   1234,
		}
		sendcif.PeerIP.FromNetAddr(ln.LocalAddr())
		p.MarshalCIF(sendcif)
		var outbuf bytes.Buffer
		err = p.Marshal(&outbuf)
		require.NoError(t, err)
		ln.WriteTo(outbuf.Bytes(), p.Header().Addr)

		// read conclusion request
		n, addr, err = ln.ReadFrom(buf)
		require.NoError(t, err)
		p, err = packet.NewPacketFromData(addr, buf[:n])
		require.NoError(t, err)
		recvcif = &packet.CIFHandshake{}
		err = p.UnmarshalCIF(recvcif)
		require.NoError(t, err)
		require.Equal(t, packet.HSTYPE_CONCLUSION, recvcif.HandshakeType)

		// write invalid conclusion response
		p.Header().IsControlPacket = true
		p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
		p.Header().SubType = 0
		p.Header().TypeSpecific = 0
		p.Header().Timestamp = 0
		p.Header().DestinationSocketId = recvcif.SRTSocketId
		sendcif = recvcif
		sendcif.IsRequest = false
		sendcif.SRTSocketId = 9876
		sendcif.SynCookie = 0
		sendcif.PeerIP.FromNetAddr(ln.LocalAddr())
		sendcif.HasHS = false
		p.MarshalCIF(sendcif)
		outbuf.Reset()
		err = p.Marshal(&outbuf)
		require.NoError(t, err)
		ln.WriteTo(outbuf.Bytes(), p.Header().Addr)
	}()

	_, err = Dial("srt", "127.0.0.1:6003", DefaultConfig())
	require.EqualError(t, err, "missing handshake extension")
}
````

## File: dial.go
````go
package srt

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/crypto"
	"github.com/datarhei/gosrt/packet"
	"github.com/datarhei/gosrt/rand"
)

// ErrClientClosed is returned when the client connection has
// been voluntarily closed.
var ErrClientClosed = errors.New("srt: client closed")

// dialer implements the Conn interface
type dialer struct {
	version uint32

	pc *net.UDPConn

	localAddr  net.Addr
	remoteAddr net.Addr

	config Config

	socketId                    uint32
	initialPacketSequenceNumber circular.Number

	crypto crypto.Crypto

	conn     *srtConn
	connLock sync.RWMutex
	connChan chan connResponse

	start time.Time

	rcvQueue chan packet.Packet // for packets that come from the wire

	sndMutex sync.Mutex
	sndData  bytes.Buffer // for packets that go to the wire

	shutdown     bool
	shutdownLock sync.RWMutex
	shutdownOnce sync.Once

	stopReader context.CancelFunc

	doneChan chan error
}

type connResponse struct {
	conn *srtConn
	err  error
}

// Dial connects to the address using the SRT protocol with the given config
// and returns a Conn interface.
//
// The address is of the form "host:port".
//
// Example:
//
//	Dial("srt", "127.0.0.1:3000", DefaultConfig())
//
// In case of an error the returned Conn is nil and the error is non-nil.
func Dial(network, address string, config Config) (Conn, error) {
	if network != "srt" {
		return nil, fmt.Errorf("the network must be 'srt'")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if config.Logger == nil {
		config.Logger = NewLogger(nil)
	}

	dl := &dialer{
		config: config,
	}

	netdialer := net.Dialer{
		Control: DialControl(config),
	}

	conn, err := netdialer.Dial("udp", address)
	if err != nil {
		return nil, fmt.Errorf("failed dialing: %w", err)
	}

	pc, ok := conn.(*net.UDPConn)
	if !ok {
		conn.Close()
		return nil, fmt.Errorf("failed dialing: connection is not a UDP connection")
	}

	dl.pc = pc

	dl.localAddr = pc.LocalAddr()
	dl.remoteAddr = pc.RemoteAddr()

	dl.conn = nil
	dl.connChan = make(chan connResponse)

	dl.rcvQueue = make(chan packet.Packet, 2048)

	dl.doneChan = make(chan error)

	dl.start = time.Now()

	// create a new socket ID
	dl.socketId, err = rand.Uint32()
	if err != nil {
		dl.Close()
		return nil, err
	}

	seqNum, err := rand.Uint32()
	if err != nil {
		dl.Close()
		return nil, err
	}
	dl.initialPacketSequenceNumber = circular.New(seqNum&packet.MAX_SEQUENCENUMBER, packet.MAX_SEQUENCENUMBER)

	go func() {
		buffer := make([]byte, MAX_MSS_SIZE) // MTU size

		for {
			if dl.isShutdown() {
				dl.doneChan <- ErrClientClosed
				return
			}

			pc.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, _, err := pc.ReadFrom(buffer)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}

				if dl.isShutdown() {
					dl.doneChan <- ErrClientClosed
					return
				}

				dl.doneChan <- err
				return
			}

			p, err := packet.NewPacketFromData(dl.remoteAddr, buffer[:n])
			if err != nil {
				continue
			}

			// non-blocking
			select {
			case dl.rcvQueue <- p:
			default:
				dl.log("dial", func() string { return "receive queue is full" })
			}
		}
	}()

	var readerCtx context.Context
	readerCtx, dl.stopReader = context.WithCancel(context.Background())
	go dl.reader(readerCtx)

	// Send the initial handshake request
	dl.sendInduction()

	dl.log("dial", func() string { return "waiting for response" })

	timer := time.AfterFunc(dl.config.ConnectionTimeout, func() {
		dl.connChan <- connResponse{
			conn: nil,
			err:  fmt.Errorf("connection timeout. server didn't respond"),
		}
	})

	// Wait for handshake to conclude
	response := <-dl.connChan
	if response.err != nil {
		timer.Stop()
		dl.Close()
		return nil, response.err
	}

	timer.Stop()

	dl.connLock.Lock()
	dl.conn = response.conn
	dl.connLock.Unlock()

	return dl, nil
}

func (dl *dialer) checkConnection() error {
	select {
	case err := <-dl.doneChan:
		dl.Close()
		return err
	default:
	}

	return nil
}

// reader reads packets from the receive queue and pushes them into the connection
func (dl *dialer) reader(ctx context.Context) {
	defer func() {
		dl.log("dial", func() string { return "left reader loop" })
	}()

	dl.log("dial", func() string { return "reader loop started" })

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-dl.rcvQueue:
			if dl.isShutdown() {
				break
			}

			dl.log("packet:recv:dump", func() string { return p.Dump() })

			if p.Header().DestinationSocketId != dl.socketId {
				// libsrt <= 1.3.0 sends the CONCLUSION response with DestinationSocketId = 0
				if !(p.Header().IsControlPacket &&
					p.Header().ControlType == packet.CTRLTYPE_HANDSHAKE &&
					p.Header().DestinationSocketId == 0) {
					break
				}
			}

			if p.Header().IsControlPacket && p.Header().ControlType == packet.CTRLTYPE_HANDSHAKE {
				dl.handleHandshake(p)
				break
			}

			dl.connLock.RLock()
			if dl.conn == nil {
				dl.connLock.RUnlock()
				break
			}

			dl.conn.push(p)
			dl.connLock.RUnlock()
		}
	}
}

// Send a packet to the wire. This function must be synchronous in order to allow to safely call Packet.Decommission() afterward.
func (dl *dialer) send(p packet.Packet) {
	dl.sndMutex.Lock()
	defer dl.sndMutex.Unlock()

	dl.sndData.Reset()

	if err := p.Marshal(&dl.sndData); err != nil {
		p.Decommission()
		dl.log("packet:send:error", func() string { return "marshalling packet failed" })
		return
	}

	buffer := dl.sndData.Bytes()

	dl.log("packet:send:dump", func() string { return p.Dump() })

	// Write the packet's contents to the wire
	dl.pc.Write(buffer)

	if p.Header().IsControlPacket {
		// Control packets can be decommissioned because they will not be sent again (data packets might be retransferred)
		p.Decommission()
	}
}

func (dl *dialer) handleHandshake(p packet.Packet) {
	cif := &packet.CIFHandshake{}

	err := p.UnmarshalCIF(cif)

	dl.log("handshake:recv:dump", func() string { return p.Dump() })
	dl.log("handshake:recv:cif", func() string { return cif.String() })

	if err != nil {
		dl.log("handshake:recv:error", func() string { return err.Error() })
		return
	}

	// assemble the response (4.3.1.  Caller-Listener Handshake)

	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(dl.start).Microseconds())
	p.Header().DestinationSocketId = 0 // must be 0 for handshake

	if cif.HandshakeType == packet.HSTYPE_INDUCTION {
		if cif.Version < 4 || cif.Version > 5 {
			dl.connChan <- connResponse{
				conn: nil,
				err:  fmt.Errorf("peer responded with unsupported handshake version (%d)", cif.Version),
			}

			return
		}

		cif.IsRequest = true
		cif.HandshakeType = packet.HSTYPE_CONCLUSION
		cif.InitialPacketSequenceNumber = dl.initialPacketSequenceNumber
		cif.MaxTransmissionUnitSize = dl.config.MSS // MTU size
		cif.MaxFlowWindowSize = dl.config.FC
		cif.SRTSocketId = dl.socketId
		cif.PeerIP.FromNetAddr(dl.localAddr)

		// Setup crypto context
		if len(dl.config.Passphrase) != 0 {
			keylen := dl.config.PBKeylen

			// If the server advertises a specific block cipher family and key size,
			// use this one, otherwise, use the configured one
			if cif.EncryptionField != 0 {
				switch cif.EncryptionField {
				case 2:
					keylen = 16
				case 3:
					keylen = 24
				case 4:
					keylen = 32
				}
			}

			cr, err := crypto.New(keylen)
			if err != nil {
				dl.connChan <- connResponse{
					conn: nil,
					err:  fmt.Errorf("failed creating crypto context: %w", err),
				}
			}

			dl.crypto = cr
		}

		// Verify version
		if cif.Version == 5 {
			dl.version = 5

			// Verify magic number
			if cif.ExtensionField != 0x4A17 {
				dl.connChan <- connResponse{
					conn: nil,
					err:  fmt.Errorf("peer sent the wrong magic number"),
				}

				return
			}

			cif.HasHS = true
			cif.SRTHS = &packet.CIFHandshakeExtension{
				SRTVersion: SRT_VERSION,
				SRTFlags: packet.CIFHandshakeExtensionFlags{
					TSBPDSND:      true,
					TSBPDRCV:      true,
					CRYPT:         true, // must always set to true
					TLPKTDROP:     true,
					PERIODICNAK:   true,
					REXMITFLG:     true,
					STREAM:        false,
					PACKET_FILTER: false,
				},
				RecvTSBPDDelay: uint16(dl.config.ReceiverLatency.Milliseconds()),
				SendTSBPDDelay: uint16(dl.config.PeerLatency.Milliseconds()),
			}

			cif.HasSID = true
			cif.StreamId = dl.config.StreamId

			if dl.crypto != nil {
				cif.HasKM = true
				cif.SRTKM = &packet.CIFKeyMaterialExtension{}

				if err := dl.crypto.MarshalKM(cif.SRTKM, dl.config.Passphrase, packet.EvenKeyEncrypted); err != nil {
					dl.connChan <- connResponse{
						conn: nil,
						err:  err,
					}

					return
				}
			}
		} else {
			dl.version = 4

			cif.EncryptionField = 0
			cif.ExtensionField = 2

			cif.HasHS = false
			cif.HasKM = false
			cif.HasSID = false
		}

		p.MarshalCIF(cif)

		dl.log("handshake:send:dump", func() string { return p.Dump() })
		dl.log("handshake:send:cif", func() string { return cif.String() })

		dl.send(p)
	} else if cif.HandshakeType == packet.HSTYPE_CONCLUSION {
		if cif.Version < 4 || cif.Version > 5 {
			dl.connChan <- connResponse{
				conn: nil,
				err:  fmt.Errorf("peer responded with unsupported handshake version (%d)", cif.Version),
			}

			return
		}

		recvTsbpdDelay := uint16(dl.config.ReceiverLatency.Milliseconds())
		sendTsbpdDelay := uint16(dl.config.PeerLatency.Milliseconds())

		if cif.Version == 5 {
			if cif.SRTHS == nil {
				dl.connChan <- connResponse{
					conn: nil,
					err:  fmt.Errorf("missing handshake extension"),
				}
				return
			}

			// Check if the peer version is sufficient
			if cif.SRTHS.SRTVersion < dl.config.MinVersion {
				dl.sendShutdown(cif.SRTSocketId)

				dl.connChan <- connResponse{
					conn: nil,
					err:  fmt.Errorf("peer SRT version is not sufficient"),
				}

				return
			}

			// Check the required SRT flags
			if !cif.SRTHS.SRTFlags.TSBPDSND || !cif.SRTHS.SRTFlags.TSBPDRCV || !cif.SRTHS.SRTFlags.TLPKTDROP || !cif.SRTHS.SRTFlags.PERIODICNAK || !cif.SRTHS.SRTFlags.REXMITFLG {
				dl.sendShutdown(cif.SRTSocketId)

				dl.connChan <- connResponse{
					conn: nil,
					err:  fmt.Errorf("peer doesn't agree on SRT flags"),
				}

				return
			}

			// We only support live streaming
			if cif.SRTHS.SRTFlags.STREAM {
				dl.sendShutdown(cif.SRTSocketId)

				dl.connChan <- connResponse{
					conn: nil,
					err:  fmt.Errorf("peer doesn't support live streaming"),
				}

				return
			}

			// Select the largest TSBPD delay advertised by the listener, but at least 120ms
			if cif.SRTHS.SendTSBPDDelay > recvTsbpdDelay {
				recvTsbpdDelay = cif.SRTHS.SendTSBPDDelay
			}

			if cif.SRTHS.RecvTSBPDDelay > sendTsbpdDelay {
				sendTsbpdDelay = cif.SRTHS.RecvTSBPDDelay
			}
		}

		// If the peer has a smaller MTU size, adjust to it
		if cif.MaxTransmissionUnitSize < dl.config.MSS {
			dl.config.MSS = cif.MaxTransmissionUnitSize
			dl.config.PayloadSize = dl.config.MSS - SRT_HEADER_SIZE - UDP_HEADER_SIZE

			if dl.config.PayloadSize < MIN_PAYLOAD_SIZE {
				dl.sendShutdown(cif.SRTSocketId)

				dl.connChan <- connResponse{
					conn: nil,
					err:  fmt.Errorf("effective MSS too small (%d bytes) to fit the minimal payload size (%d bytes)", dl.config.MSS, MIN_PAYLOAD_SIZE),
				}

				return
			}
		}

		// Create a new connection
		conn := newSRTConn(srtConnConfig{
			version:                     cif.Version,
			isCaller:                    true,
			localAddr:                   dl.localAddr,
			remoteAddr:                  dl.remoteAddr,
			config:                      dl.config,
			start:                       dl.start,
			socketId:                    dl.socketId,
			peerSocketId:                cif.SRTSocketId,
			tsbpdTimeBase:               uint64(time.Since(dl.start).Microseconds()),
			tsbpdDelay:                  uint64(recvTsbpdDelay) * 1000,
			peerTsbpdDelay:              uint64(sendTsbpdDelay) * 1000,
			initialPacketSequenceNumber: cif.InitialPacketSequenceNumber,
			crypto:                      dl.crypto,
			keyBaseEncryption:           packet.EvenKeyEncrypted,
			onSend:                      dl.send,
			onShutdown:                  func(*srtConn) { dl.Close() },
			logger:                      dl.config.Logger,
		})

		dl.log("connection:new", func() string { return fmt.Sprintf("%#08x (%s)", conn.SocketId(), conn.StreamId()) })

		dl.connChan <- connResponse{
			conn: conn,
			err:  nil,
		}
	} else {
		var err error

		if cif.HandshakeType.IsRejection() {
			err = fmt.Errorf("connection rejected: %s", cif.HandshakeType.String())
		} else {
			err = fmt.Errorf("unsupported handshake: %s", cif.HandshakeType.String())
		}

		dl.connChan <- connResponse{
			conn: nil,
			err:  err,
		}
	}
}

func (dl *dialer) sendInduction() {
	p := packet.NewPacket(dl.remoteAddr)

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0

	p.Header().Timestamp = uint32(time.Since(dl.start).Microseconds())
	p.Header().DestinationSocketId = 0

	cif := &packet.CIFHandshake{
		IsRequest:                   true,
		Version:                     4,
		EncryptionField:             0,
		ExtensionField:              2,
		InitialPacketSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     dl.config.MSS, // MTU size
		MaxFlowWindowSize:           dl.config.FC,
		HandshakeType:               packet.HSTYPE_INDUCTION,
		SRTSocketId:                 dl.socketId,
		SynCookie:                   0,
	}

	cif.PeerIP.FromNetAddr(dl.localAddr)

	p.MarshalCIF(cif)

	dl.log("handshake:send:dump", func() string { return p.Dump() })
	dl.log("handshake:send:cif", func() string { return cif.String() })

	dl.send(p)
}

func (dl *dialer) sendShutdown(peerSocketId uint32) {
	p := packet.NewPacket(dl.remoteAddr)

	data := [4]byte{}
	binary.BigEndian.PutUint32(data[0:], 0)

	p.SetData(data[0:4])

	p.Header().IsControlPacket = true

	p.Header().ControlType = packet.CTRLTYPE_SHUTDOWN
	p.Header().TypeSpecific = 0

	p.Header().Timestamp = uint32(time.Since(dl.start).Microseconds())
	p.Header().DestinationSocketId = peerSocketId

	dl.log("control:send:shutdown:dump", func() string { return p.Dump() })

	dl.send(p)
}

func (dl *dialer) LocalAddr() net.Addr {
	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return nil
	}

	return dl.conn.LocalAddr()
}

func (dl *dialer) RemoteAddr() net.Addr {
	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return nil
	}

	return dl.conn.RemoteAddr()
}

func (dl *dialer) SocketId() uint32 {
	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return 0
	}

	return dl.conn.SocketId()
}

func (dl *dialer) PeerSocketId() uint32 {
	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return 0
	}

	return dl.conn.PeerSocketId()
}

func (dl *dialer) StreamId() string {
	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return ""
	}

	return dl.conn.StreamId()
}

func (dl *dialer) Version() uint32 {
	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return 0
	}

	return dl.conn.Version()
}

func (dl *dialer) isShutdown() bool {
	dl.shutdownLock.RLock()
	defer dl.shutdownLock.RUnlock()

	return dl.shutdown
}

func (dl *dialer) Close() error {
	dl.shutdownOnce.Do(func() {
		dl.shutdownLock.Lock()
		dl.shutdown = true
		dl.shutdownLock.Unlock()

		dl.connLock.RLock()
		if dl.conn != nil {
			dl.conn.Close()
		}
		dl.connLock.RUnlock()

		dl.stopReader()

		dl.log("dial", func() string { return "closing socket" })
		dl.pc.Close()

		select {
		case <-dl.doneChan:
		default:
		}
	})

	return nil
}

func (dl *dialer) Read(p []byte) (n int, err error) {
	if err := dl.checkConnection(); err != nil {
		return 0, err
	}

	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return 0, fmt.Errorf("no connection")
	}

	return dl.conn.Read(p)
}

func (dl *dialer) ReadPacket() (packet.Packet, error) {
	if err := dl.checkConnection(); err != nil {
		return nil, err
	}

	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return nil, fmt.Errorf("no connection")
	}

	return dl.conn.ReadPacket()
}

func (dl *dialer) Write(p []byte) (n int, err error) {
	if err := dl.checkConnection(); err != nil {
		return 0, err
	}

	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return 0, fmt.Errorf("no connection")
	}

	return dl.conn.Write(p)
}

func (dl *dialer) WritePacket(p packet.Packet) error {
	if err := dl.checkConnection(); err != nil {
		return err
	}

	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return fmt.Errorf("no connection")
	}

	return dl.conn.WritePacket(p)
}

func (dl *dialer) SetDeadline(t time.Time) error      { return dl.conn.SetDeadline(t) }
func (dl *dialer) SetReadDeadline(t time.Time) error  { return dl.conn.SetReadDeadline(t) }
func (dl *dialer) SetWriteDeadline(t time.Time) error { return dl.conn.SetWriteDeadline(t) }

func (dl *dialer) Stats(s *Statistics) {
	dl.connLock.RLock()
	defer dl.connLock.RUnlock()

	if dl.conn == nil {
		return
	}

	dl.conn.Stats(s)
}

func (dl *dialer) log(topic string, message func() string) {
	if dl.config.Logger == nil {
		return
	}

	dl.config.Logger.Print(topic, dl.socketId, 2, message)
}
````

## File: doc.go
````go
/*
Package srt provides an interface for network I/O using the SRT protocol (https://github.com/Haivision/srt).

The package gives access to the basic interface provided by the Dial, Listen, and Accept functions and the associated
Conn and Listener interfaces.

The Dial function connects to a server:

	conn, err := srt.Dial("srt", "golang.org:6000", srt.Config{
		StreamId: "...",
	})
	if err != nil {
		// handle error
	}

	buffer := make([]byte, 2048)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			// handle error
		}

		// handle received data
	}

	conn.Close()

The Listen function creates servers:

	ln, err := srt.Listen("srt", ":6000", srt.Config{...})
	if err != nil {
		// handle error
	}

	for {
		conn, mode, err := ln.Accept(handleConnect)
		if err != nil {
			// handle error
		}

		if mode == srt.REJECT {
			// rejected connection, ignore
			continue
		}

		if mode == srt.PUBLISH {
			go handlePublish(conn)
		} else {
			go handleSubscribe(conn)
		}
	}

The ln.Accept function expects a function that takes a srt.ConnRequest
and returns a srt.ConnType. The srt.ConnRequest lets you retrieve the
streamid with on which you can decide what mode (srt.ConnType) to return.

Check out the Server type that wraps the Listen and Accept into a
convenient framework for your own SRT server.
*/
package srt
````

## File: listen_test.go
````go
package srt

import (
	"bytes"
	"context"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/packet"

	"github.com/stretchr/testify/require"
)

func TestListenReuse(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	ln.Close()

	ln, err = Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	ln.Close()
}

func TestListen(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenWg := sync.WaitGroup{}
	listenWg.Add(1)

	go func(ln Listener) {
		listenWg.Done()
		for {
			_, _, err := ln.Accept(func(req ConnRequest) ConnType {
				require.Equal(t, "foobar", req.StreamId())
				require.False(t, req.IsEncrypted())

				return SUBSCRIBE
			})

			if err == ErrListenerClosed {
				return
			}

			require.NoError(t, err)
		}
	}(ln)

	listenWg.Wait()

	config := DefaultConfig()
	config.StreamId = "foobar"

	conn, err := Dial("srt", "127.0.0.1:6003", config)
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)

	ln.Close()
}

func TestListenCrypt(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenWg := sync.WaitGroup{}
	listenWg.Add(1)

	go func(ln Listener) {
		listenWg.Done()
		for {
			_, _, err := ln.Accept(func(req ConnRequest) ConnType {
				require.Equal(t, "foobar", req.StreamId())
				require.True(t, req.IsEncrypted())

				if req.SetPassphrase("zaboofzaboof") != nil {
					return REJECT
				}

				return SUBSCRIBE
			})

			if err == ErrListenerClosed {
				return
			}

			require.NoError(t, err)
		}
	}(ln)

	listenWg.Wait()

	config := DefaultConfig()
	config.StreamId = "foobar"
	config.Passphrase = "zaboofzaboof"

	conn, err := Dial("srt", "127.0.0.1:6003", config)
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)

	config.Passphrase = "raboofraboof"

	_, err = Dial("srt", "127.0.0.1:6003", config)
	require.Error(t, err)

	ln.Close()
}

func TestListenHSV4(t *testing.T) {
	start := time.Now()

	lc := net.ListenConfig{
		Control: ListenControl(DefaultConfig()),
	}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:6003")
	require.NoError(t, err)

	pc := lp.(*net.UDPConn)

	listenWg := sync.WaitGroup{}

	packets := make(chan packet.Packet, 16)

	listenWg.Add(1)

	go func() {
		buffer := make([]byte, MAX_MSS_SIZE)
		listenWg.Done()
		for {
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				return
			}

			p, err := packet.NewPacketFromData(addr, buffer[:n])
			require.NoError(t, err)

			if p.Header().ControlType != packet.CTRLTYPE_HANDSHAKE {
				continue
			}

			packets <- p
		}
	}()

	listenWg.Wait()

	go func() {
		conn, err := Dial("srt", "127.0.0.1:6003", DefaultConfig())
		if err != nil {
			if err == ErrClientClosed {
				return
			}
			require.NoError(t, err)
		}
		require.NotNil(t, conn)

		conn.Close()
	}()

	p := <-packets

	recvcif := &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, uint32(4), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(2), recvcif.ExtensionField)
	require.Equal(t, packet.HSTYPE_INDUCTION, recvcif.HandshakeType)
	require.Empty(t, recvcif.SynCookie)

	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = recvcif.SRTSocketId

	sendcif := &packet.CIFHandshake{
		IsRequest:                   false,
		Version:                     4,
		EncryptionField:             0,
		ExtensionField:              2,
		InitialPacketSequenceNumber: recvcif.InitialPacketSequenceNumber,
		MaxTransmissionUnitSize:     recvcif.MaxTransmissionUnitSize,
		MaxFlowWindowSize:           recvcif.MaxFlowWindowSize,
		HandshakeType:               packet.HSTYPE_INDUCTION,
		SRTSocketId:                 recvcif.SRTSocketId,
		SynCookie:                   1234,
	}

	sendcif.PeerIP.FromNetAddr(pc.LocalAddr())

	p.MarshalCIF(sendcif)

	var data bytes.Buffer

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.WriteTo(data.Bytes(), p.Header().Addr)

	p = <-packets

	recvcif = &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, uint32(4), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(2), recvcif.ExtensionField)
	require.Equal(t, packet.HSTYPE_CONCLUSION, recvcif.HandshakeType)
	require.Equal(t, sendcif.SynCookie, recvcif.SynCookie)

	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = recvcif.SRTSocketId

	sendcif = recvcif
	sendcif.IsRequest = false
	sendcif.SRTSocketId = 9876
	sendcif.SynCookie = 0

	sendcif.PeerIP.FromNetAddr(pc.LocalAddr())

	p.MarshalCIF(sendcif)

	data.Reset()

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.WriteTo(data.Bytes(), p.Header().Addr)

	pc.Close()
}

func TestListenHSV5(t *testing.T) {
	start := time.Now()

	lc := net.ListenConfig{
		Control: ListenControl(DefaultConfig()),
	}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:6003")
	require.NoError(t, err)

	pc := lp.(*net.UDPConn)

	listenWg := sync.WaitGroup{}

	packets := make(chan packet.Packet, 16)

	listenWg.Add(1)

	go func() {
		buffer := make([]byte, MAX_MSS_SIZE)
		listenWg.Done()
		for {
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				return
			}

			p, err := packet.NewPacketFromData(addr, buffer[:n])
			require.NoError(t, err)

			if p.Header().ControlType != packet.CTRLTYPE_HANDSHAKE {
				continue
			}

			packets <- p
		}
	}()

	listenWg.Wait()

	go func() {
		config := DefaultConfig()
		config.StreamId = "foobar"
		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if err != nil {
			if err == ErrClientClosed {
				return
			}
			require.NoError(t, err)
		}
		require.NotNil(t, conn)

		conn.Close()
	}()

	p := <-packets

	recvcif := &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, uint32(4), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(2), recvcif.ExtensionField)
	require.Equal(t, packet.HSTYPE_INDUCTION, recvcif.HandshakeType)
	require.Empty(t, recvcif.SynCookie)

	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = recvcif.SRTSocketId

	sendcif := &packet.CIFHandshake{
		IsRequest:                   false,
		Version:                     5,
		EncryptionField:             0,
		ExtensionField:              0x4A17,
		InitialPacketSequenceNumber: recvcif.InitialPacketSequenceNumber,
		MaxTransmissionUnitSize:     recvcif.MaxTransmissionUnitSize,
		MaxFlowWindowSize:           recvcif.MaxFlowWindowSize,
		HandshakeType:               packet.HSTYPE_INDUCTION,
		SRTSocketId:                 recvcif.SRTSocketId,
		SynCookie:                   1234,
	}

	sendcif.PeerIP.FromNetAddr(pc.LocalAddr())

	p.MarshalCIF(sendcif)

	var data bytes.Buffer

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.WriteTo(data.Bytes(), p.Header().Addr)

	p = <-packets

	recvcif = &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	require.Equal(t, uint32(5), recvcif.Version)
	require.Equal(t, uint16(0), recvcif.EncryptionField)
	require.Equal(t, uint16(5), recvcif.ExtensionField)
	require.Equal(t, packet.HSTYPE_CONCLUSION, recvcif.HandshakeType)
	require.Equal(t, sendcif.SynCookie, recvcif.SynCookie)

	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = uint32(time.Since(start).Microseconds())
	p.Header().DestinationSocketId = recvcif.SRTSocketId

	sendcif = recvcif
	sendcif.IsRequest = false
	sendcif.SRTSocketId = 9876
	sendcif.SynCookie = 0

	sendcif.PeerIP.FromNetAddr(pc.LocalAddr())

	p.MarshalCIF(sendcif)

	data.Reset()

	err = p.Marshal(&data)
	require.NoError(t, err)

	pc.WriteTo(data.Bytes(), p.Header().Addr)

	pc.Close()
}

func TestListenAsync(t *testing.T) {
	const parallelCount = 2
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)
	var (
		// All streams are pending
		pendingWg  sync.WaitGroup
		pendingSet sync.Map // Set of which streams are pending
		// All streams are connected
		connectedWg sync.WaitGroup
		// All listener goroutines are stopped
		listenerWg sync.WaitGroup
	)
	listenerWg.Add(parallelCount)
	pendingWg.Add(parallelCount)
	connectedWg.Add(parallelCount)
	for i := range parallelCount {
		go func() {
			defer listenerWg.Done()
			for {
				_, _, err := ln.Accept(func(req ConnRequest) ConnType {
					// Only call Done() if we're the first request for this stream
					if _, ok := pendingSet.Swap(req.StreamId(), struct{}{}); !ok {
						pendingWg.Done()
					}
					// Wait for all streams to be pending Before returning
					pendingWg.Wait()
					return PUBLISH
				})
				if err == ErrListenerClosed {
					return
				}
				require.NoError(t, err)
			}
		}()

		go func(streamId string) {
			config := DefaultConfig()
			config.StreamId = streamId
			conn, err := Dial("srt", "127.0.0.1:6003", config)
			require.NoError(t, err)
			connectedWg.Done()
			conn.Close()
		}(strconv.Itoa(i))
	}

	// Wait for all streams to be connected
	connectedWg.Wait()
	ln.Close()
	listenerWg.Wait()
}

func TestListenHSV5MissingExtension(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenDone := make(chan struct{})
	defer func() { <-listenDone }()

	go func() {
		defer close(listenDone)
		for {
			_, _, err := ln.Accept(func(req ConnRequest) ConnType {
				return SUBSCRIBE
			})
			if err != nil {
				break
			}
		}
	}()

	conn, err := net.Dial("udp", "127.0.0.1:6003")
	require.NoError(t, err)
	defer conn.Close()

	// send induction request
	p := packet.NewPacket(conn.RemoteAddr())
	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = 0
	p.Header().DestinationSocketId = 0
	sendcif := &packet.CIFHandshake{
		IsRequest:                   true,
		Version:                     4,
		EncryptionField:             0,
		ExtensionField:              2,
		InitialPacketSequenceNumber: circular.New(10000, packet.MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     MAX_MSS_SIZE,
		MaxFlowWindowSize:           25600,
		HandshakeType:               packet.HSTYPE_INDUCTION,
		SRTSocketId:                 55555,
		SynCookie:                   0,
	}
	sendcif.PeerIP.FromNetAddr(conn.LocalAddr())
	p.MarshalCIF(sendcif)
	var buf bytes.Buffer
	err = p.Marshal(&buf)
	require.NoError(t, err)
	_, err = conn.Write(buf.Bytes())
	require.NoError(t, err)

	// read induction response
	inbuf := make([]byte, MAX_MSS_SIZE)
	n, err := conn.Read(inbuf)
	require.NoError(t, err)
	p, err = packet.NewPacketFromData(conn.RemoteAddr(), inbuf[:n])
	require.NoError(t, err)
	recvcif := &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	// send conclusion
	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = 0
	p.Header().DestinationSocketId = 0 // recvcif.SRTSocketId
	sendcif.Version = 5
	sendcif.ExtensionField = recvcif.ExtensionField
	sendcif.HandshakeType = packet.HSTYPE_CONCLUSION
	sendcif.SynCookie = recvcif.SynCookie
	sendcif.HasSID = true
	sendcif.StreamId = "foobar"
	p.MarshalCIF(sendcif)
	buf.Reset()
	err = p.Marshal(&buf)
	require.NoError(t, err)
	_, err = conn.Write(buf.Bytes())
	require.NoError(t, err)

	// read error
	n, err = conn.Read(inbuf)
	require.NoError(t, err)
	p, err = packet.NewPacketFromData(conn.RemoteAddr(), inbuf[:n])
	require.NoError(t, err)
	recvcif = &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)
	require.Equal(t, recvcif.HandshakeType, packet.HandshakeType(REJ_ROGUE))

	ln.Close()
}

func TestListenParallelRequests(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	listenDone := make(chan struct{})
	defer func() { <-listenDone }()

	var reqReady sync.WaitGroup
	reqReady.Add(4)

	var serverSideConnReady sync.WaitGroup
	serverSideConnReady.Add(4)

	go func() {
		defer close(listenDone)

		for {
			req, err := ln.Accept2()
			if err != nil {
				break
			}

			reqReady.Done()

			go func() {
				defer serverSideConnReady.Done()

				// wait for all requests to be pending
				reqReady.Wait()

				conn, err := req.Accept()
				require.NoError(t, err)
				conn.Close()
			}()
		}
	}()

	var clientSideConnReady sync.WaitGroup

	for range 4 {
		clientSideConnReady.Go(func() {
			config := DefaultConfig()
			config.StreamId = "foobar"

			conn, err := Dial("srt", "127.0.0.1:6003", config)
			require.NoError(t, err)

			err = conn.Close()
			require.NoError(t, err)
		})
	}

	serverSideConnReady.Wait()
	clientSideConnReady.Wait()

	ln.Close()
}

func TestListenDiscardRepeatedHandshakes(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	singleReqReceived := make(chan struct{})

	listenDone := make(chan struct{})
	defer func() { <-listenDone }()

	defer ln.Close()

	go func() {
		defer close(listenDone)

		for {
			req, err := ln.Accept2()
			if err != nil {
				break
			}

			close(singleReqReceived)
			defer req.Reject(REJ_CLOSE)
		}
	}()

	for range 4 {
		conn, err := net.Dial("udp", "127.0.0.1:6003")
		require.NoError(t, err)
		defer conn.Close()

		// send induction request
		p := packet.NewPacket(conn.RemoteAddr())
		p.Header().IsControlPacket = true
		p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
		p.Header().SubType = 0
		p.Header().TypeSpecific = 0
		p.Header().Timestamp = 0
		p.Header().DestinationSocketId = 0
		sendcif := &packet.CIFHandshake{
			IsRequest:                   true,
			Version:                     4,
			EncryptionField:             0,
			ExtensionField:              2,
			InitialPacketSequenceNumber: circular.New(10000, packet.MAX_SEQUENCENUMBER),
			MaxTransmissionUnitSize:     MAX_MSS_SIZE,
			MaxFlowWindowSize:           25600,
			HandshakeType:               packet.HSTYPE_INDUCTION,
			SRTSocketId:                 55555,
			SynCookie:                   0,
		}
		sendcif.PeerIP.FromNetAddr(conn.LocalAddr())
		p.MarshalCIF(sendcif)
		var buf bytes.Buffer
		err = p.Marshal(&buf)
		require.NoError(t, err)
		_, err = conn.Write(buf.Bytes())
		require.NoError(t, err)

		// read induction response
		inbuf := make([]byte, 1024)
		n, err := conn.Read(inbuf)
		require.NoError(t, err)
		p, err = packet.NewPacketFromData(conn.RemoteAddr(), inbuf[:n])
		require.NoError(t, err)
		recvcif := &packet.CIFHandshake{}
		err = p.UnmarshalCIF(recvcif)
		require.NoError(t, err)

		// send conclusion
		p.Header().IsControlPacket = true
		p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
		p.Header().SubType = 0
		p.Header().TypeSpecific = 0
		p.Header().Timestamp = 0
		p.Header().DestinationSocketId = 0 // recvcif.SRTSocketId
		sendcif.Version = 5
		sendcif.ExtensionField = recvcif.ExtensionField
		sendcif.HandshakeType = packet.HSTYPE_CONCLUSION
		sendcif.SynCookie = recvcif.SynCookie
		sendcif.HasHS = true
		sendcif.SRTHS = &packet.CIFHandshakeExtension{
			SRTVersion: SRT_VERSION,
			SRTFlags: packet.CIFHandshakeExtensionFlags{
				TSBPDSND:      true,
				TSBPDRCV:      true,
				CRYPT:         true, // must always set to true
				TLPKTDROP:     true,
				PERIODICNAK:   true,
				REXMITFLG:     true,
				STREAM:        false,
				PACKET_FILTER: false,
			},
			RecvTSBPDDelay: uint16(120),
			SendTSBPDDelay: uint16(120),
		}
		sendcif.HasSID = true
		sendcif.StreamId = "foobar"
		p.MarshalCIF(sendcif)
		buf.Reset()
		err = p.Marshal(&buf)
		require.NoError(t, err)
		_, err = conn.Write(buf.Bytes())
		require.NoError(t, err)
	}

	<-singleReqReceived
}

func TestListenMultipleIPs(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)
	defer ln.Close()

	serverDone := make(chan struct{})

	go func() {
		defer close(serverDone)

		req, err := ln.Accept2()
		require.NoError(t, err)

		conn, err := req.Accept()
		require.NoError(t, err)
		defer conn.Close()

		localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
		require.True(t, ok)
		require.Equal(t, "127.0.0.1", localAddr.IP.String())
	}()

	clientConn, err := Dial("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)
	defer clientConn.Close()

	<-serverDone
}
func TestListenAcceptAndDiscardRepeatedHandshakes(t *testing.T) {
	ln, err := Listen("srt", "127.0.0.1:6003", DefaultConfig())
	require.NoError(t, err)

	singleReqAccepted := make(chan struct{})

	listenDone := make(chan struct{})
	defer func() { <-listenDone }()

	defer ln.Close()

	go func() {
		defer close(listenDone)

		for {
			req, err := ln.Accept2()
			if err != nil {
				break
			}

			conn, err := req.Accept()
			require.NoError(t, err)
			defer conn.Close()

			close(singleReqAccepted)
		}
	}()

	conn, err := net.Dial("udp", "127.0.0.1:6003")
	require.NoError(t, err)
	defer conn.Close()

	// write induction request
	p := packet.NewPacket(conn.RemoteAddr())
	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = 0
	p.Header().DestinationSocketId = 0
	sendcif := &packet.CIFHandshake{
		IsRequest:                   true,
		Version:                     4,
		EncryptionField:             0,
		ExtensionField:              2,
		InitialPacketSequenceNumber: circular.New(10000, packet.MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     MAX_MSS_SIZE,
		MaxFlowWindowSize:           25600,
		HandshakeType:               packet.HSTYPE_INDUCTION,
		SRTSocketId:                 55555,
		SynCookie:                   0,
	}
	sendcif.PeerIP.FromNetAddr(conn.LocalAddr())
	p.MarshalCIF(sendcif)
	var buf bytes.Buffer
	err = p.Marshal(&buf)
	require.NoError(t, err)
	_, err = conn.Write(buf.Bytes())
	require.NoError(t, err)

	// read induction response
	inbuf := make([]byte, MAX_MSS_SIZE)
	n, err := conn.Read(inbuf)
	require.NoError(t, err)
	p, err = packet.NewPacketFromData(conn.RemoteAddr(), inbuf[:n])
	require.NoError(t, err)
	recvcif := &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)

	// write conclusion request
	p.Header().IsControlPacket = true
	p.Header().ControlType = packet.CTRLTYPE_HANDSHAKE
	p.Header().SubType = 0
	p.Header().TypeSpecific = 0
	p.Header().Timestamp = 0
	p.Header().DestinationSocketId = 0
	sendcif.Version = 5
	sendcif.ExtensionField = recvcif.ExtensionField
	sendcif.HandshakeType = packet.HSTYPE_CONCLUSION
	sendcif.SRTSocketId = 234425644
	sendcif.SynCookie = recvcif.SynCookie
	sendcif.HasHS = true
	sendcif.SRTHS = &packet.CIFHandshakeExtension{
		SRTVersion: SRT_VERSION,
		SRTFlags: packet.CIFHandshakeExtensionFlags{
			TSBPDSND:      true,
			TSBPDRCV:      true,
			CRYPT:         true,
			TLPKTDROP:     true,
			PERIODICNAK:   true,
			REXMITFLG:     true,
			STREAM:        false,
			PACKET_FILTER: false,
		},
		RecvTSBPDDelay: uint16(120),
		SendTSBPDDelay: uint16(120),
	}
	sendcif.HasSID = true
	sendcif.StreamId = "foobar"
	p.MarshalCIF(sendcif)
	buf.Reset()
	err = p.Marshal(&buf)
	require.NoError(t, err)
	_, err = conn.Write(buf.Bytes())
	require.NoError(t, err)

	// read conclusion response
	n, err = conn.Read(inbuf)
	require.NoError(t, err)
	p, err = packet.NewPacketFromData(conn.RemoteAddr(), inbuf[:n])
	require.NoError(t, err)
	recvcif = &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)
	require.Equal(t, packet.HSTYPE_CONCLUSION, recvcif.HandshakeType)
	require.False(t, recvcif.IsRequest)

	<-singleReqAccepted

	// write conclusion request, again
	_, err = conn.Write(buf.Bytes())
	require.NoError(t, err)

	// read conclusion response, again (must receive the duplicated response robustly)
	n, err = conn.Read(inbuf)
	require.NoError(t, err)
	p, err = packet.NewPacketFromData(conn.RemoteAddr(), inbuf[:n])
	require.NoError(t, err)
	recvcif = &packet.CIFHandshake{}
	err = p.UnmarshalCIF(recvcif)
	require.NoError(t, err)
	require.Equal(t, packet.HSTYPE_CONCLUSION, recvcif.HandshakeType)
	require.False(t, recvcif.IsRequest)

	// wait some time to make sure that close(singleReqAccepted) is not triggered
	time.Sleep(500 * time.Millisecond)
}
````

## File: listen.go
````go
package srt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"sync"
	"time"

	srtnet "github.com/datarhei/gosrt/net"
	"github.com/datarhei/gosrt/packet"
)

// ConnType represents the kind of connection as returned
// from the AcceptFunc. It is one of REJECT, PUBLISH, or SUBSCRIBE.
type ConnType int

// String returns a string representation of the ConnType.
func (c ConnType) String() string {
	switch c {
	case REJECT:
		return "REJECT"
	case PUBLISH:
		return "PUBLISH"
	case SUBSCRIBE:
		return "SUBSCRIBE"
	default:
		return ""
	}
}

const (
	REJECT    ConnType = ConnType(1 << iota) // Reject a connection
	PUBLISH                                  // This connection is meant to write data to the server
	SUBSCRIBE                                // This connection is meant to read data from a PUBLISHed stream
)

// RejectionReason are the rejection reasons that can be returned from the AcceptFunc in order to send
// another reason than the default one (REJ_PEER) to the client.
type RejectionReason uint32

// Table 7: Handshake Rejection Reason Codes
const (
	REJ_UNKNOWN    RejectionReason = 1000 // unknown reason
	REJ_SYSTEM     RejectionReason = 1001 // system function error
	REJ_PEER       RejectionReason = 1002 // rejected by peer
	REJ_RESOURCE   RejectionReason = 1003 // resource allocation problem
	REJ_ROGUE      RejectionReason = 1004 // incorrect data in handshake
	REJ_BACKLOG    RejectionReason = 1005 // listener's backlog exceeded
	REJ_IPE        RejectionReason = 1006 // internal program error
	REJ_CLOSE      RejectionReason = 1007 // socket is closing
	REJ_VERSION    RejectionReason = 1008 // peer is older version than agent's min
	REJ_RDVCOOKIE  RejectionReason = 1009 // rendezvous cookie collision
	REJ_BADSECRET  RejectionReason = 1010 // wrong password
	REJ_UNSECURE   RejectionReason = 1011 // password required or unexpected
	REJ_MESSAGEAPI RejectionReason = 1012 // stream flag collision
	REJ_CONGESTION RejectionReason = 1013 // incompatible congestion-controller type
	REJ_FILTER     RejectionReason = 1014 // incompatible packet filter
	REJ_GROUP      RejectionReason = 1015 // incompatible group
)

// These are the extended rejection reasons that may be less well supported
// Codes & their meanings taken from https://github.com/Haivision/srt/blob/f477af533562505abf5295f059cf2156b17be740/srtcore/access_control.h
const (
	REJX_BAD_REQUEST   RejectionReason = 1400 // General syntax error in the SocketID specification (also a fallback code for undefined cases)
	REJX_UNAUTHORIZED  RejectionReason = 1401 // Authentication failed, provided that the user was correctly identified and access to the required resource would be granted
	REJX_OVERLOAD      RejectionReason = 1402 // The server is too heavily loaded, or you have exceeded credits for accessing the service and the resource.
	REJX_FORBIDDEN     RejectionReason = 1403 // Access denied to the resource by any kind of reason.
	REJX_NOTFOUND      RejectionReason = 1404 // Resource not found at this time.
	REJX_BAD_MODE      RejectionReason = 1405 // The mode specified in `m` key in StreamID is not supported for this request.
	REJX_UNACCEPTABLE  RejectionReason = 1406 // The requested parameters specified in SocketID cannot be satisfied for the requested resource. Also when m=publish and the data format is not acceptable.
	REJX_CONFLICT      RejectionReason = 1407 // The resource being accessed is already locked for modification. This is in case of m=publish and the specified resource is currently read-only.
	REJX_NOTSUP_MEDIA  RejectionReason = 1415 // The media type is not supported by the application. This is the `t` key that specifies the media type as stream, file and auth, possibly extended by the application.
	REJX_LOCKED        RejectionReason = 1423 // The resource being accessed is locked for any access.
	REJX_FAILED_DEPEND RejectionReason = 1424 // The request failed because it specified a dependent session ID that has been disconnected.
	REJX_ISE           RejectionReason = 1500 // Unexpected internal server error
	REJX_UNIMPLEMENTED RejectionReason = 1501 // The request was recognized, but the current version doesn't support it.
	REJX_GW            RejectionReason = 1502 // The server acts as a gateway and the target endpoint rejected the connection.
	REJX_DOWN          RejectionReason = 1503 // The service has been temporarily taken over by a stub reporting this error. The real service can be down for maintenance or crashed.
	REJX_VERSION       RejectionReason = 1505 // SRT version not supported. This might be either unsupported backward compatibility, or an upper value of a version.
	REJX_NOROOM        RejectionReason = 1507 // The data stream cannot be archived due to lacking storage space. This is in case when the request type was to send a file or the live stream to be archived.
)

// ErrListenerClosed is returned when the listener is about to shutdown.
var ErrListenerClosed = errors.New("srt: listener closed")

// AcceptFunc receives a connection request and returns the type of connection
// and is required by the Listener for each Accept of a new connection.
type AcceptFunc func(req ConnRequest) ConnType

// Listener waits for new connections
type Listener interface {
	// Accept2 waits for new connections.
	// On closing the err will be ErrListenerClosed.
	Accept2() (ConnRequest, error)

	// Accept waits for new connections. For each new connection the AcceptFunc
	// gets called. Conn is a new connection if AcceptFunc is PUBLISH or SUBSCRIBE.
	// If AcceptFunc returns REJECT, Conn is nil. In case of failure error is not
	// nil, Conn is nil and ConnType is REJECT. On closing the listener err will
	// be ErrListenerClosed and ConnType is REJECT.
	//
	// Deprecated: replaced by Accept2().
	Accept(AcceptFunc) (Conn, ConnType, error)

	// Close closes the listener. It will stop accepting new connections and
	// close all currently established connections.
	Close()

	// Addr returns the address of the listener.
	Addr() net.Addr
}

// listener implements the Listener interface.
type listener struct {
	pc   *packetConn
	addr net.Addr

	config Config

	backlog     chan packet.Packet
	conns       map[uint32]*srtConn
	connsByPeer map[uint32]*srtConn
	lock        sync.RWMutex

	start time.Time

	rcvQueue chan packet.Packet

	sndMutex sync.Mutex
	sndData  bytes.Buffer

	syncookie *srtnet.SYNCookie

	shutdown     bool
	shutdownLock sync.RWMutex
	shutdownOnce sync.Once

	stopReader context.CancelFunc

	doneChan chan struct{}
	doneErr  error
	doneOnce sync.Once
}

// Listen returns a new listener on the SRT protocol on the address with
// the provided config. The network parameter needs to be "srt".
//
// The address has the form "host:port".
//
// Examples:
//
//	Listen("srt", "127.0.0.1:3000", DefaultConfig())
//
// In case of an error, the returned Listener is nil and the error is non-nil.
func Listen(network, address string, config Config) (Listener, error) {
	if network != "srt" {
		return nil, fmt.Errorf("listen: the network must be 'srt'")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("listen: invalid config: %w", err)
	}

	if config.Logger == nil {
		config.Logger = NewLogger(nil)
	}

	ln := &listener{
		config: config,
	}

	lc := net.ListenConfig{
		Control: ListenControl(config),
	}

	network = "udp"

	ip, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("listen: invalid address: %w", err)
	}

	if len(ip) != 0 {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return nil, fmt.Errorf("listen: invalid address: %w", err)
		}

		if addr.Is4() {
			network = "udp4"
		} else if addr.Is6() {
			network = "udp6"
		}
	}

	lp, err := lc.ListenPacket(context.Background(), network, address)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	pc := lp.(*net.UDPConn)

	ln.addr = pc.LocalAddr()
	if ln.addr == nil {
		return nil, fmt.Errorf("listen: no local address")
	}

	ln.pc = newPacketConn(pc)
	ln.conns = make(map[uint32]*srtConn)
	ln.connsByPeer = make(map[uint32]*srtConn)

	ln.backlog = make(chan packet.Packet, 128)

	ln.rcvQueue = make(chan packet.Packet, 2048)

	syncookie, err := srtnet.NewSYNCookie(ln.addr.String(), nil)
	if err != nil {
		ln.Close()
		return nil, err
	}
	ln.syncookie = syncookie

	ln.doneChan = make(chan struct{})

	ln.start = time.Now()

	var readerCtx context.Context
	readerCtx, ln.stopReader = context.WithCancel(context.Background())
	go ln.reader(readerCtx)

	go func() {
		buffer := make([]byte, config.MSS) // MTU size

		for {
			if ln.isShutdown() {
				ln.markDone(ErrListenerClosed)
				return
			}

			ln.pc.SetReadDeadline(time.Now().Add(3 * time.Second))
			n, addr, localIP, err := ln.pc.readFromTo(buffer)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}

				if ln.isShutdown() {
					ln.markDone(ErrListenerClosed)
					return
				}

				ln.markDone(err)
				return
			}

			p, err := packet.NewPacketFromData(addr, buffer[:n])
			if err != nil {
				continue
			}

			if localIP != nil {
				laddr := ln.addr.(*net.UDPAddr)
				p.Header().LocalAddr = &net.UDPAddr{IP: localIP, Port: laddr.Port}
			}

			// non-blocking
			select {
			case ln.rcvQueue <- p:
			default:
				ln.log("listen", func() string { return "receive queue is full" })
			}
		}
	}()

	return ln, nil
}

func (ln *listener) Accept2() (ConnRequest, error) {
	if ln.isShutdown() {
		return nil, ErrListenerClosed
	}

	for {
		select {
		case <-ln.doneChan:
			return nil, ln.error()

		case p := <-ln.backlog:
			req := newConnRequest(ln, p)
			if req == nil {
				break
			}

			return req, nil
		}
	}
}

func (ln *listener) Accept(acceptFn AcceptFunc) (Conn, ConnType, error) {
	for {
		req, err := ln.Accept2()
		if err != nil {
			return nil, REJECT, err
		}

		if acceptFn == nil {
			req.Reject(REJ_PEER)
			continue
		}

		mode := acceptFn(req)
		if mode != PUBLISH && mode != SUBSCRIBE {
			// Figure out the reason
			reason := REJ_PEER
			if req.(*connRequest).rejectionReason > 0 {
				reason = req.(*connRequest).rejectionReason
			}
			req.Reject(reason)
			continue
		}

		conn, err := req.Accept()
		if err != nil {
			continue
		}

		return conn, mode, nil
	}
}

// markDone marks the listener as done by closing
// the done channel & sets the error
func (ln *listener) markDone(err error) {
	ln.doneOnce.Do(func() {
		ln.lock.Lock()
		defer ln.lock.Unlock()
		ln.doneErr = err
		close(ln.doneChan)
	})
}

// error returns the error that caused the listener to be done
// if it's nil then the listener is not done
func (ln *listener) error() error {
	ln.lock.Lock()
	defer ln.lock.Unlock()
	return ln.doneErr
}

func (ln *listener) handleShutdown(c *srtConn) {
	ln.lock.Lock()
	delete(ln.conns, c.socketId)
	delete(ln.connsByPeer, c.peerSocketId)
	ln.lock.Unlock()
}

func (ln *listener) isShutdown() bool {
	ln.shutdownLock.RLock()
	defer ln.shutdownLock.RUnlock()

	return ln.shutdown
}

func (ln *listener) Close() {
	ln.shutdownOnce.Do(func() {
		ln.shutdownLock.Lock()
		ln.shutdown = true
		ln.shutdownLock.Unlock()

		ln.lock.RLock()
		for _, conn := range ln.conns {
			if conn == nil {
				continue
			}
			conn.close()
		}
		ln.lock.RUnlock()

		ln.stopReader()

		ln.log("listen", func() string { return "closing socket" })

		ln.pc.Close()
	})
}

func (ln *listener) Addr() net.Addr {
	addrString := "0.0.0.0:0"
	if ln.addr != nil {
		addrString = ln.addr.String()
	}

	addr, _ := net.ResolveUDPAddr("udp", addrString)
	return addr
}

func (ln *listener) reader(ctx context.Context) {
	defer func() {
		ln.log("listen", func() string { return "left reader loop" })
	}()

	ln.log("listen", func() string { return "reader loop started" })

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-ln.rcvQueue:
			if ln.isShutdown() {
				break
			}

			ln.log("packet:recv:dump", func() string { return p.Dump() })

			if p.Header().DestinationSocketId == 0 {
				if p.Header().IsControlPacket && p.Header().ControlType == packet.CTRLTYPE_HANDSHAKE {
					select {
					case ln.backlog <- p:
					default:
						ln.log("handshake:recv:error", func() string { return "backlog is full" })
					}
				}
				break
			}

			ln.lock.RLock()
			conn, ok := ln.conns[p.Header().DestinationSocketId]
			ln.lock.RUnlock()

			if !ok || conn == nil {
				// ignore the packet, we don't know the destination
				break
			}

			if !ln.config.AllowPeerIpChange {
				if p.Header().Addr.String() != conn.RemoteAddr().String() {
					// ignore the packet, it's not from the expected peer
					// https://haivision.github.io/srt-rfc/draft-sharabayko-srt.html#name-security-considerations
					break
				}
			}

			conn.push(p)
		}
	}
}

// Send a packet to the wire. This function must be synchronous in order to allow to safely call Packet.Decommission() afterward.
func (ln *listener) send(p packet.Packet) {
	ln.sndMutex.Lock()
	defer ln.sndMutex.Unlock()

	ln.sndData.Reset()

	if err := p.Marshal(&ln.sndData); err != nil {
		p.Decommission()
		ln.log("packet:send:error", func() string { return "marshalling packet failed" })
		return
	}

	buffer := ln.sndData.Bytes()

	ln.log("packet:send:dump", func() string { return p.Dump() })

	// Write the packet's contents to the wire
	ln.pc.writeToFrom(buffer, p.Header().Addr, p.Header().LocalAddr)

	if p.Header().IsControlPacket {
		// Control packets can be decommissioned because they will not be sent again (data packets might be retransferred)
		p.Decommission()
	}
}

func (ln *listener) log(topic string, message func() string) {
	if ln.config.Logger == nil {
		return
	}

	ln.config.Logger.Print(topic, 0, 2, message)
}
````

## File: log_test.go
````go
package srt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasTopic(t *testing.T) {
	l := NewLogger([]string{
		"packet:recv:dump",
	})

	ok := l.HasTopic("foobar")
	require.False(t, ok)

	ok = l.HasTopic("packet:recv:dump")
	require.True(t, ok)

	ok = l.HasTopic("packet:recv")
	require.False(t, ok)

	ok = l.HasTopic("packet")
	require.False(t, ok)
}

var result bool

func BenchmarkHasTopicNil(b *testing.B) {
	l := NewLogger(nil)

	var r bool
	for n := 0; n < b.N; n++ {
		r = l.HasTopic("foobar")
	}

	result = r
}

func BenchmarkHasTopicD1(b *testing.B) {
	l := NewLogger([]string{
		"packet:recv:dump",
	})

	var r bool
	for n := 0; n < b.N; n++ {
		r = l.HasTopic("packet")
	}

	result = r
}

func BenchmarkHasTopicD2(b *testing.B) {
	l := NewLogger([]string{
		"packet:recv:dump",
	})

	var r bool
	for n := 0; n < b.N; n++ {
		r = l.HasTopic("packet:recv")
	}

	result = r
}

func BenchmarkHasTopicD3(b *testing.B) {
	l := NewLogger([]string{
		"packet:recv:dump",
	})

	var r bool
	for n := 0; n < b.N; n++ {
		r = l.HasTopic("packet:recv:dump")
	}

	result = r
}
````

## File: log.go
````go
package srt

import (
	"runtime"
	"strings"
	"time"
)

// Logger is for logging debug messages.
type Logger interface {
	// HasTopic returns whether this Logger is logging messages of that topic.
	HasTopic(topic string) bool

	// Print adds a new message to the message queue. The message itself is
	// a function that returns the string to be logges. It will only be
	// executed if HasTopic returns true on the given topic.
	Print(topic string, socketId uint32, skip int, message func() string)

	// Listen returns a read channel for Log messages.
	Listen() <-chan Log

	// Close closes the logger. No more messages will be logged.
	Close()
}

// logger implements a Logger
type logger struct {
	logQueue chan Log
	topics   map[string]bool
}

// NewLogger returns a Logger that only listens on the given list of topics.
func NewLogger(topics []string) Logger {
	l := &logger{
		logQueue: make(chan Log, 1024),
		topics:   make(map[string]bool),
	}

	for _, topic := range topics {
		l.topics[topic] = true
	}

	return l
}

func (l *logger) HasTopic(topic string) bool {
	if len(l.topics) == 0 {
		return false
	}

	if ok := l.topics[topic]; ok {
		return true
	}

	len := len(topic)

	for {
		i := strings.LastIndexByte(topic[:len], ':')
		if i == -1 {
			break
		}

		len = i

		if ok := l.topics[topic[:len]]; !ok {
			continue
		}

		return true
	}

	return false
}

func (l *logger) Print(topic string, socketId uint32, skip int, message func() string) {
	if !l.HasTopic(topic) {
		return
	}

	_, file, line, _ := runtime.Caller(skip)

	msg := Log{
		Time:     time.Now(),
		SocketId: socketId,
		Topic:    topic,
		Message:  message(),
		File:     file,
		Line:     line,
	}

	// Write to log queue, but don't block if it's full
	select {
	case l.logQueue <- msg:
	default:
	}
}

func (l *logger) Listen() <-chan Log {
	return l.logQueue
}

func (l *logger) Close() {
	close(l.logQueue)
}

// Log represents a log message
type Log struct {
	Time     time.Time // Time of when this message has been logged
	SocketId uint32    // The socketid if connection related, 0 otherwise
	Topic    string    // The topic of this message
	Message  string    // The message itself
	File     string    // The file in which this message has been dispatched
	Line     int       // The line number in the file in which this message has been dispatched
}
````

## File: net_windows.go
````go
//go:build windows

package srt

import (
	"syscall"

	"golang.org/x/sys/windows"
)

func ListenControl(config Config) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		var opErr error
		err := c.Control(func(fd uintptr) {
			// Set REUSEADDR
			opErr = windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_REUSEADDR, 1)
			if opErr != nil {
				return
			}

			// Set TOS
			if config.IPTOS > 0 {
				opErr = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, windows.IP_TOS, config.IPTOS)
				if opErr != nil {
					return
				}
			}

			// Set TTL
			if config.IPTTL > 0 {
				opErr = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, windows.IP_TTL, config.IPTTL)
				if opErr != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}
		return opErr
	}
}

func DialControl(config Config) func(network string, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		var opErr error
		err := c.Control(func(fd uintptr) {
			// Set TOS
			if config.IPTOS > 0 {
				opErr = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, windows.IP_TOS, config.IPTOS)
				if opErr != nil {
					return
				}
			}

			// Set TTL
			if config.IPTTL > 0 {
				opErr = windows.SetsockoptInt(windows.Handle(fd), windows.IPPROTO_IP, windows.IP_TTL, config.IPTTL)
				if opErr != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}
		return opErr
	}
}
````

## File: net.go
````go
//go:build !windows

package srt

import "syscall"

func ListenControl(config Config) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		var opErr error
		err := c.Control(func(fd uintptr) {
			// Set REUSEADDR
			opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			if opErr != nil {
				return
			}

			// Set TOS
			if config.IPTOS > 0 {
				opErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TOS, config.IPTOS)
				if opErr != nil {
					return
				}
			}

			// Set TTL
			if config.IPTTL > 0 {
				opErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TTL, config.IPTTL)
				if opErr != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}
		return opErr
	}
}

func DialControl(config Config) func(network string, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		var opErr error
		err := c.Control(func(fd uintptr) {
			// Set TOS
			if config.IPTOS > 0 {
				opErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TOS, config.IPTOS)
				if opErr != nil {
					return
				}
			}

			// Set TTL
			if config.IPTTL > 0 {
				opErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TTL, config.IPTTL)
				if opErr != nil {
					return
				}
			}
		})
		if err != nil {
			return err
		}
		return opErr
	}
}
````

## File: net/ip_test.go
````go
package net

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIPDefault(t *testing.T) {
	ip := IP{}

	ip.setDefault()

	require.Equal(t, "127.0.0.1", ip.String())
}

func TestIPParse(t *testing.T) {
	ip := IP{}

	ip.Parse("192.168.1.3")

	require.Equal(t, "192.168.1.3", ip.String())

	ip.Parse("fhdhdf")

	require.Equal(t, "127.0.0.1", ip.String())
}

func TestIPFrom(t *testing.T) {
	ip := IP{}

	ip.FromNetIP(net.ParseIP("192.168.2.56"))

	require.Equal(t, "192.168.2.56", ip.String())

	ip.FromNetIP(net.ParseIP("127.0.0.1"))

	require.Equal(t, "127.0.0.1", ip.String())

	udpaddr, err := net.ResolveUDPAddr("udp", "localhost:12345")

	require.NoError(t, err)
	ip.FromNetAddr(udpaddr)

	require.Equal(t, "127.0.0.1", ip.String())

	ipaddr, err := net.ResolveIPAddr("ip", "localhost")

	require.NoError(t, err)
	ip.FromNetAddr(ipaddr)

	require.Equal(t, "127.0.0.1", ip.String())
}

func TestIPUnmarshal(t *testing.T) {
	ip := IP{}

	b0 := [5]byte{}

	err := ip.Unmarshal(b0[:])

	require.Error(t, err)

	b1 := [...]byte{1, 0, 168, 192}

	err = ip.Unmarshal(b1[:])

	require.NoError(t, err)
	require.Equal(t, "192.168.0.1", ip.String())

	b2 := [...]byte{1, 0, 168, 192, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	err = ip.Unmarshal(b2[:])

	require.NoError(t, err)
	require.Equal(t, "192.168.0.1", ip.String())

	b3 := [...]byte{1, 0, 0, 0, 0, 0, 0, 0, 0xc5, 0x71, 0x26, 0xdb, 0x94, 0x8c, 0x30, 0xfd}

	err = ip.Unmarshal(b3[:])

	require.NoError(t, err)
	require.Equal(t, "fd30:8c94:db26:71c5::1", ip.String())
}

func TestIPMarshal(t *testing.T) {
	ip := IP{}

	ip.Parse("192.168.0.1")

	b := [16]byte{}

	ip.Marshal(b[:])

	require.Equal(t, [...]byte{1, 0, 168, 192, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, b)

	ip.Parse("fd30:8c94:db26:71c5::1")

	ip.Marshal(b[:])

	require.Equal(t, [...]byte{1, 0, 0, 0, 0, 0, 0, 0, 0xc5, 0x71, 0x26, 0xdb, 0x94, 0x8c, 0x30, 0xfd}, b)
}
````

## File: net/ip.go
````go
package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type IP struct {
	ip net.IP
}

func (i *IP) setDefault() {
	i.ip = net.IPv4(127, 0, 0, 1)
}

func (i *IP) isValid() bool {
	if i.ip == nil || i.ip.String() == "<nil>" || i.ip.IsUnspecified() {
		return false
	}

	return true
}

func (i IP) String() string {
	if i.ip == nil {
		return ""
	}

	return i.ip.String()
}

func (i *IP) Parse(ip string) {
	i.setDefault()

	iip := net.ParseIP(ip)
	if iip == nil {
		return
	}

	i.ip = iip

	if !i.isValid() {
		i.setDefault()
	}
}

func (i *IP) FromNetIP(ip net.IP) {
	if ip == nil {
		return
	}

	iip := net.ParseIP(ip.String())
	if iip == nil {
		return
	}

	i.ip = iip

	if !i.isValid() {
		i.setDefault()
	}
}

func (i *IP) FromNetAddr(addr net.Addr) {
	if addr == nil {
		i.setDefault()
		return
	}

	if addr.Network() != "udp" {
		i.setDefault()
		return
	}

	if a, err := net.ResolveUDPAddr("udp", addr.String()); err == nil {
		if a == nil || a.IP == nil {
			i.setDefault()
		} else {
			i.ip = a.IP
		}
	} else {
		i.setDefault()
	}
}

// Unmarshal converts 16 bytes in host byte order to IP
func (i *IP) Unmarshal(data []byte) error {
	if len(data) != 4 && len(data) != 16 {
		return fmt.Errorf("invalid number of bytes")
	}

	if len(data) == 4 {
		ip0 := binary.LittleEndian.Uint32(data[0:])

		i.ip = net.IPv4(byte((ip0&0xff000000)>>24), byte((ip0&0x00ff0000)>>16), byte((ip0&0x0000ff00)>>8), byte(ip0&0x0000ff))
	} else {
		ip3 := binary.LittleEndian.Uint32(data[0:])
		ip2 := binary.LittleEndian.Uint32(data[4:])
		ip1 := binary.LittleEndian.Uint32(data[8:])
		ip0 := binary.LittleEndian.Uint32(data[12:])

		if ip0 == 0 && ip1 == 0 && ip2 == 0 {
			i.ip = net.IPv4(byte((ip3&0xff000000)>>24), byte((ip3&0x00ff0000)>>16), byte((ip3&0x0000ff00)>>8), byte(ip3&0x0000ff))
		} else {
			var b strings.Builder

			fmt.Fprintf(&b, "%04x:", (ip0&0xffff0000)>>16)
			fmt.Fprintf(&b, "%04x:", ip0&0x0000ffff)
			fmt.Fprintf(&b, "%04x:", (ip1&0xffff0000)>>16)
			fmt.Fprintf(&b, "%04x:", ip1&0x0000ffff)
			fmt.Fprintf(&b, "%04x:", (ip2&0xffff0000)>>16)
			fmt.Fprintf(&b, "%04x:", ip2&0x0000ffff)
			fmt.Fprintf(&b, "%04x:", (ip3&0xffff0000)>>16)
			fmt.Fprintf(&b, "%04x", ip3&0x0000ffff)

			iip := net.ParseIP(b.String())
			if iip == nil {
				return fmt.Errorf("invalid ip")
			}

			i.ip = iip
		}
	}

	if !i.isValid() {
		i.setDefault()
	}

	return nil
}

// Marshal converts an IP to 16 byte host byte order
func (i *IP) Marshal(data []byte) {
	if i.ip == nil || !i.isValid() {
		i.setDefault()
	}

	if len([]byte(i.ip)) == 4 {
		i.ip = net.IPv4(i.ip[0], i.ip[1], i.ip[2], i.ip[3])
	}

	if len(data) < 16 {
		return
	}

	data[0] = i.ip[15]
	data[1] = i.ip[14]
	data[2] = i.ip[13]
	data[3] = i.ip[12]

	if i.ip.To4() != nil {
		data[4] = 0
		data[5] = 0
		data[6] = 0
		data[7] = 0

		data[8] = 0
		data[9] = 0
		data[10] = 0
		data[11] = 0

		data[12] = 0
		data[13] = 0
		data[14] = 0
		data[15] = 0
	} else {
		data[4] = i.ip[11]
		data[5] = i.ip[10]
		data[6] = i.ip[9]
		data[7] = i.ip[8]

		data[8] = i.ip[7]
		data[9] = i.ip[6]
		data[10] = i.ip[5]
		data[11] = i.ip[4]

		data[12] = i.ip[3]
		data[13] = i.ip[2]
		data[14] = i.ip[1]
		data[15] = i.ip[0]
	}
}
````

## File: net/syncookie_test.go
````go
package net

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSYNCookie(t *testing.T) {
	counter := int64(0)

	s, err := NewSYNCookie("192.168.0.1", func() int64 {
		return counter
	})
	require.NoError(t, err)

	s.secret1 = "dl2INvNSQTZ5zQu9MxNmGyAVmNkB33io"
	s.secret2 = "nwj2qrsh3xyC8OmCp1gObD0iOtQNQsLi"

	cookie := s.Get("192.168.0.2")

	require.Equal(t, uint32(0xe6303651), cookie)

	require.True(t, s.Verify(cookie, "192.168.0.2"))

	require.False(t, s.Verify(cookie, "192.168.0.3"))
	require.False(t, s.Verify(cookie-95854, "192.168.0.2"))

	counter = 1

	require.True(t, s.Verify(cookie, "192.168.0.2"))

	counter = 2

	require.False(t, s.Verify(cookie, "192.168.0.2"))
}
````

## File: net/syncookie.go
````go
package net

import (
	"crypto/md5"
	"encoding/binary"
	"strconv"
	"time"

	"github.com/datarhei/gosrt/rand"
)

// SYNCookie implements a syn cookie for the SRT handshake.
type SYNCookie struct {
	secret1 string
	secret2 string
	daddr   string
	counter func() int64
}

func defaultCounter() int64 {
	return time.Now().Unix() >> 6
}

// NewSYNCookie returns a SYNCookie for a destination address.
func NewSYNCookie(daddr string, counter func() int64) (*SYNCookie, error) {
	s := &SYNCookie{
		daddr:   daddr,
		counter: counter,
	}

	if s.counter == nil {
		s.counter = defaultCounter
	}

	var err error
	s.secret1, err = rand.RandomString(32, rand.AlphaNumericCharset)
	if err != nil {
		return nil, err
	}

	s.secret2, err = rand.RandomString(32, rand.AlphaNumericCharset)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Get returns the current syn cookie with a sender address.
func (s *SYNCookie) Get(saddr string) uint32 {
	return s.calculate(s.counter(), saddr)
}

// Verify verfies that two syn cookies relate.
func (s *SYNCookie) Verify(cookie uint32, saddr string) bool {
	counter := s.counter()

	if s.calculate(counter, saddr) == cookie {
		return true
	}

	if s.calculate(counter-1, saddr) == cookie {
		return true
	}

	return false
}

func (s *SYNCookie) calculate(counter int64, saddr string) uint32 {
	data := s.secret1 + s.daddr + saddr + s.secret2 + strconv.FormatInt(counter, 10)

	md5sum := md5.Sum([]byte(data))

	return binary.BigEndian.Uint32(md5sum[0:])
}
````

## File: packet_conn.go
````go
package srt

import (
	"net"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// this is net.PacketConn with two additional methods
// that allow to get and set local IP.
type packetConn struct {
	net.PacketConn

	pc4 *ipv4.PacketConn
	pc6 *ipv6.PacketConn
}

func newPacketConn(wrapped net.PacketConn) *packetConn {
	c := &packetConn{
		PacketConn: wrapped,
	}

	// Enable PKTINFO to capture the destination IP of incoming packets.
	// Both pc4 and pc6 are attempted independently: on a dual-stack AF_INET6
	// socket pc6 delivers the destination (IPv4-mapped for IPv4 packets) while
	// pc4.WriteTo is still needed to pin the IPv4 source address on replies.
	pc6 := ipv6.NewPacketConn(wrapped)
	if err := pc6.SetControlMessage(ipv6.FlagDst, true); err == nil {
		c.pc6 = pc6
	}
	pc4 := ipv4.NewPacketConn(wrapped)
	if err := pc4.SetControlMessage(ipv4.FlagDst, true); err == nil {
		c.pc4 = pc4
	}

	return c
}

func (c *packetConn) readFromTo(buffer []byte) (int, net.Addr, net.IP, error) {
	switch {
	case c.pc6 != nil:
		n, cm, addr, err := c.pc6.ReadFrom(buffer)
		if cm != nil && cm.Dst != nil {
			// Normalize IPv4-mapped addresses (::ffff:x.x.x.x) to plain IPv4
			if ip4 := cm.Dst.To4(); ip4 != nil {
				return n, addr, ip4, err
			}
			return n, addr, cm.Dst, err
		}
		return n, addr, nil, err

	case c.pc4 != nil:
		n, cm, addr, err := c.pc4.ReadFrom(buffer)
		if cm != nil {
			return n, addr, cm.Dst, err
		}
		return n, addr, nil, err

	default:
		n, addr, err := c.ReadFrom(buffer)
		return n, addr, nil, err
	}
}

func (c *packetConn) writeToFrom(buffer []byte, remoteAddr net.Addr, localAddr net.Addr) {
	if localAddrUDP, ok := localAddr.(*net.UDPAddr); ok && localAddrUDP != nil {
		if _, ok := remoteAddr.(*net.UDPAddr); ok {
			// For IPv4 destinations use pc4 even on dual-stack sockets
			if ip4 := localAddrUDP.IP.To4(); ip4 != nil && c.pc4 != nil {
				c.pc4.WriteTo(buffer, &ipv4.ControlMessage{Src: ip4}, remoteAddr)
				return
			}
			if c.pc6 != nil {
				c.pc6.WriteTo(buffer, &ipv6.ControlMessage{Src: localAddrUDP.IP}, remoteAddr)
				return
			}
		}
	}

	c.WriteTo(buffer, remoteAddr)
}
````

## File: packet/packet_test.go
````go
package packet

import (
	"bytes"
	"encoding/hex"
	"net"
	"sync"
	"testing"

	"github.com/datarhei/gosrt/circular"
	srtnet "github.com/datarhei/gosrt/net"

	"github.com/stretchr/testify/require"
)

func TestEmptyPacket(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	p := NewPacket(addr)

	var buf bytes.Buffer

	p.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "00000000c00000010000000000000000", data)
}

func TestArbitraryPacket(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	p := NewPacket(addr)
	p.SetData([]byte("hello world!"))

	var buf bytes.Buffer

	p.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "00000000c0000001000000000000000068656c6c6f20776f726c6421", data)
}

func TestArbitraryControlPacket(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	p := NewPacket(addr)
	p.Header().IsControlPacket = true
	p.Header().ControlType = CTRLTYPE_KEEPALIVE
	p.Header().SubType = 112
	p.Header().TypeSpecific = 42

	var buf bytes.Buffer

	p.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "800100700000002a0000000000000000", data)
}

func FuzzPacket(f *testing.F) {
	f.Add("00000000c00000010000000000000000")
	f.Add("00000000c0000001000000000000000068656c6c6f20776f726c6421")
	f.Add("800100700000002a0000000000000000")

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	f.Fuzz(func(t *testing.T, orig string) {
		data, err := hex.DecodeString(orig)
		if err != nil {
			return
		}
		if len(data) == 0 {
			return
		}
		p, err := NewPacketFromData(addr, data)
		if err != nil {
			return
		}

		var buf bytes.Buffer
		buf.Reset()
		p.Marshal(&buf)

		if !bytes.Equal(data, buf.Bytes()) {
			t.Errorf("Before: %q, after: %q\n%s", orig, hex.EncodeToString(buf.Bytes()), p.Dump())
		}
	})
}

func TestUnmarshalPacket(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	data, _ := hex.DecodeString("00000000c0000001000000000000000068656c6c6f20776f726c6421")

	p, err := NewPacketFromData(addr, data)
	require.NoError(t, err)

	require.Equal(t, p.Header().Timestamp, uint32(0))
	require.Equal(t, p.Header().Addr.String(), "127.0.0.1:6000")
	require.False(t, p.Header().IsControlPacket)
	require.Equal(t, p.Header().PacketPositionFlag, SinglePacket)
	require.Equal(t, p.Header().KeyBaseEncryptionFlag, UnencryptedPacket)
	require.Equal(t, p.Header().MessageNumber, uint32(1))

	require.Equal(t, uint64(12), p.Len())
	require.Equal(t, "hello world!", string(p.Data()))
}

func TestPacketString(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	p := NewPacket(addr)
	p.SetData([]byte("hello world!"))

	require.Greater(t, len(p.String()), 0)
}

func TestHandshakeV4(t *testing.T) {
	ip := srtnet.IP{}
	ip.Parse("127.0.0.1")

	cif := &CIFHandshake{
		IsRequest:                   false,
		Version:                     4,
		EncryptionField:             0,
		ExtensionField:              2,
		InitialPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     1500,
		MaxFlowWindowSize:           100,
		HandshakeType:               HSTYPE_CONCLUSION,
		SRTSocketId:                 0x274921,
		SynCookie:                   0x123456,
		PeerIP:                      ip,
		HasHS:                       false,
		HasKM:                       false,
		HasSID:                      false,
		HasCongestionCtl:            false,
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "00000004000000020000002a000005dc00000064ffffffff00274921001234560100007f000000000000000000000000", data)

	cif2 := &CIFHandshake{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestHandshakeV5(t *testing.T) {
	ip := srtnet.IP{}
	ip.Parse("127.0.0.1")

	cif := &CIFHandshake{
		IsRequest:                   false,
		Version:                     5,
		EncryptionField:             0,
		ExtensionField:              0,
		InitialPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     1500,
		MaxFlowWindowSize:           100,
		HandshakeType:               HSTYPE_CONCLUSION,
		SRTSocketId:                 0x274921,
		SynCookie:                   0x123456,
		PeerIP:                      ip,
		HasHS:                       true,
		HasKM:                       true,
		HasSID:                      true,
		HasCongestionCtl:            true,
		SRTHS: &CIFHandshakeExtension{
			SRTVersion: 0x010402,
			SRTFlags: CIFHandshakeExtensionFlags{
				TSBPDSND:      true,
				TSBPDRCV:      true,
				CRYPT:         true,
				TLPKTDROP:     true,
				PERIODICNAK:   true,
				REXMITFLG:     true,
				STREAM:        false,
				PACKET_FILTER: false,
			},
			RecvTSBPDDelay: 100,
			SendTSBPDDelay: 100,
		},
		SRTKM: &CIFKeyMaterialExtension{
			S:                     0,
			Version:               1,
			PacketType:            2,
			Sign:                  0x2029,
			Resv1:                 0,
			KeyBasedEncryption:    EvenKeyEncrypted,
			KeyEncryptionKeyIndex: 0,
			Cipher:                2,
			Authentication:        0,
			StreamEncapsulation:   2,
			Resv2:                 0,
			Resv3:                 0,
			SLen:                  16,
			KLen:                  16,
			Salt:                  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
			Wrap:                  []byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
		},
		StreamId:      "/live/stream.foobar",
		CongestionCtl: "foob",
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "00000005000200070000002a000005dc00000064ffffffff00274921001234560100007f00000000000000000000000000020003000104020000003f006400640004000e122029010000000002000200000004040102030405060708090a0b0c0d0e0f10f0f1f2f3f4f5f6f71112131415161718191a1b1c1d1e1f200005000576696c2f74732f656d6165726f6f662e0072616200060001626f6f66", data)

	cif2 := &CIFHandshake{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestHandshakeV5UnsupportedExtension(t *testing.T) {
	ip := srtnet.IP{}
	ip.Parse("127.0.0.1")

	cif := &CIFHandshake{
		IsRequest:                   false,
		Version:                     5,
		EncryptionField:             0,
		ExtensionField:              0,
		InitialPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     1500,
		MaxFlowWindowSize:           100,
		HandshakeType:               HSTYPE_CONCLUSION,
		SRTSocketId:                 0x274921,
		SynCookie:                   0x123456,
		PeerIP:                      ip,
		HasHS:                       true,
		HasKM:                       false,
		HasSID:                      true,
		HasCongestionCtl:            true,
		SRTHS: &CIFHandshakeExtension{
			SRTVersion: 0x010402,
			SRTFlags: CIFHandshakeExtensionFlags{
				TSBPDSND:      true,
				TSBPDRCV:      true,
				CRYPT:         true,
				TLPKTDROP:     true,
				PERIODICNAK:   true,
				REXMITFLG:     true,
				STREAM:        false,
				PACKET_FILTER: false,
			},
			RecvTSBPDDelay: 100,
			SendTSBPDDelay: 100,
		},
		StreamId:      "/live/stream.foobar",
		CongestionCtl: "foob",
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "00000005000000050000002a000005dc00000064ffffffff00274921001234560100007f00000000000000000000000000020003000104020000003f006400640005000576696c2f74732f656d6165726f6f662e0072616200060001626f6f66", data)

	// Adding an unknown extension bd01000400000000000000000000000000000000 (extensionType = 48385, extensionLen = 4)
	decoded, err := hex.DecodeString("00000005000000050000002a000005dc00000064ffffffff00274921001234560100007f00000000000000000000000000020003000104020000003f00640064bd010004000000000000000000000000000000000005000576696c2f74732f656d6165726f6f662e0072616200060001626f6f66")
	require.NoError(t, err)

	buf.Reset()
	buf.Write(decoded)

	cif2 := &CIFHandshake{}

	err = cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestHandshakeString(t *testing.T) {
	ip := srtnet.IP{}
	ip.Parse("127.0.0.1")

	cif := &CIFHandshake{
		IsRequest:                   false,
		Version:                     5,
		EncryptionField:             0,
		ExtensionField:              0,
		InitialPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		MaxTransmissionUnitSize:     1500,
		MaxFlowWindowSize:           100,
		HandshakeType:               HSTYPE_CONCLUSION,
		SRTSocketId:                 0x274921,
		SynCookie:                   0x123456,
		PeerIP:                      ip,
		HasHS:                       true,
		HasKM:                       false,
		HasSID:                      true,
		HasCongestionCtl:            false,
		SRTHS: &CIFHandshakeExtension{
			SRTVersion: 0x010402,
			SRTFlags: CIFHandshakeExtensionFlags{
				TSBPDSND:      true,
				TSBPDRCV:      true,
				CRYPT:         true,
				TLPKTDROP:     true,
				PERIODICNAK:   true,
				REXMITFLG:     true,
				STREAM:        false,
				PACKET_FILTER: false,
			},
			RecvTSBPDDelay: 100,
			SendTSBPDDelay: 100,
		},
		SRTKM:    nil,
		StreamId: "/live/stream.foobar",
	}

	require.Greater(t, len(cif.String()), 0)
}

func TestKM(t *testing.T) {
	cif := &CIFKeyMaterialExtension{
		S:                     0,
		Version:               1,
		PacketType:            2,
		Sign:                  0x2029,
		Resv1:                 0,
		KeyBasedEncryption:    EvenKeyEncrypted,
		KeyEncryptionKeyIndex: 0,
		Cipher:                2,
		Authentication:        0,
		StreamEncapsulation:   2,
		Resv2:                 0,
		Resv3:                 0,
		SLen:                  16,
		KLen:                  16,
		Salt:                  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		Wrap:                  []byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "122029010000000002000200000004040102030405060708090a0b0c0d0e0f10f0f1f2f3f4f5f6f71112131415161718191a1b1c1d1e1f20", data)

	cif2 := &CIFKeyMaterialExtension{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestKMString(t *testing.T) {
	cif := &CIFKeyMaterialExtension{
		S:                     0,
		Version:               1,
		PacketType:            2,
		Sign:                  0x2029,
		Resv1:                 0,
		KeyBasedEncryption:    EvenKeyEncrypted,
		KeyEncryptionKeyIndex: 0,
		Cipher:                2,
		Authentication:        0,
		StreamEncapsulation:   2,
		Resv2:                 0,
		Resv3:                 0,
		SLen:                  16,
		KLen:                  16,
		Salt:                  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10},
		Wrap:                  []byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20},
	}

	require.Greater(t, len(cif.String()), 0)
}

func TestFullACK(t *testing.T) {
	cif := &CIFACK{
		IsLite:                      false,
		IsSmall:                     false,
		LastACKPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		RTT:                         38473,
		RTTVar:                      9084,
		AvailableBufferSize:         48533,
		PacketsReceivingRate:        20,
		EstimatedLinkCapacity:       0,
		ReceivingRate:               73637,
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "0000002a000096490000237c0000bd95000000140000000000011fa5", data)

	cif2 := &CIFACK{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestFullACKString(t *testing.T) {
	cif := &CIFACK{
		IsLite:                      false,
		IsSmall:                     false,
		LastACKPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		RTT:                         38473,
		RTTVar:                      9084,
		AvailableBufferSize:         48533,
		PacketsReceivingRate:        20,
		EstimatedLinkCapacity:       0,
		ReceivingRate:               73637,
	}

	require.Greater(t, len(cif.String()), 0)
}

func TestSmallACK(t *testing.T) {
	cif := &CIFACK{
		IsLite:                      false,
		IsSmall:                     true,
		LastACKPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		RTT:                         38473,
		RTTVar:                      9084,
		AvailableBufferSize:         48533,
		PacketsReceivingRate:        0,
		EstimatedLinkCapacity:       0,
		ReceivingRate:               0,
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "0000002a000096490000237c0000bd95", data)

	cif2 := &CIFACK{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestSmallACKString(t *testing.T) {
	cif := &CIFACK{
		IsLite:                      false,
		IsSmall:                     true,
		LastACKPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		RTT:                         38473,
		RTTVar:                      9084,
		AvailableBufferSize:         48533,
		PacketsReceivingRate:        0,
		EstimatedLinkCapacity:       0,
		ReceivingRate:               0,
	}

	require.Greater(t, len(cif.String()), 0)
}

func TestLiteACK(t *testing.T) {
	cif := &CIFACK{
		IsLite:                      true,
		IsSmall:                     false,
		LastACKPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		RTT:                         0,
		RTTVar:                      0,
		AvailableBufferSize:         0,
		PacketsReceivingRate:        0,
		EstimatedLinkCapacity:       0,
		ReceivingRate:               0,
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "0000002a", data)

	cif2 := &CIFACK{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestLiteACKString(t *testing.T) {
	cif := &CIFACK{
		IsLite:                      true,
		IsSmall:                     false,
		LastACKPacketSequenceNumber: circular.New(42, MAX_SEQUENCENUMBER),
		RTT:                         0,
		RTTVar:                      0,
		AvailableBufferSize:         0,
		PacketsReceivingRate:        0,
		EstimatedLinkCapacity:       0,
		ReceivingRate:               0,
	}

	require.Greater(t, len(cif.String()), 0)
}

func TestNAK(t *testing.T) {
	cif := &CIFNAK{
		LostPacketSequenceNumber: []circular.Number{
			circular.New(42, MAX_SEQUENCENUMBER),
			circular.New(42, MAX_SEQUENCENUMBER),
			circular.New(45, MAX_SEQUENCENUMBER),
			circular.New(49, MAX_SEQUENCENUMBER),
		},
	}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "0000002a8000002d00000031", data)

	cif2 := &CIFNAK{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestNAKString(t *testing.T) {
	cif := &CIFNAK{
		LostPacketSequenceNumber: []circular.Number{
			circular.New(42, MAX_SEQUENCENUMBER),
			circular.New(42, MAX_SEQUENCENUMBER),
			circular.New(45, MAX_SEQUENCENUMBER),
			circular.New(49, MAX_SEQUENCENUMBER),
		},
	}

	require.Greater(t, len(cif.String()), 0)
}

func TestShutdown(t *testing.T) {
	cif := &CIFShutdown{}

	var buf bytes.Buffer

	cif.Marshal(&buf)

	data := hex.EncodeToString(buf.Bytes())

	require.Equal(t, "00000000", data)

	cif2 := &CIFShutdown{}

	err := cif2.Unmarshal(buf.Bytes())

	require.NoError(t, err)
	require.Equal(t, cif, cif2)
}

func TestShutdownString(t *testing.T) {
	cif := &CIFShutdown{}

	require.Greater(t, len(cif.String()), 0)
}

func BenchmarkNewPacket(b *testing.B) {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	for i := 0; i < b.N; i++ {
		pkt := NewPacket(addr)

		pkt.Decommission()
	}
}

func BenchmarkNewPacketWithData(b *testing.B) {
	data := make([]byte, 1316)
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:6000")

	p := NewPacket(addr)
	p.SetData(data)

	var buf bytes.Buffer

	p.Marshal(&buf)

	data = buf.Bytes()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pkt, _ := NewPacketFromData(addr, data)

		if pkt != nil {
			pkt.Decommission()
		}
	}
}

func BenchmarkNoBufferpool(b *testing.B) {
	data := make([]byte, 1316)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pdata := make([]byte, len(data)-16)
		copy(pdata, data[16:])
	}
}

func BenchmarkBufferpool(b *testing.B) {
	pool := sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	data := make([]byte, 1316)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := pool.Get().(*bytes.Buffer)

		p.Reset()
		p.Write(data[16:])

		pool.Put(p)
	}
}
````

## File: packet/packet.go
````go
// Package packet provides types and implementations for the different SRT packet types
package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/datarhei/gosrt/circular"
	srtnet "github.com/datarhei/gosrt/net"
)

const MAX_SEQUENCENUMBER uint32 = 0b01111111_11111111_11111111_11111111
const MAX_TIMESTAMP uint32 = 0b11111111_11111111_11111111_11111111
const MAX_PAYLOAD_SIZE = 1456

// Table 1: SRT Control Packet Types
type CtrlType uint16

const (
	CTRLTYPE_HANDSHAKE CtrlType = 0x0000
	CTRLTYPE_KEEPALIVE CtrlType = 0x0001
	CTRLTYPE_ACK       CtrlType = 0x0002
	CTRLTYPE_NAK       CtrlType = 0x0003
	CTRLTYPE_WARN      CtrlType = 0x0004 // unimplemented, receiver->sender
	CTRLTYPE_SHUTDOWN  CtrlType = 0x0005
	CTRLTYPE_ACKACK    CtrlType = 0x0006
	CRTLTYPE_DROPREQ   CtrlType = 0x0007 // unimplemented, sender->receiver
	CRTLTYPE_PEERERROR CtrlType = 0x0008 // unimplemented, receiver->sender (only for file transfers)
	CTRLTYPE_USER      CtrlType = 0x7FFF
)

func (h CtrlType) String() string {
	switch h {
	case CTRLTYPE_HANDSHAKE:
		return "HANDSHAKE"
	case CTRLTYPE_KEEPALIVE:
		return "KEEPALIVE"
	case CTRLTYPE_ACK:
		return "ACK"
	case CTRLTYPE_NAK:
		return "NAK"
	case CTRLTYPE_WARN:
		return "WARN"
	case CTRLTYPE_SHUTDOWN:
		return "SHUTDOWN"
	case CTRLTYPE_ACKACK:
		return "ACKACK"
	case CRTLTYPE_DROPREQ:
		return "DROPREQ"
	case CRTLTYPE_PEERERROR:
		return "PEERERROR"
	case CTRLTYPE_USER:
		return "USER"
	}

	return "unknown"
}

func (h CtrlType) Value() uint16 {
	return uint16(h)
}

type HandshakeType uint32

// Table 4: Handshake Type
const (
	HSTYPE_DONE       HandshakeType = 0xFFFFFFFD
	HSTYPE_AGREEMENT  HandshakeType = 0xFFFFFFFE
	HSTYPE_CONCLUSION HandshakeType = 0xFFFFFFFF
	HSTYPE_WAVEHAND   HandshakeType = 0x00000000
	HSTYPE_INDUCTION  HandshakeType = 0x00000001
)

func (h HandshakeType) String() string {
	switch h {
	case HSTYPE_DONE:
		return "DONE"
	case HSTYPE_AGREEMENT:
		return "AGREEMENT"
	case HSTYPE_CONCLUSION:
		return "CONCLUSION"
	case HSTYPE_WAVEHAND:
		return "WAVEHAND"
	case HSTYPE_INDUCTION:
		return "INDUCTION"
	}

	return "REJECT (" + strconv.FormatUint(uint64(h), 32) + ")"
}

func (h HandshakeType) IsHandshake() bool {
	switch h {
	case HSTYPE_DONE:
	case HSTYPE_AGREEMENT:
	case HSTYPE_CONCLUSION:
	case HSTYPE_WAVEHAND:
	case HSTYPE_INDUCTION:
	default:
		return false
	}

	return true
}

func (h HandshakeType) IsRejection() bool {
	return !h.IsHandshake()
}

func (h HandshakeType) Val() uint32 {
	return uint32(h)
}

// Table 6: Handshake Extension Message Flags
const (
	SRTFLAG_TSBPDSND      uint32 = 1 << 0
	SRTFLAG_TSBPDRCV      uint32 = 1 << 1
	SRTFLAG_CRYPT         uint32 = 1 << 2
	SRTFLAG_TLPKTDROP     uint32 = 1 << 3
	SRTFLAG_PERIODICNAK   uint32 = 1 << 4
	SRTFLAG_REXMITFLG     uint32 = 1 << 5
	SRTFLAG_STREAM        uint32 = 1 << 6
	SRTFLAG_PACKET_FILTER uint32 = 1 << 7
)

// Table 5: Handshake Extension Type values
type CtrlSubType uint16

const (
	CTRLSUBTYPE_NONE   CtrlSubType = 0
	EXTTYPE_HSREQ      CtrlSubType = 1
	EXTTYPE_HSRSP      CtrlSubType = 2
	EXTTYPE_KMREQ      CtrlSubType = 3
	EXTTYPE_KMRSP      CtrlSubType = 4
	EXTTYPE_SID        CtrlSubType = 5
	EXTTYPE_CONGESTION CtrlSubType = 6
	EXTTYPE_FILTER     CtrlSubType = 7 // unimplemented
	EXTTYPE_GROUP      CtrlSubType = 8 // unimplemented
)

func (h CtrlSubType) String() string {
	switch h {
	case CTRLSUBTYPE_NONE:
		return "NONE"
	case EXTTYPE_HSREQ:
		return "EXTTYPE_HSREQ"
	case EXTTYPE_HSRSP:
		return "EXTTYPE_HSRSP"
	case EXTTYPE_KMREQ:
		return "EXTTYPE_KMREQ"
	case EXTTYPE_KMRSP:
		return "EXTTYPE_KMRSP"
	case EXTTYPE_SID:
		return "EXTTYPE_SID"
	case EXTTYPE_CONGESTION:
		return "EXTTYPE_CONGESTION"
	case EXTTYPE_FILTER:
		return "EXTTYPE_FILTER"
	case EXTTYPE_GROUP:
		return "EXTTYPE_GROUP"
	}

	return "unknown"
}

func (h CtrlSubType) Value() uint16 {
	return uint16(h)
}

type Packet interface {
	// String returns a string representation of the packet.
	String() string

	// Clone clones a packet.
	Clone() Packet

	// Header returns a pointer to the packet header.
	Header() *PacketHeader

	// Data returns the payload the packets holds. The packets stays the
	// owner of the data, i.e. modifying the returned data will also
	// modify the payload.
	Data() []byte

	// SetData replaces the payload of the packet with the provided one.
	SetData([]byte)

	// Len return the length of the payload in the packet.
	Len() uint64

	// Marshal writes the bytes representation of the packet to the provided writer.
	Marshal(w io.Writer) error

	// Unmarshal parses the given data into the packet header and its payload. Returns an error on failure.
	Unmarshal(data []byte) error

	// Dump returns the same as String with an additional hex-dump of the marshalled packet.
	Dump() string

	// MarshalCIF writes the byte representation of a control information field as payload
	// of the packet. Only for control packets.
	MarshalCIF(c CIF) error

	// UnmarshalCIF parses the payload into a control information field struct. Returns an error
	// on failure.
	UnmarshalCIF(c CIF) error

	// Decommission frees the payload. The packet shouldn't be uses afterwards.
	Decommission()
}

//  3. Packet Structure

type PacketHeader struct {
	Addr            net.Addr
	LocalAddr       net.Addr
	IsControlPacket bool
	PktTsbpdTime    uint64 // microseconds

	// control packet fields

	ControlType  CtrlType    // Control Packet Type.  The use of these bits is determined by the control packet type definition.
	SubType      CtrlSubType // This field specifies an additional subtype for specific packets.
	TypeSpecific uint32      // The use of this field depends on the particular control packet type. Handshake packets do not use this field.

	// data packet fields

	PacketSequenceNumber    circular.Number  // The sequential number of the data packet.
	PacketPositionFlag      PacketPosition   // This field indicates the position of the data packet in the message. The value "10b" (binary) means the first packet of the message. "00b" indicates a packet in the middle. "01b" designates the last packet. If a single data packet forms the whole message, the value is "11b".
	OrderFlag               bool             // Indicates whether the message should be delivered by the receiver in order (1) or not (0). Certain restrictions apply depending on the data transmission mode used (Section 4.2).
	KeyBaseEncryptionFlag   PacketEncryption // The flag bits indicate whether or not data is encrypted. The value "00b" (binary) means data is not encrypted. "01b" indicates that data is encrypted with an even key, and "10b" is used for odd key encryption. Refer to Section 6.  The value "11b" is only used in control packets.
	RetransmittedPacketFlag bool             // This flag is clear when a packet is transmitted the first time. The flag is set to "1" when a packet is retransmitted.
	MessageNumber           uint32           // The sequential number of consecutive data packets that form a message (see PP field).

	// common fields

	Timestamp           uint32 // microseconds
	DestinationSocketId uint32
}

type pkt struct {
	header PacketHeader

	payload *bytes.Buffer
}

type pool struct {
	pool sync.Pool
}

func newPool() *pool {
	return &pool{
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

func (p *pool) Get() *bytes.Buffer {
	b := p.pool.Get().(*bytes.Buffer)
	b.Reset()

	return b
}

func (p *pool) Put(b *bytes.Buffer) {
	p.pool.Put(b)
}

var payloadPool *pool = newPool()

func NewPacketFromData(addr net.Addr, rawdata []byte) (Packet, error) {
	p := NewPacket(addr)

	if len(rawdata) != 0 {
		if err := p.Unmarshal(rawdata); err != nil {
			p.Decommission()
			return nil, fmt.Errorf("invalid data: %w", err)
		}
	}

	return p, nil
}

func NewPacket(addr net.Addr) Packet {
	p := &pkt{
		header: PacketHeader{
			Addr:                  addr,
			PacketSequenceNumber:  circular.New(0, MAX_SEQUENCENUMBER),
			PacketPositionFlag:    SinglePacket,
			OrderFlag:             false,
			KeyBaseEncryptionFlag: UnencryptedPacket,
			MessageNumber:         1,
		},
		payload: payloadPool.Get(),
	}

	return p
}

func (p *pkt) Decommission() {
	if p.payload == nil {
		return
	}

	payloadPool.Put(p.payload)
	p.payload = nil
}

func (p pkt) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "timestamp=%#08x (%d), destId=%#08x\n", p.header.Timestamp, p.header.Timestamp, p.header.DestinationSocketId)

	if p.header.IsControlPacket {
		fmt.Fprintf(&b, "control packet:\n")
		fmt.Fprintf(&b, "   controlType=%#04x (%s)\n", p.header.ControlType.Value(), p.header.ControlType.String())
		fmt.Fprintf(&b, "   subType=%#04x (%s)\n", p.header.SubType.Value(), p.header.SubType.String())
		fmt.Fprintf(&b, "   typeSpecific=%#08x\n", p.header.TypeSpecific)
	} else {
		fmt.Fprintf(&b, "data packet:\n")
		fmt.Fprintf(&b, "   packetSequenceNumber=%#08x (%d)\n", p.header.PacketSequenceNumber.Val(), p.header.PacketSequenceNumber.Val())
		fmt.Fprintf(&b, "   packetPositionFlag=%s\n", p.header.PacketPositionFlag)
		fmt.Fprintf(&b, "   orderFlag=%v\n", p.header.OrderFlag)
		fmt.Fprintf(&b, "   keyBaseEncryptionFlag=%s\n", p.header.KeyBaseEncryptionFlag)
		fmt.Fprintf(&b, "   retransmittedPacketFlag=%v\n", p.header.RetransmittedPacketFlag)
		fmt.Fprintf(&b, "   messageNumber=%#08x (%d)\n", p.header.MessageNumber, p.header.MessageNumber)
	}

	fmt.Fprintf(&b, "data (%d bytes)", p.Len())

	return b.String()
}

func (p *pkt) Clone() Packet {
	clone := *p

	clone.payload = payloadPool.Get()
	clone.payload.Write(p.payload.Bytes())

	return &clone
}

func (p *pkt) Header() *PacketHeader {
	return &p.header
}

func (p *pkt) SetData(data []byte) {
	p.payload.Reset()
	p.payload.Write(data)
}

func (p *pkt) Data() []byte {
	return p.payload.Bytes()
}

func (p *pkt) Len() uint64 {
	return uint64(p.payload.Len())
}

func (p *pkt) Unmarshal(data []byte) error {
	if len(data) < 16 {
		return fmt.Errorf("data too short to unmarshal")
	}

	p.header.IsControlPacket = (data[0] & 0x80) != 0

	if p.header.IsControlPacket {
		p.header.ControlType = CtrlType(binary.BigEndian.Uint16(data[0:]) & ^uint16(1<<15)) // clear the first bit
		p.header.SubType = CtrlSubType(binary.BigEndian.Uint16(data[2:]))
		p.header.TypeSpecific = binary.BigEndian.Uint32(data[4:])
	} else {
		p.header.PacketSequenceNumber = circular.New(binary.BigEndian.Uint32(data[0:]), MAX_SEQUENCENUMBER)
		p.header.PacketPositionFlag = PacketPosition((data[4] & 0b11000000) >> 6)
		p.header.OrderFlag = (data[4] & 0b00100000) != 0
		p.header.KeyBaseEncryptionFlag = PacketEncryption((data[4] & 0b00011000) >> 3)
		p.header.RetransmittedPacketFlag = (data[4] & 0b00000100) != 0
		p.header.MessageNumber = binary.BigEndian.Uint32(data[4:]) & ^uint32(0b11111100<<24)
	}

	p.header.Timestamp = binary.BigEndian.Uint32(data[8:])
	p.header.DestinationSocketId = binary.BigEndian.Uint32(data[12:])

	p.payload.Reset()
	p.payload.Write(data[16:])

	return nil
}

func (p *pkt) Marshal(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("invalid writer")
	}

	var buffer [16]byte

	if p.payload == nil {
		return fmt.Errorf("invalid payload")
	}

	if p.header.IsControlPacket {
		binary.BigEndian.PutUint16(buffer[0:], p.header.ControlType.Value()) // control type
		binary.BigEndian.PutUint16(buffer[2:], p.header.SubType.Value())     // sub type
		binary.BigEndian.PutUint32(buffer[4:], p.header.TypeSpecific)        // type specific

		buffer[0] |= 0x80
	} else {
		binary.BigEndian.PutUint32(buffer[0:], p.header.PacketSequenceNumber.Val()) // sequence number

		var field uint32 = 0

		field |= ((p.header.PacketPositionFlag.Val() & 0b11) << 6) // 0b11000000
		if p.header.OrderFlag {
			field |= (1 << 5) // 0b11100000
		}
		field |= ((p.header.KeyBaseEncryptionFlag.Val() & 0b11) << 3) // 0b11111000
		if p.header.RetransmittedPacketFlag {
			field |= (1 << 2) // 0b11111100
		}
		field = field << 24 // 0b11111100_00000000_00000000_00000000
		field += (p.header.MessageNumber & 0b00000011_11111111_11111111_11111111)

		binary.BigEndian.PutUint32(buffer[4:], field) // sequence number
	}

	binary.BigEndian.PutUint32(buffer[8:], p.header.Timestamp)            // timestamp
	binary.BigEndian.PutUint32(buffer[12:], p.header.DestinationSocketId) // destination socket ID

	w.Write(buffer[0:])
	w.Write(p.payload.Bytes())

	return nil
}

func (p *pkt) Dump() string {
	var data bytes.Buffer
	p.Marshal(&data)

	return p.String() + "\n" + hex.Dump(data.Bytes())
}

func (p *pkt) MarshalCIF(c CIF) error {
	if !p.header.IsControlPacket {
		return fmt.Errorf("packet is not a control packet")
	}

	p.payload.Reset()
	return c.Marshal(p.payload)
}

func (p *pkt) UnmarshalCIF(c CIF) error {
	if !p.header.IsControlPacket {
		return nil
	}

	return c.Unmarshal(p.payload.Bytes())
}

// CIF reepresents a control information field
type CIF interface {
	// Marshal writes a byte representation of the CIF to the provided writer.
	Marshal(w io.Writer) error

	// Unmarshal parses the provided bytes into the CIF. Returns a non nil error of failure.
	Unmarshal(data []byte) error

	// String returns a string representation of the CIF.
	String() string
}

// 3.2.1.  Handshake

// CIFHandshake represents the SRT handshake messages.
type CIFHandshake struct {
	IsRequest bool

	Version                     uint32          // A base protocol version number. Currently used values are 4 and 5. Values greater than 5 are reserved for future use.
	EncryptionField             uint16          // Block cipher family and key size. The values of this field are described in Table 2. The default value is AES-128.
	ExtensionField              uint16          // This field is a message specific extension related to Handshake Type field. The value MUST be set to 0 except for the following cases. (1) If the handshake control packet is the INDUCTION message, this field is sent back by the Listener. (2) In the case of a CONCLUSION message, this field value should contain a combination of Extension Type values. For more details, see Section 4.3.1.
	InitialPacketSequenceNumber circular.Number // The sequence number of the very first data packet to be sent.
	MaxTransmissionUnitSize     uint32          // This value is typically set to 1500, which is the default Maximum Transmission Unit (MTU) size for Ethernet, but can be less.
	MaxFlowWindowSize           uint32          // The value of this field is the maximum number of data packets allowed to be "in flight" (i.e. the number of sent packets for which an ACK control packet has not yet been received).
	HandshakeType               HandshakeType   // This field indicates the handshake packet type. The possible values are described in Table 4. For more details refer to Section 4.3.
	SRTSocketId                 uint32          // This field holds the ID of the source SRT socket from which a handshake packet is issued.
	SynCookie                   uint32          // Randomized value for processing a handshake. The value of this field is specified by the handshake message type. See Section 4.3.
	PeerIP                      srtnet.IP       // IPv4 or IPv6 address of the packet's sender. The value consists of four 32-bit fields. In the case of IPv4 addresses, fields 2, 3 and 4 are filled with zeroes.

	HasHS            bool
	HasKM            bool
	HasSID           bool
	HasCongestionCtl bool

	// 3.2.1.1.  Handshake Extension Message
	SRTHS *CIFHandshakeExtension

	// 3.2.1.2.  Key Material Extension Message
	SRTKM *CIFKeyMaterialExtension

	// 3.2.1.3.  Stream ID Extension Message
	StreamId string

	// ??? Congestion Control Extension message (handshake.md #### Congestion controller)
	CongestionCtl string
}

func (c CIFHandshake) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "--- handshake ---\n")

	fmt.Fprintf(&b, "   version: %#08x\n", c.Version)
	fmt.Fprintf(&b, "   encryptionField: %#04x\n", c.EncryptionField)
	fmt.Fprintf(&b, "   extensionField: %#04x\n", c.ExtensionField)
	fmt.Fprintf(&b, "   initialPacketSequenceNumber: %#08x\n", c.InitialPacketSequenceNumber.Val())
	fmt.Fprintf(&b, "   maxTransmissionUnitSize: %#08x (%d)\n", c.MaxTransmissionUnitSize, c.MaxTransmissionUnitSize)
	fmt.Fprintf(&b, "   maxFlowWindowSize: %#08x (%d)\n", c.MaxFlowWindowSize, c.MaxFlowWindowSize)
	fmt.Fprintf(&b, "   handshakeType: %#08x (%s)\n", c.HandshakeType.Val(), c.HandshakeType.String())
	fmt.Fprintf(&b, "   srtSocketId: %#08x\n", c.SRTSocketId)
	fmt.Fprintf(&b, "   synCookie: %#08x\n", c.SynCookie)
	fmt.Fprintf(&b, "   peerIP: %s\n", c.PeerIP)

	if c.Version == 5 {
		if c.HasHS {
			fmt.Fprintf(&b, "%s\n", c.SRTHS.String())
		}

		if c.HasKM {
			fmt.Fprintf(&b, "%s\n", c.SRTKM.String())
		}

		if c.HasSID {
			fmt.Fprintf(&b, "--- SIDExt ---\n")
			fmt.Fprintf(&b, "   streamId : %s\n", c.StreamId)
			fmt.Fprintf(&b, "--- /SIDExt ---\n")
		}

		if c.HasCongestionCtl {
			fmt.Fprintf(&b, "--- CongestionExt ---\n")
			fmt.Fprintf(&b, "   congestion : %s\n", c.CongestionCtl)
			fmt.Fprintf(&b, "--- /CongestionExt ---\n")
		}
	}

	fmt.Fprintf(&b, "--- /handshake ---")

	return b.String()
}

func (c *CIFHandshake) Unmarshal(data []byte) error {
	if len(data) < 48 {
		return fmt.Errorf("data too short to unmarshal")
	}

	c.Version = binary.BigEndian.Uint32(data[0:])
	c.EncryptionField = binary.BigEndian.Uint16(data[4:])
	c.ExtensionField = binary.BigEndian.Uint16(data[6:])
	c.InitialPacketSequenceNumber = circular.New(binary.BigEndian.Uint32(data[8:])&MAX_SEQUENCENUMBER, MAX_SEQUENCENUMBER)
	c.MaxTransmissionUnitSize = binary.BigEndian.Uint32(data[12:])
	c.MaxFlowWindowSize = binary.BigEndian.Uint32(data[16:])
	c.HandshakeType = HandshakeType(binary.BigEndian.Uint32(data[20:]))
	c.SRTSocketId = binary.BigEndian.Uint32(data[24:])
	c.SynCookie = binary.BigEndian.Uint32(data[28:])
	c.PeerIP.Unmarshal(data[32:48])

	if c.HandshakeType == HSTYPE_INDUCTION {
		// Nothing more to unmarshal
		return nil
	}

	if c.HandshakeType != HSTYPE_CONCLUSION {
		// Everything else is currently not supported
		return nil
	}

	if c.ExtensionField == 0 {
		return nil
	}

	if len(data) <= 48 {
		// No extension data
		return nil
	}

	switch c.EncryptionField {
	case 0:
	case 2:
	case 3:
	case 4:
	default:
		return fmt.Errorf("invalid encryption field value (%d)", c.EncryptionField)
	}

	pivot := data[48:]

	for {
		extensionType := CtrlSubType(binary.BigEndian.Uint16(pivot[0:]))
		extensionLength := int(binary.BigEndian.Uint16(pivot[2:])) * 4

		pivot = pivot[4:]

		if extensionType == EXTTYPE_HSREQ || extensionType == EXTTYPE_HSRSP {
			// 3.2.1.1.  Handshake Extension Message
			if extensionLength != 12 || len(pivot) < extensionLength {
				return fmt.Errorf("invalid extension length of %d bytes (%s)", extensionLength, extensionType.String())
			}

			c.HasHS = true

			c.SRTHS = &CIFHandshakeExtension{}

			if err := c.SRTHS.Unmarshal(pivot); err != nil {
				return fmt.Errorf("CIFHandshakeExtension: %w", err)
			}
		} else if extensionType == EXTTYPE_KMREQ || extensionType == EXTTYPE_KMRSP {
			// 3.2.1.2.  Key Material Extension Message
			if len(pivot) < extensionLength {
				return fmt.Errorf("invalid extension length of %d bytes (%s)", extensionLength, extensionType.String())
			}

			c.HasKM = true

			c.SRTKM = &CIFKeyMaterialExtension{}

			if err := c.SRTKM.Unmarshal(pivot); err != nil {
				return fmt.Errorf("CIFKeyMaterialExtension: %w", err)
			}

			if c.EncryptionField == 0 {
				// using default cipher family and key size (AES-128)
				c.EncryptionField = 2
			}

			if c.EncryptionField == 2 && c.SRTKM.KLen != 16 {
				return fmt.Errorf("invalid key length for AES-128 (%d bit)", c.SRTKM.KLen*8)
			} else if c.EncryptionField == 3 && c.SRTKM.KLen != 24 {
				return fmt.Errorf("invalid key length for AES-192 (%d bit)", c.SRTKM.KLen*8)
			} else if c.EncryptionField == 4 && c.SRTKM.KLen != 32 {
				return fmt.Errorf("invalid key length for AES-256 (%d bit)", c.SRTKM.KLen*8)
			}
		} else if extensionType == EXTTYPE_SID {
			// 3.2.1.3.  Stream ID Extension Message
			if extensionLength > 512 || len(pivot) < extensionLength {
				return fmt.Errorf("invalid extension length of %d bytes (%s)", extensionLength, extensionType.String())
			}

			c.HasSID = true

			var b strings.Builder

			for i := 0; i < extensionLength; i += 4 {
				b.WriteByte(pivot[i+3])
				b.WriteByte(pivot[i+2])
				b.WriteByte(pivot[i+1])
				b.WriteByte(pivot[i+0])
			}

			c.StreamId = strings.TrimRight(b.String(), "\x00")
		} else if extensionType == EXTTYPE_CONGESTION {
			// ??? Congestion Control Extension message (handshake.md #### Congestion controller)
			if extensionLength > 4 || len(pivot) < extensionLength {
				return fmt.Errorf("invalid extension length of %d bytes (%s)", extensionLength, extensionType.String())
			}

			c.HasCongestionCtl = true

			var b strings.Builder

			for i := 0; i < extensionLength; i += 4 {
				b.WriteByte(pivot[i+3])
				b.WriteByte(pivot[i+2])
				b.WriteByte(pivot[i+1])
				b.WriteByte(pivot[i+0])
			}

			c.CongestionCtl = strings.TrimRight(b.String(), "\x00")
		} else if extensionType == EXTTYPE_FILTER || extensionType == EXTTYPE_GROUP {
			// Skip unimplemented extensions
			if len(pivot) < extensionLength {
				return fmt.Errorf("invalid extension length of %d bytes (%s)", extensionLength, extensionType.String())
			}
		} else {
			// Skip unknown extensions
			if len(pivot) < extensionLength {
				return fmt.Errorf("invalid extension length of %d bytes (%s)", extensionLength, extensionType.String())
			}
		}

		if len(pivot) > extensionLength {
			pivot = pivot[extensionLength:]
		} else {
			break
		}
	}

	return nil
}

func (c *CIFHandshake) Marshal(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("invalid writer")
	}

	var buffer [48]byte

	if len(c.StreamId) == 0 {
		c.HasSID = false
	}

	if c.Version == 5 {
		if c.HandshakeType == HSTYPE_CONCLUSION {
			c.ExtensionField = 0
		}

		if c.HasHS {
			c.ExtensionField = c.ExtensionField | 1
		}

		if c.HasKM {
			c.EncryptionField = c.SRTKM.KLen / 8
			c.ExtensionField = c.ExtensionField | 2
		}

		if c.HasSID {
			c.ExtensionField = c.ExtensionField | 4
		}

		if c.HasCongestionCtl {
			c.ExtensionField = c.ExtensionField | 4
		}
	} else {
		c.EncryptionField = 0
		c.ExtensionField = 2
	}

	binary.BigEndian.PutUint32(buffer[0:], c.Version)                           // version
	binary.BigEndian.PutUint16(buffer[4:], c.EncryptionField)                   // encryption field
	binary.BigEndian.PutUint16(buffer[6:], c.ExtensionField)                    // extension field
	binary.BigEndian.PutUint32(buffer[8:], c.InitialPacketSequenceNumber.Val()) // initialPacketSequenceNumber
	binary.BigEndian.PutUint32(buffer[12:], c.MaxTransmissionUnitSize)          // maxTransmissionUnitSize
	binary.BigEndian.PutUint32(buffer[16:], c.MaxFlowWindowSize)                // maxFlowWindowSize
	binary.BigEndian.PutUint32(buffer[20:], c.HandshakeType.Val())              // handshakeType
	binary.BigEndian.PutUint32(buffer[24:], c.SRTSocketId)                      // Socket ID of the Listener, should be some own generated ID
	binary.BigEndian.PutUint32(buffer[28:], c.SynCookie)                        // SYN cookie
	c.PeerIP.Marshal(buffer[32:])                                               // peerIP

	w.Write(buffer[:48])

	if c.HasHS {
		var data bytes.Buffer

		c.SRTHS.Marshal(&data)

		if c.IsRequest {
			binary.BigEndian.PutUint16(buffer[0:], EXTTYPE_HSREQ.Value())
		} else {
			binary.BigEndian.PutUint16(buffer[0:], EXTTYPE_HSRSP.Value())
		}

		binary.BigEndian.PutUint16(buffer[2:], 3)

		w.Write(buffer[:4])
		w.Write(data.Bytes())
	}

	if c.HasKM {
		var data bytes.Buffer

		c.SRTKM.Marshal(&data)

		if c.IsRequest {
			binary.BigEndian.PutUint16(buffer[0:], EXTTYPE_KMREQ.Value())
		} else {
			binary.BigEndian.PutUint16(buffer[0:], EXTTYPE_KMRSP.Value())
		}

		binary.BigEndian.PutUint16(buffer[2:], uint16(data.Len()/4))

		w.Write(buffer[:4])
		w.Write(data.Bytes())
	}

	if c.HasSID {
		streamId := bytes.NewBufferString(c.StreamId)

		missing := (4 - streamId.Len()%4)
		if missing < 4 {
			for range missing {
				streamId.WriteByte(0)
			}
		}

		binary.BigEndian.PutUint16(buffer[0:], EXTTYPE_SID.Value())
		binary.BigEndian.PutUint16(buffer[2:], uint16(streamId.Len()/4))

		w.Write(buffer[:4])

		b := streamId.Bytes()

		for i := 0; i < len(b); i += 4 {
			buffer[0] = b[i+3]
			buffer[1] = b[i+2]
			buffer[2] = b[i+1]
			buffer[3] = b[i+0]

			w.Write(buffer[:4])
		}
	}

	if c.HasCongestionCtl && c.CongestionCtl != "live" {
		congestion := bytes.NewBufferString(c.CongestionCtl)

		missing := (4 - congestion.Len()%4)
		if missing < 4 {
			for range missing {
				congestion.WriteByte(0)
			}
		}

		binary.BigEndian.PutUint16(buffer[0:], EXTTYPE_CONGESTION.Value())
		binary.BigEndian.PutUint16(buffer[2:], uint16(congestion.Len()/4))

		w.Write(buffer[:4])

		b := congestion.Bytes()

		for i := 0; i < len(b); i += 4 {
			buffer[0] = b[i+3]
			buffer[1] = b[i+2]
			buffer[2] = b[i+1]
			buffer[3] = b[i+0]

			w.Write(buffer[:4])
		}
	}

	return nil
}

// 3.2.1.1.1.  Handshake Extension Message Flags

// CIFHandshakeExtensionFlags represents the Handshake Extension Message Flags
type CIFHandshakeExtensionFlags struct {
	TSBPDSND      bool // Defines if the TSBPD mechanism (Section 4.5) will be used for sending.
	TSBPDRCV      bool // Defines if the TSBPD mechanism (Section 4.5) will be used for receiving.
	CRYPT         bool // MUST be set. It is a legacy flag that indicates the party understands KK field of the SRT Packet (Figure 3).
	TLPKTDROP     bool // Should be set if too-late packet drop mechanism will be used during transmission.  See Section 4.6.
	PERIODICNAK   bool // Indicates the peer will send periodic NAK packets. See Section 4.8.2.
	REXMITFLG     bool // MUST be set. It is a legacy flag that indicates the peer understands the R field of the SRT DATA Packet
	STREAM        bool // Identifies the transmission mode (Section 4.2) to be used in the connection. If the flag is set, the buffer mode (Section 4.2.2) is used. Otherwise, the message mode (Section 4.2.1) is used.
	PACKET_FILTER bool // Indicates if the peer supports packet filter.
}

// 3.2.1.1.  Handshake Extension Message

// CIFHandshakeExtension represents the Handshake Extension Message
type CIFHandshakeExtension struct {
	SRTVersion     uint32
	SRTFlags       CIFHandshakeExtensionFlags
	RecvTSBPDDelay uint16 // milliseconds, see "4.4.  SRT Buffer Latency"
	SendTSBPDDelay uint16 // milliseconds, see "4.4.  SRT Buffer Latency"
}

func (c CIFHandshakeExtension) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "--- HSExt ---\n")

	fmt.Fprintf(&b, "   srtVersion: %#08x\n", c.SRTVersion)
	fmt.Fprintf(&b, "   srtFlags:\n")
	fmt.Fprintf(&b, "      TSBPDSND     : %v\n", c.SRTFlags.TSBPDSND)
	fmt.Fprintf(&b, "      TSBPDRCV     : %v\n", c.SRTFlags.TSBPDRCV)
	fmt.Fprintf(&b, "      CRYPT        : %v\n", c.SRTFlags.CRYPT)
	fmt.Fprintf(&b, "      TLPKTDROP    : %v\n", c.SRTFlags.TLPKTDROP)
	fmt.Fprintf(&b, "      PERIODICNAK  : %v\n", c.SRTFlags.PERIODICNAK)
	fmt.Fprintf(&b, "      REXMITFLG    : %v\n", c.SRTFlags.REXMITFLG)
	fmt.Fprintf(&b, "      STREAM       : %v\n", c.SRTFlags.STREAM)
	fmt.Fprintf(&b, "      PACKET_FILTER: %v\n", c.SRTFlags.PACKET_FILTER)
	fmt.Fprintf(&b, "   recvTSBPDDelay: %#04x (%dms)\n", c.RecvTSBPDDelay, c.RecvTSBPDDelay)
	fmt.Fprintf(&b, "   sendTSBPDDelay: %#04x (%dms)\n", c.SendTSBPDDelay, c.SendTSBPDDelay)

	fmt.Fprintf(&b, "--- /HSExt ---")

	return b.String()
}

func (c *CIFHandshakeExtension) Unmarshal(data []byte) error {
	if len(data) < 12 {
		return fmt.Errorf("data too short to unmarshal")
	}

	c.SRTVersion = binary.BigEndian.Uint32(data[0:])
	srtFlags := binary.BigEndian.Uint32(data[4:])

	c.SRTFlags.TSBPDSND = (srtFlags&SRTFLAG_TSBPDSND != 0)
	c.SRTFlags.TSBPDRCV = (srtFlags&SRTFLAG_TSBPDRCV != 0)
	c.SRTFlags.CRYPT = (srtFlags&SRTFLAG_CRYPT != 0)
	c.SRTFlags.TLPKTDROP = (srtFlags&SRTFLAG_TLPKTDROP != 0)
	c.SRTFlags.PERIODICNAK = (srtFlags&SRTFLAG_PERIODICNAK != 0)
	c.SRTFlags.REXMITFLG = (srtFlags&SRTFLAG_REXMITFLG != 0)
	c.SRTFlags.STREAM = (srtFlags&SRTFLAG_STREAM != 0)
	c.SRTFlags.PACKET_FILTER = (srtFlags&SRTFLAG_PACKET_FILTER != 0)

	c.RecvTSBPDDelay = binary.BigEndian.Uint16(data[8:])
	c.SendTSBPDDelay = binary.BigEndian.Uint16(data[10:])

	return nil
}

func (c *CIFHandshakeExtension) Marshal(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("invalid writer")
	}

	var buffer [12]byte

	binary.BigEndian.PutUint32(buffer[0:], c.SRTVersion)
	var srtFlags uint32 = 0

	if c.SRTFlags.TSBPDSND {
		srtFlags |= SRTFLAG_TSBPDSND
	}

	if c.SRTFlags.TSBPDRCV {
		srtFlags |= SRTFLAG_TSBPDRCV
	}

	if c.SRTFlags.CRYPT {
		srtFlags |= SRTFLAG_CRYPT
	}

	if c.SRTFlags.TLPKTDROP {
		srtFlags |= SRTFLAG_TLPKTDROP
	}

	if c.SRTFlags.PERIODICNAK {
		srtFlags |= SRTFLAG_PERIODICNAK
	}

	if c.SRTFlags.REXMITFLG {
		srtFlags |= SRTFLAG_REXMITFLG
	}

	if c.SRTFlags.STREAM {
		srtFlags |= SRTFLAG_STREAM
	}

	if c.SRTFlags.PACKET_FILTER {
		srtFlags |= SRTFLAG_PACKET_FILTER
	}

	binary.BigEndian.PutUint32(buffer[4:], srtFlags)
	binary.BigEndian.PutUint16(buffer[8:], c.RecvTSBPDDelay)
	binary.BigEndian.PutUint16(buffer[10:], c.SendTSBPDDelay)

	_, err := w.Write(buffer[:12])

	return err
}

// 3.2.2.  Key Material

const (
	KM_NOSECRET  uint32 = 3
	KM_BADSECRET uint32 = 4
)

// CIFKeyMaterialExtension represents the Key Material message. It is used as part of
// the v5 handshake or on its own after a v4 handshake.
type CIFKeyMaterialExtension struct {
	Error                 uint32
	S                     uint8            // This is a fixed-width field that is reserved for future usage. value = {0}
	Version               uint8            // This is a fixed-width field that indicates the SRT version. value = {1}
	PacketType            uint8            // This is a fixed-width field that indicates the Packet Type: 0: Reserved, 1: Media Stream Message (MSmsg), 2: Keying Material Message (KMmsg), 7: Reserved to discriminate MPEG-TS packet (0x47=sync byte). value = {2}
	Sign                  uint16           // This is a fixed-width field that contains the signature 'HAI' encoded as a PnP Vendor ID [PNPID] (in big-endian order). value = {0x2029}
	Resv1                 uint8            // This is a fixed-width field reserved for flag extension or other usage. value = {0}
	KeyBasedEncryption    PacketEncryption // This is a fixed-width field that indicates which SEKs (odd and/or even) are provided in the extension: 00b: No SEK is provided (invalid extension format); 01b: Even key is provided; 10b: Odd key is provided; 11b: Both even and odd keys are provided.
	KeyEncryptionKeyIndex uint32           // This is a fixed-width field for specifying the KEK index (big-endian order) was used to wrap (and optionally authenticate) the SEK(s). The value 0 is used to indicate the default key of the current stream. Other values are reserved for the possible use of a key management system in the future to retrieve a cryptographic context. 0: Default stream associated key (stream/system default); 1..255: Reserved for manually indexed keys. value = {0}
	Cipher                uint8            // This is a fixed-width field for specifying encryption cipher and mode: 0: None or KEKI indexed crypto context; 2: AES-CTR [SP800-38A].
	Authentication        uint8            // This is a fixed-width field for specifying a message authentication code algorithm: 0: None or KEKI indexed crypto context.
	StreamEncapsulation   uint8            // This is a fixed-width field for describing the stream encapsulation: 0: Unspecified or KEKI indexed crypto context; 1: MPEG-TS/UDP; 2: MPEG-TS/SRT. value = {2}
	Resv2                 uint8            // This is a fixed-width field reserved for future use. value = {0}
	Resv3                 uint16           // This is a fixed-width field reserved for future use. value = {0}
	SLen                  uint16           // This is a fixed-width field for specifying salt length SLen in bytes divided by 4. Can be zero if no salt/IV present. The only valid length of salt defined is 128 bits.
	KLen                  uint16           // This is a fixed-width field for specifying SEK length in bytes divided by 4. Size of one key even if two keys present. MUST match the key size specified in the Encryption Field of the handshake packet Table 2.
	Salt                  []byte           // This is a variable-width field that complements the keying material by specifying a salt key.
	Wrap                  []byte           // (64 + n * KLen * 8) bits. This is a variable- width field for specifying Wrapped key(s), where n = (KK + 1)/2 and the size of the wrap field is ((n * KLen) + 8) bytes.
}

func (c CIFKeyMaterialExtension) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "--- KMExt ---\n")

	fmt.Fprintf(&b, "   s: %d\n", c.S)
	fmt.Fprintf(&b, "   version: %d\n", c.Version)
	fmt.Fprintf(&b, "   packetType: %d\n", c.PacketType)
	fmt.Fprintf(&b, "   sign: %#08x\n", c.Sign)
	fmt.Fprintf(&b, "   resv1: %d\n", c.Resv1)
	fmt.Fprintf(&b, "   keyBasedEncryption: %s\n", c.KeyBasedEncryption.String())
	fmt.Fprintf(&b, "   keyEncryptionKeyIndex: %d\n", c.KeyEncryptionKeyIndex)
	fmt.Fprintf(&b, "   cipher: %d\n", c.Cipher)
	fmt.Fprintf(&b, "   authentication: %d\n", c.Authentication)
	fmt.Fprintf(&b, "   streamEncapsulation: %d\n", c.StreamEncapsulation)
	fmt.Fprintf(&b, "   resv2: %d\n", c.Resv2)
	fmt.Fprintf(&b, "   resv3: %d\n", c.Resv3)
	fmt.Fprintf(&b, "   sLen: %d (%d)\n", c.SLen, c.SLen/4)
	fmt.Fprintf(&b, "   kLen: %d (%d)\n", c.KLen, c.KLen/4)
	fmt.Fprintf(&b, "   salt: %#08x\n", c.Salt)
	fmt.Fprintf(&b, "   wrap: %#08x\n", c.Wrap)

	fmt.Fprintf(&b, "--- /KMExt ---")

	return b.String()
}

func (c *CIFKeyMaterialExtension) Unmarshal(data []byte) error {
	if len(data) == 4 {
		// This is an error response
		c.Error = binary.LittleEndian.Uint32(data[0:])
		if c.Error != KM_NOSECRET && c.Error != KM_BADSECRET {
			return fmt.Errorf("invalid error (%d)", c.Error)
		}
		return nil
	} else if len(data) < 16 {
		return fmt.Errorf("data too short to unmarshal")
	}

	c.S = uint8(data[0] & 0b1000_0000 >> 7)
	if c.S != 0 {
		return fmt.Errorf("invalid value for S")
	}

	c.Version = uint8(data[0] & 0b0111_0000 >> 4)
	if c.Version != 1 {
		return fmt.Errorf("invalid version")
	}

	c.PacketType = uint8(data[0] & 0b0000_1111)
	if c.PacketType != 2 {
		return fmt.Errorf("invalid packet type (%d)", c.PacketType)
	}

	c.Sign = binary.BigEndian.Uint16(data[1:])
	if c.Sign != 0x2029 {
		return fmt.Errorf("invalid signature (%#08x)", c.Sign)
	}

	c.Resv1 = uint8(data[3] & 0b1111_1100 >> 2)
	c.KeyBasedEncryption = PacketEncryption(data[3] & 0b0000_0011)
	if !c.KeyBasedEncryption.IsValid() || c.KeyBasedEncryption == UnencryptedPacket {
		return fmt.Errorf("invalid extension format (KK must not be 0)")
	}

	c.KeyEncryptionKeyIndex = binary.BigEndian.Uint32(data[4:])
	if c.KeyEncryptionKeyIndex != 0 {
		return fmt.Errorf("invalid key encryption key index (%d)", c.KeyEncryptionKeyIndex)
	}

	c.Cipher = uint8(data[8])
	c.Authentication = uint8(data[9])
	c.StreamEncapsulation = uint8(data[10])
	if c.StreamEncapsulation != 2 {
		return fmt.Errorf("invalid stream encapsulation (%d)", c.StreamEncapsulation)
	}

	c.Resv2 = uint8(data[11])
	c.Resv3 = binary.BigEndian.Uint16(data[12:])
	c.SLen = uint16(data[14]) * 4
	c.KLen = uint16(data[15]) * 4

	switch c.KLen {
	case 16:
	case 24:
	case 32:
	default:
		return fmt.Errorf("invalid key length")
	}

	offset := 16

	if c.SLen != 0 {
		if c.SLen != 16 {
			return fmt.Errorf("invalid salt length")
		}

		if len(data[offset:]) < 16 {
			return fmt.Errorf("data too short to unmarshal")
		}

		c.Salt = make([]byte, 16)
		copy(c.Salt, data[offset:])

		offset += 16
	}

	n := 1
	if c.KeyBasedEncryption == EvenAndOddKey {
		n = 2
	}

	if len(data[offset:]) < n*int(c.KLen)+8 {
		return fmt.Errorf("data too short to unmarshal")
	}

	c.Wrap = make([]byte, n*int(c.KLen)+8)
	copy(c.Wrap, data[offset:])

	return nil
}

func (c *CIFKeyMaterialExtension) Marshal(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("invalid writer")
	}

	var buffer [128]byte

	b := byte(0)

	b |= (c.S << 7) & 0b1000_0000
	b |= (c.Version << 4) & 0b0111_0000
	b |= c.PacketType & 0b0000_1111

	buffer[0] = b
	binary.BigEndian.PutUint16(buffer[1:], c.Sign)

	b = 0
	b |= (c.Resv1 << 2) & 0b1111_1100
	b |= uint8(c.KeyBasedEncryption) & 0b0000_0011

	buffer[3] = b
	binary.BigEndian.PutUint32(buffer[4:], c.KeyEncryptionKeyIndex)

	buffer[8] = byte(c.Cipher)
	buffer[9] = byte(c.Authentication)
	buffer[10] = byte(c.StreamEncapsulation)
	buffer[11] = byte(c.Resv2)

	binary.BigEndian.PutUint16(buffer[12:], c.Resv3)

	buffer[14] = byte(c.SLen / 4)
	buffer[15] = byte(c.KLen / 4)

	offset := 16

	if c.SLen != 0 {
		copy(buffer[offset:], c.Salt[0:])
		offset += len(c.Salt)
	}

	copy(buffer[offset:], c.Wrap)
	offset += len(c.Wrap)

	_, err := w.Write(buffer[:offset])

	return err
}

// 3.2.4.  ACK (Acknowledgment)

// CIFACK represents an ACK message.
type CIFACK struct {
	IsLite                      bool
	IsSmall                     bool
	LastACKPacketSequenceNumber circular.Number
	RTT                         uint32 // microseconds
	RTTVar                      uint32 // microseconds
	AvailableBufferSize         uint32 // bytes
	PacketsReceivingRate        uint32 // packets/s
	EstimatedLinkCapacity       uint32
	ReceivingRate               uint32 // bytes/s
}

func (c CIFACK) String() string {
	var b strings.Builder

	ackType := "full"
	if c.IsLite {
		ackType = "lite"
	} else if c.IsSmall {
		ackType = "small"
	}

	fmt.Fprintf(&b, "--- ACK (type: %s) ---\n", ackType)

	fmt.Fprintf(&b, "   lastACKPacketSequenceNumber: %#08x (%d)\n", c.LastACKPacketSequenceNumber.Val(), c.LastACKPacketSequenceNumber.Val())

	if !c.IsLite {
		fmt.Fprintf(&b, "   rtt: %#08x (%dus)\n", c.RTT, c.RTT)
		fmt.Fprintf(&b, "   rttVar: %#08x (%dus)\n", c.RTTVar, c.RTTVar)
		fmt.Fprintf(&b, "   availableBufferSize: %#08x\n", c.AvailableBufferSize)
		fmt.Fprintf(&b, "   packetsReceivingRate: %#08x\n", c.PacketsReceivingRate)
		fmt.Fprintf(&b, "   estimatedLinkCapacity: %#08x\n", c.EstimatedLinkCapacity)
		fmt.Fprintf(&b, "   receivingRate: %#08x\n", c.ReceivingRate)
	}

	fmt.Fprintf(&b, "--- /ACK ---")

	return b.String()
}

func (c *CIFACK) Unmarshal(data []byte) error {
	c.IsLite = false
	c.IsSmall = false

	if len(data) == 4 {
		c.IsLite = true

		c.LastACKPacketSequenceNumber = circular.New(binary.BigEndian.Uint32(data[0:])&MAX_SEQUENCENUMBER, MAX_SEQUENCENUMBER)

		return nil
	} else if len(data) == 16 {
		c.IsSmall = true

		c.LastACKPacketSequenceNumber = circular.New(binary.BigEndian.Uint32(data[0:])&MAX_SEQUENCENUMBER, MAX_SEQUENCENUMBER)
		c.RTT = binary.BigEndian.Uint32(data[4:])
		c.RTTVar = binary.BigEndian.Uint32(data[8:])
		c.AvailableBufferSize = binary.BigEndian.Uint32(data[12:])

		return nil
	}

	if len(data) < 28 {
		return fmt.Errorf("data too short to unmarshal")
	}

	c.LastACKPacketSequenceNumber = circular.New(binary.BigEndian.Uint32(data[0:])&MAX_SEQUENCENUMBER, MAX_SEQUENCENUMBER)
	c.RTT = binary.BigEndian.Uint32(data[4:])
	c.RTTVar = binary.BigEndian.Uint32(data[8:])
	c.AvailableBufferSize = binary.BigEndian.Uint32(data[12:])
	c.PacketsReceivingRate = binary.BigEndian.Uint32(data[16:])
	c.EstimatedLinkCapacity = binary.BigEndian.Uint32(data[20:])
	c.ReceivingRate = binary.BigEndian.Uint32(data[24:])

	return nil
}

func (c *CIFACK) Marshal(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("invalid writer")
	}

	var buffer [28]byte

	binary.BigEndian.PutUint32(buffer[0:], c.LastACKPacketSequenceNumber.Val())
	binary.BigEndian.PutUint32(buffer[4:], c.RTT)
	binary.BigEndian.PutUint32(buffer[8:], c.RTTVar)
	binary.BigEndian.PutUint32(buffer[12:], c.AvailableBufferSize)
	binary.BigEndian.PutUint32(buffer[16:], c.PacketsReceivingRate)
	binary.BigEndian.PutUint32(buffer[20:], c.EstimatedLinkCapacity)
	binary.BigEndian.PutUint32(buffer[24:], c.ReceivingRate)

	if c.IsLite {
		w.Write(buffer[0:4])
	} else if c.IsSmall {
		w.Write(buffer[0:16])
	} else {
		w.Write(buffer[0:])
	}

	return nil
}

// 3.2.5.  NAK (Loss Report)

// CIFNAK represents a NAK message
type CIFNAK struct {
	LostPacketSequenceNumber []circular.Number
}

func (c CIFNAK) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "--- NAK ---\n")

	if len(c.LostPacketSequenceNumber)%2 != 0 {
		fmt.Fprintf(&b, "   invalid list of sequence numbers\n")
		return b.String()
	}

	for i := 0; i < len(c.LostPacketSequenceNumber); i += 2 {
		if c.LostPacketSequenceNumber[i].Equals(c.LostPacketSequenceNumber[i+1]) {
			fmt.Fprintf(&b, "   single: %#08x\n", c.LostPacketSequenceNumber[i].Val())
		} else {
			fmt.Fprintf(&b, "      row: %#08x to %#08x\n", c.LostPacketSequenceNumber[i].Val(), c.LostPacketSequenceNumber[i+1].Val())
		}
	}

	fmt.Fprintf(&b, "--- /NAK ---")

	return b.String()
}

func (c *CIFNAK) Unmarshal(data []byte) error {
	if len(data)%4 != 0 {
		return fmt.Errorf("data has wrong length to unmarshal")
	}

	// Appendix A

	c.LostPacketSequenceNumber = []circular.Number{}

	var sequenceNumber circular.Number
	isRange := false

	for i := 0; i < len(data); i += 4 {
		sequenceNumber = circular.New(binary.BigEndian.Uint32(data[i:])&MAX_SEQUENCENUMBER, MAX_SEQUENCENUMBER)

		if data[i]&0b10000000 == 0 {
			c.LostPacketSequenceNumber = append(c.LostPacketSequenceNumber, sequenceNumber)

			if !isRange {
				c.LostPacketSequenceNumber = append(c.LostPacketSequenceNumber, sequenceNumber)
			}

			isRange = false
		} else {
			c.LostPacketSequenceNumber = append(c.LostPacketSequenceNumber, sequenceNumber)
			isRange = true
		}
	}

	if len(c.LostPacketSequenceNumber)%2 != 0 {
		return fmt.Errorf("data too short to unmarshal")
	}

	sort.Slice(c.LostPacketSequenceNumber, func(i, j int) bool { return c.LostPacketSequenceNumber[i].Lt(c.LostPacketSequenceNumber[j]) })

	return nil
}

func (c *CIFNAK) Marshal(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("invalid writer")
	}

	if len(c.LostPacketSequenceNumber)%2 != 0 {
		return fmt.Errorf("invalid length of lost packet sequence numbers")
	}

	// Appendix A

	var buffer [8]byte
	bytesWritten := 0

	for i := 0; i < len(c.LostPacketSequenceNumber); i += 2 {
		if c.LostPacketSequenceNumber[i] == c.LostPacketSequenceNumber[i+1] {
			binary.BigEndian.PutUint32(buffer[0:], c.LostPacketSequenceNumber[i].Val())
			w.Write(buffer[0:4])

			bytesWritten += 4
		} else {
			binary.BigEndian.PutUint32(buffer[0:], c.LostPacketSequenceNumber[i].Val()|0b10000000_00000000_00000000_00000000)
			binary.BigEndian.PutUint32(buffer[4:], c.LostPacketSequenceNumber[i+1].Val())
			w.Write(buffer[0:])

			bytesWritten += 8
		}

		if bytesWritten >= MAX_PAYLOAD_SIZE-4 {
			break
		}
	}

	return nil
}

//  3.2.7. Shutdown

// CIFShutdown represents a shutdown message.
type CIFShutdown struct{}

func (c CIFShutdown) String() string {
	return "--- Shutdown ---"
}

func (c *CIFShutdown) Unmarshal(data []byte) error {
	if len(data) != 0 && len(data) != 4 {
		return fmt.Errorf("invalid length")
	}

	return nil
}

func (c *CIFShutdown) Marshal(w io.Writer) error {
	if w == nil {
		return fmt.Errorf("invalid writer")
	}

	var buffer [4]byte

	binary.BigEndian.PutUint32(buffer[0:], 0)

	_, err := w.Write(buffer[0:])

	return err
}

//  3.1. Data Packets

type PacketPosition uint

const (
	FirstPacket  PacketPosition = 2
	MiddlePacket PacketPosition = 0
	LastPacket   PacketPosition = 1
	SinglePacket PacketPosition = 3
)

func (p PacketPosition) String() string {
	switch uint(p) {
	case 0:
		return "middle"
	case 1:
		return "last"
	case 2:
		return "first"
	case 3:
		return "single"
	}

	return `¯\_(ツ)_/¯`
}

func (p PacketPosition) IsValid() bool {
	return p < 4
}

func (p PacketPosition) Val() uint32 {
	return uint32(p)
}

//  3.1. Data Packets

type PacketEncryption uint

const (
	UnencryptedPacket PacketEncryption = 0
	EvenKeyEncrypted  PacketEncryption = 1
	OddKeyEncrypted   PacketEncryption = 2
	EvenAndOddKey     PacketEncryption = 3
)

func (p PacketEncryption) String() string {
	switch uint(p) {
	case 0:
		return "unencrypted"
	case 1:
		return "even key"
	case 2:
		return "odd key"
	case 3:
		return "even and odd key"
	}

	return `¯\_(ツ)_/¯`
}

func (p PacketEncryption) IsValid() bool {
	return p < 4
}

func (p PacketEncryption) Opposite() PacketEncryption {
	if p == EvenKeyEncrypted {
		return OddKeyEncrypted
	}

	if p == OddKeyEncrypted {
		return EvenKeyEncrypted
	}

	return p
}

func (p PacketEncryption) Val() uint32 {
	return uint32(p)
}
````

## File: pubsub_test.go
````go
package srt

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	message := "Hello World!"
	channel := NewPubSub(PubSubConfig{})

	config := DefaultConfig()

	server := Server{
		Addr:   "127.0.0.1:6003",
		Config: &config,
		HandleConnect: func(req ConnRequest) ConnType {
			streamid := req.StreamId()

			if streamid == "publish" {
				return PUBLISH
			} else if streamid == "subscribe" {
				return SUBSCRIBE
			}

			return REJECT
		},
		HandlePublish: func(conn Conn) {
			channel.Publish(conn)

			conn.Close()
		},
		HandleSubscribe: func(conn Conn) {
			channel.Subscribe(conn)

			conn.Close()
		},
	}

	err := server.Listen()
	require.NoError(t, err)

	go func() {
		err := server.Serve()
		if err == ErrServerClosed {
			return
		}
		require.NoError(t, err)
	}()

	readerReadyWg := sync.WaitGroup{}
	readerReadyWg.Add(2)

	readerDoneWg := sync.WaitGroup{}
	readerDoneWg.Add(2)

	dataReader1 := bytes.Buffer{}
	dataReader2 := bytes.Buffer{}

	go func() {
		config := DefaultConfig()
		config.StreamId = "subscribe"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		buffer := make([]byte, 2048)

		readerReadyWg.Done()

		for {
			n, err := conn.Read(buffer)
			if n != 0 {
				dataReader1.Write(buffer[:n])
			}

			if err != nil {
				break
			}
		}

		err = conn.Close()
		require.NoError(t, err)

		readerDoneWg.Done()
	}()

	go func() {
		config := DefaultConfig()
		config.StreamId = "subscribe"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		buffer := make([]byte, 2048)

		readerReadyWg.Done()

		for {
			n, err := conn.Read(buffer)
			if n != 0 {
				dataReader2.Write(buffer[:n])
			}

			if err != nil {
				break
			}
		}

		err = conn.Close()
		require.NoError(t, err)

		readerDoneWg.Done()
	}()

	readerReadyWg.Wait()

	writerWg := sync.WaitGroup{}

	writerWg.Go(func() {
		config := DefaultConfig()
		config.StreamId = "publish"

		conn, err := Dial("srt", "127.0.0.1:6003", config)
		if !assert.NoError(t, err) {
			panic(err.Error())
		}

		n, err := conn.Write([]byte(message))
		require.NoError(t, err)
		require.Equal(t, 12, n)

		time.Sleep(3 * time.Second)

		err = conn.Close()
		require.NoError(t, err)
	})

	writerWg.Wait()
	readerDoneWg.Wait()

	server.Shutdown()

	reader1 := dataReader1.String()
	reader2 := dataReader2.String()

	require.Equal(t, message, reader1)
	require.Equal(t, message, reader2)
}
````

## File: pubsub.go
````go
package srt

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/datarhei/gosrt/packet"
)

// PubSub is a publish/subscriber service for SRT connections.
type PubSub interface {
	// Publish accepts a SRT connection where it reads from. It blocks
	// until the connection closes. The returned error indicates why it
	// stopped. There can be only one publisher.
	Publish(c Conn) error

	// Subscribe accepts a SRT connection where it writes the data from
	// the publisher to. It blocks until an error happens. If the publisher
	// disconnects, io.EOF is returned. There can be an arbitrary number
	// of subscribers.
	Subscribe(c Conn) error
}

// pubSub is an implementation of the PubSub interface
type pubSub struct {
	incoming      chan packet.Packet
	ctx           context.Context
	cancel        context.CancelFunc
	publish       bool
	publishLock   sync.Mutex
	listeners     map[uint32]chan packet.Packet
	listenersLock sync.Mutex
	logger        Logger
}

// PubSubConfig is for configuring a new PubSub
type PubSubConfig struct {
	Logger Logger // Optional logger
}

// NewPubSub returns a PubSub. After the publishing connection closed
// this PubSub can't be used anymore.
func NewPubSub(config PubSubConfig) PubSub {
	pb := &pubSub{
		incoming:  make(chan packet.Packet, 8192),
		listeners: make(map[uint32]chan packet.Packet),
		logger:    config.Logger,
	}

	pb.ctx, pb.cancel = context.WithCancel(context.Background())

	if pb.logger == nil {
		pb.logger = NewLogger(nil)
	}

	go pb.broadcast()

	return pb
}

func (pb *pubSub) broadcast() {
	defer func() {
		pb.logger.Print("pubsub:close", 0, 1, func() string { return "exiting broadcast loop" })
	}()

	pb.logger.Print("pubsub:new", 0, 1, func() string { return "starting broadcast loop" })

	for {
		select {
		case <-pb.ctx.Done():
			return
		case p := <-pb.incoming:
			pb.listenersLock.Lock()
			for socketId, c := range pb.listeners {
				pp := p.Clone()

				select {
				case c <- pp:
				default:
					pb.logger.Print("pubsub:error", socketId, 1, func() string { return "broadcast target queue is full" })
				}
			}
			pb.listenersLock.Unlock()

			// We don't need this packet anymore
			p.Decommission()
		}
	}
}

func (pb *pubSub) Publish(c Conn) error {
	pb.publishLock.Lock()
	defer pb.publishLock.Unlock()

	if pb.publish {
		err := fmt.Errorf("only one publisher is allowed")
		pb.logger.Print("pubsub:error", 0, 1, func() string { return err.Error() })
		return err
	}

	var p packet.Packet
	var err error

	socketId := c.SocketId()

	pb.logger.Print("pubsub:publish", socketId, 1, func() string { return "new publisher" })

	pb.publish = true

	for {
		p, err = c.ReadPacket()
		if err != nil {
			pb.logger.Print("pubsub:error", socketId, 1, func() string { return err.Error() })
			break
		}

		select {
		case pb.incoming <- p:
		default:
			pb.logger.Print("pubsub:error", socketId, 1, func() string { return "incoming queue is full" })
		}
	}

	pb.cancel()

	return err
}

func (pb *pubSub) Subscribe(c Conn) error {
	l := make(chan packet.Packet, 8192)
	socketId := c.SocketId()

	pb.logger.Print("pubsub:subscribe", socketId, 1, func() string { return "new subscriber" })

	pb.listenersLock.Lock()
	pb.listeners[socketId] = l
	pb.listenersLock.Unlock()

	defer func() {
		pb.listenersLock.Lock()
		delete(pb.listeners, socketId)
		pb.listenersLock.Unlock()
	}()

	for {
		select {
		case <-pb.ctx.Done():
			return io.EOF
		case p := <-l:
			err := c.WritePacket(p)
			p.Decommission()
			if err != nil {
				pb.logger.Print("pubsub:error", socketId, 1, func() string { return err.Error() })
				return err
			}
		}
	}
}
````

## File: rand/rand_test.go
````go
package rand

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomString(t *testing.T) {
	s1, err := RandomString(42, AlphaNumericCharset)
	require.NoError(t, err)

	s2, err := RandomString(42, AlphaNumericCharset)
	require.NoError(t, err)

	require.NotEqual(t, s1, s2)
}

func TestUint32(t *testing.T) {
	u1, err := Uint32()
	require.NoError(t, err)

	u2, err := Uint32()
	require.NoError(t, err)

	require.NotEqual(t, u1, u2)
}

func TestInt63(t *testing.T) {
	u1, err := Int63()
	require.NoError(t, err)

	u2, err := Int63()
	require.NoError(t, err)

	require.NotEqual(t, u1, u2)
}

func TestInt63n(t *testing.T) {
	u1, err := Int63n(42)
	require.NoError(t, err)

	u2, err := Int63n(42)
	require.NoError(t, err)

	require.NotEqual(t, u1, u2)

	u3, err := Int63n(64)
	require.NoError(t, err)

	u4, err := Int63n(64)
	require.NoError(t, err)

	require.NotEqual(t, u3, u4)
}
````

## File: rand/rand.go
````go
package rand

import "crypto/rand"

var AlphaNumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// https://www.calhoun.io/creating-random-strings-in-go/
func RandomString(length int, charset string) (string, error) {
	b := make([]byte, length)
	for i := range b {
		j, err := Int63n(int64(len(charset)))
		if err != nil {
			return "", err
		}
		b[i] = charset[j]
	}

	return string(b), nil
}

func Read(b []byte) (int, error) {
	return rand.Read(b)
}

func Uint32() (uint32, error) {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}

	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

func Int63() (int64, error) {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}

	return int64(uint64(b[0]&0b01111111)<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])), nil
}

// https://cs.opensource.google/go/go/+/refs/tags/go1.20.4:src/math/rand/rand.go;l=119
func Int63n(n int64) (int64, error) {
	if n&(n-1) == 0 { // n is power of two, can mask
		r, err := Int63()
		if err != nil {
			return 0, err
		}
		return r & (n - 1), nil
	}

	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))

	v, err := Int63()
	if err != nil {
		return 0, err
	}

	for v > max {
		v, err = Int63()
		if err != nil {
			return 0, err
		}
	}

	return v % n, nil
}
````

## File: README.md
````markdown
# GoSRT

Implementation of the SRT protocol in pure Go with minimal dependencies.

<p align="left">
  <a href="http://srtalliance.org/">
    <img alt="SRT" src="https://github.com/datarhei/misc/blob/main/img/gosrt.png?raw=true" width="600"/>
  </a>
</p>

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Tests](https://github.com/datarhei/gosrt/actions/workflows/go-tests.yml/badge.svg)
[![codecov](https://codecov.io/gh/datarhei/gosrt/branch/main/graph/badge.svg?token=90YMPZRAFK)](https://codecov.io/gh/datarhei/gosrt)
[![Go Report Card](https://goreportcard.com/badge/github.com/datarhei/gosrt)](https://goreportcard.com/report/github.com/datarhei/gosrt)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/datarhei/gosrt)](https://pkg.go.dev/github.com/datarhei/gosrt)

-   [SRT reference implementation](https://github.com/Haivision/srt)
-   [SRT RFC](https://haivision.github.io/srt-rfc/draft-sharabayko-srt.html)
-   [SRT Technical Overview](https://github.com/Haivision/srt/files/2489142/SRT_Protocol_TechnicalOverview_DRAFT_2018-10-17.pdf)

## Implementations

This implementation of the SRT protocol has live streaming of video/audio in mind. Because of this, the buffer mode and File Transfer
Congestion Control (FileCC) are not implemented.

|     |                                           |
| --- | ----------------------------------------- |
| ✅  | Handshake v4 and v5                       |
| ✅  | Message mode                              |
| ✅  | Caller-Listener Handshake                 |
| ✅  | Timestamp-Based Packet Delivery (TSBPD)   |
| ✅  | Too-Late Packet Drop (TLPKTDROP)          |
| ✅  | Live Congestion Control (LiveCC)          |
| ✅  | NAK and Peridoc NAK                       |
| ✅  | Encryption                                |
| ❌  | Buffer mode                               |
| ❌  | Rendezvous Handshake                      |
| ❌  | File Transfer Congestion Control (FileCC) |
| ❌  | Connection Bonding                        |

The parts that are implemented are based on what has been published in the SRT RFC.

## Requirements

A Go version of 1.20+ is required.

## Installation

```shell
go get github.com/datarhei/gosrt
```

## Caller example

```go
import "github.com/datarhei/gosrt"

conn, err := srt.Dial("srt", "golang.org:6000", srt.Config{
    StreamId: "...",
})
if err != nil {
    // handle error
}

buffer := make([]byte, 2048)

for {
    n, err := conn.Read(buffer)
    if err != nil {
        // handle error
    }

    // handle received data
}

conn.Close()
```

In the `contrib/client` directory you'll find a complete example of a SRT client.

## Listener example

```go
import "github.com/datarhei/gosrt"

ln, err := srt.Listen("srt", ":6000", srt.Config{...})
if err != nil {
    // handle error
}

for {
    req, err := ln.Accept2()
    if err != nil {
        // handle error
    }

    go func(req ConnRequest) {
        // check connection request by inspecting the connection request
        // and either rejecting it ...

        if somethingIsWrong {
            req.Reject(srt.REJ_PEER)
            return
        }

        // ... or accepting it ...

        conn, err := req.Accept()
        if err != nil {
            return
        }

        // ... and decide whether it is a publishing or subscribing connection.

        if publish {
            handlePublish(conn)
        } else {
            handleSubscribe(conn)
        }
    }(req)
}
```

In the `contrib/server` directory you'll find a complete example of a SRT server. For your convenience
this module provides the `Server` type which is a light framework for creating your own SRT server. The
example server is based on this type.

## Contributed client

In the `contrib/client` directory you'll find an example implementation of a SRT client.

Build the client application with

```shell
cd contrib/client && go build
```

The application requires only two options:

| Option  | Description          |
| ------- | -------------------- |
| `-from` | Address to read from |
| `-to`   | Address to write to  |

Both options accept an address. Valid addresses are: `-` for `stdin`, resp. `stdout`, a `srt://` address, or an `udp://` address.

### SRT URL

A SRT URL is of the form `srt://[host]:[port]/?[options]` where options are in the form of a `HTTP` query string. These are the
known options (similar to [srt-live-transmit](https://github.com/Haivision/srt/blob/master/docs/apps/srt-live-transmit.md)):

| Option               | Values                 | Description                                                             |
| -------------------- | ---------------------- | ----------------------------------------------------------------------- |
| `mode`               | `listener` or `caller` | Enforce listener or caller mode.                                        |
| `congestion`         | `live`                 | Congestion control. Currently only `live` is supported.                 |
| `conntimeo`          | `ms`                   | Connection timeout.                                                     |
| `drifttracer`        | `bool`                 | Enable drift tracer. Not implemented.                                   |
| `enforcedencryption` | `bool`                 | Accept connection only if both parties have encryption enabled.         |
| `fc`                 | `bytes`                | Flow control window size.                                               |
| `inputbw`            | `bytes`                | Input bandwidth. Ignored.                                               |
| `iptos`              | 0...255                | IP socket type of service. Broken.                                      |
| `ipttl`              | 1...255                | Defines IP socket "time to live" option. Broken.                        |
| `ipv6only`           | `bool`                 | Use IPv6 only. Not implemented.                                         |
| `kmpreannounce`      | `packets`              | Duration of Stream Encryption key switchover.                           |
| `kmrefreshrate`      | `packets`              | Stream encryption key refresh rate.                                     |
| `latency`            | `ms`                   | Maximum accepted transmission latency.                                  |
| `lossmaxttl`         | `ms`                   | Packet reorder tolerance. Not implemented.                              |
| `maxbw`              | `bytes`                | Bandwidth limit. Ignored.                                               |
| `mininputbw`         | `bytes`                | Minimum allowed estimate of `inputbw`.                                  |
| `messageapi`         | `bool`                 | Enable SRT message mode. Must be `false`.                               |
| `mss`                | 76...                  | MTU size.                                                               |
| `nakreport`          | `bool`                 | Enable periodic NAK reports.                                            |
| `oheadbw`            | 10...100               | Limits bandwidth overhead. Percents. Ignored.                           |
| `packetfilter`       | `string`               | Set up the packet filter. Not implemented.                              |
| `passphrase`         | `string`               | Password for the encrypted transmission.                                |
| `payloadsize`        | `bytes`                | Maximum payload size.                                                   |
| `pbkeylen`           | `16`, `24`, or `32`    | Crypto key length in bytes.                                             |
| `peeridletimeo`      | `ms`                   | Peer idle timeout.                                                      |
| `peerlatency`        | `ms`                   | Minimum receiver latency to be requested by sender.                     |
| `rcvbuf`             | `bytes`                | Receiver buffer size.                                                   |
| `rcvlatency`         | `ms`                   | Receiver-side latency.                                                  |
| `sndbuf`             | `bytes`                | Sender buffer size.                                                     |
| `snddropdelay`       | `ms`                   | Sender's delay before dropping packets.                                 |
| `streamid`           | `string`               | Stream ID (settable in caller mode only, visible on the listener peer). |
| `tlpktdrop`          | `bool`                 | Drop too late packets.                                                  |
| `transtype`          | `live`                 | Transmission type. Must be `live`.                                      |
| `tsbpdmode`          | `bool`                 | Enable timestamp-based packet delivery mode.                            |

### Usage

Reading from a SRT sender and play with `ffplay`:

```shell
./client -from "srt://127.0.0.1:6001/?mode=listener&streamid=..." -to - | ffplay -f mpegts -i -
```

Reading from UDP and sending to a SRT server:

```shell
./client -from udp://:6000 -to "srt://127.0.0.1:6001/?mode=caller&streamid=..."
```

Simulate point-to-point transfer on localhost. Open one console and start `ffmpeg` (you need at least version 4.3.2, built with SRT enabled) to send to an UDP address:

```shell
ffmpeg \
    -f lavfi \
    -re \
    -i testsrc2=rate=25:size=640x360 \
    -codec:v libx264 \
    -b:v 1024k \
    -maxrate:v 1024k \
    -bufsize:v 1024k \
    -preset ultrafast \
    -r 25 \
    -g 50 \
    -pix_fmt yuv420p \
    -flags2 local_header \
    -f mpegts \
    "udp://127.0.0.1:6000?pkt_size=1316"
```

In another console read from the UDP and start a SRT listenr:

```shell
./client -from udp://:6000 -to "srt://127.0.0.1:6001/?mode=listener&streamid=foobar"
```

In the third console connect to that stream and play the video with `ffplay`:

```shell
./client -from "srt://127.0.0.1:6001/?mode=caller&streamid=foobar" -to - | ffplay -f mpegts -i -
```

## Contributed server

In the `contrib/server` directory you'll find an example implementation of a SRT server. This server allows you to publish
a stream that can be read by many clients.

Build the client application with

```shell
cd contrib/server && go build
```

The application has these options:

| Option        | Default   | Description                                |
| ------------- | --------- | ------------------------------------------ |
| `-addr`       | required  | Address to listen on                       |
| `-app`        | `/`       | Path prefix for streamid                   |
| `-token`      | (not set) | Token query param for streamid             |
| `-passphrase` | (not set) | Passphrase for de- and enrcypting the data |
| `-logtopics`  | (not set) | Topics for the log output                  |
| `-profile`    | `false`   | Enable profiling                           |

This example server expects the streamID (without any prefix) to be an URL path with optional query parameter, e.g. `/live/stream`. If the `-app`
option is used, then the path must start with that path, e.g. the value is `/live` then the streamID must start with that value. The `-token`
option can be used to define a token for that stream as some kind of access control, e.g. with `-token foobar` the streamID might look like
`/live/stream?token=foobar`.

Use `-passphrase` in order to enable and enforce encryption.

Use `-logtopics` in order to write debug output. The value are a comma separated list of topics you want to be written to `stderr`, e.g. `connection,listen`. Check the [Logging](#logging) section in order to find out more about the different topics.

Use `-profile` in order to write a CPU profile.

### StreamID

In SRT the StreamID is used to transport somewhat arbitrary information from the caller to the listener. The provided example server uses this
machanism to decide who is the sender and who is the receiver. The server must know if the connecting client wants to publish a stream or
if it wants to subscribe to a stream.

The example server looks for the `publish:` prefix in the StreamID. If this prefix is present, the server assumes that it is the receiver
and the client will send the data. The subcribing clients must use the same StreamID (withouth the `publish:` prefix) in order to be able to
receive data.

If you implement your own server you are free to interpret the streamID as you wish.

### Usage

Running a server listening on port 6001 with defaults:

```shell
./server -addr ":6001"
```

Now you can use the contributed client to publish a stream:

```shell
./client -from ... -to "srt://127.0.0.1:6001/?mode=caller&streamid=publish:/live/stream"
```

or directly from `ffmpeg`:

```shell
ffmpeg \
    -f lavfi \
    -re \
    -i testsrc2=rate=25:size=640x360 \
    -codec:v libx264 \
    -b:v 1024k \
    -maxrate:v 1024k \
    -bufsize:v 1024k \
    -preset ultrafast \
    -r 25 \
    -g 50 \
    -pix_fmt yuv420p \
    -flags2 local_header \
    -f mpegts \
    -transtype live \
    "srt://127.0.0.1:6001?streamid=publish:/live/stream"
```

If the server is not on localhost, you might adjust the `peerlatency` in order to avoid packet loss: `-peerlatency 1000000`.

Now you can play the stream:

```shell
ffplay -f mpegts -transtype live -i "srt://127.0.0.1:6001?streamid=/live/stream"
```

You will most likely first see some error messages from `ffplay` because it tries to make sense of the received data until a keyframe arrives. If you
get more errors during playback, you might increase the receive buffer by adding e.g. `-rcvlatency 1000000` to the command line.

### Encryption

The stream can be encrypted with a passphrase. First start the server with a passphrase. If you are using `srt-live-transmit`, the passphrase has to be at least 10 characters long otherwise it will not be accepted.

```shell
./server -addr :6001 -passphrase foobarfoobar
```

Send an encrpyted stream to the server:

```shell
ffmpeg \
    -f lavfi \
    -re \
    -i testsrc2=rate=25:size=640x360 \
    -codec:v libx264 \
    -b:v 1024k \
    -maxrate:v 1024k \
    -bufsize:v 1024k \
    -preset ultrafast \
    -r 25 \
    -g 50 \
    -pix_fmt yuv420p \
    -flags2 local_header \
    -f mpegts \
    -transtype live \
    "srt://127.0.0.1:6001?streamid=publish:/live/stream&passphrase=foobarfoobar"
```

Receive an encrypted stream from the server:

```shell
ffplay -f mpegts -transtype live -i "srt://127.0.0.1:6001?streamid=/live/stream&passphrase=foobarfoobar"
```

You will most likely first see some error messages from `ffplay` because it tries to make sense of the received data until a keyframe arrives. If you
get more errors during playback, you might increase the receive buffer by adding e.g. `-rcvlatency 1000000` to the command line.

## Logging

This SRT module has a built-in logging facility for debugging purposes. Check the `Logger` interface and the `NewLogger(topics []string)` function. Because logging everything would be too much output if you wonly want to debug something specific, you have the possibility to limit the logging to specific areas like everything regarding a connection or only the handshake. That's why there are various topics.

In the contributed server you see an example of how logging is used. Here's the essence:

```go
logger := srt.NewLogger([]string{"connection", "handshake"})

config := srt.DefaultConfig
config.Logger = logger

ln, err := srt.Listen("udp", ":6000", config)
if err != nil {
    // handle error
}

go func() {
    for m := range logger.Listen() {
        fmt.Fprintf(os.Stderr, "%#08x %s (in %s:%d)\n%s \n", m.SocketId, m.Topic, m.File, m.Line, m.Message)
    }
}()

for {
    conn, mode, err := ln.Accept(acceptFn)
    ...
}
```

Currently known topics are:

```
connection:close
connection:error
connection:filter
connection:new
connection:rtt
connection:tsbpd
control:recv:ACK:cif
control:recv:ACK:dump
control:recv:ACK:error
control:recv:ACKACK:dump
control:recv:ACKACK:error
control:recv:KM:cif
control:recv:KM:dump
control:recv:KM:error
control:recv:NAK:cif
control:recv:NAK:dump
control:recv:NAK:error
control:recv:keepalive:dump
control:recv:shutdown:dump
control:send:ACK:cif
control:send:ACK:dump
control:send:ACKACK:dump
control:send:KM:cif
control:send:KM:dump
control:send:KM:error
control:send:NAK:cif
control:send:NAK:dump
control:send:keepalive:dump
control:send:shutdown:cif
control:send:shutdown:dump
data:recv:dump
data:send:dump
dial
handshake:recv:cif
handshake:recv:dump
handshake:recv:error
handshake:send:cif
handshake:send:dump
listen
packet:recv:dump
packet:send:dump
```

You can run `make logtopics` in order to extract the list of topics.

## Docker

The docker image you can build with `docker build -t srt .` provides the example SRT client and server as mentioned in the paragraph above.
E.g. run the server with `docker run -it --rm -p 6001:6001/udp srt srt-server -addr :6001`.
````

## File: SECURITY.md
````markdown
# Security Policy

## Supported Versions

All versions in the list receive security updates.

| Version | Supported          |
| ------- | ------------------ |
| 0.1.1   | :white_check_mark: |
| -       | :x:                |

## Reporting a Vulnerability

If you have found or just suspect a security problem somewhere in Restreamer or Core, report it on support@datarhei.com.

We treat security issues with confidentiality until controlled and disclosed responsibly.
````

## File: server_test.go
````go
package srt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	server := Server{
		Addr: "127.0.0.1:6003",
		HandleConnect: func(req ConnRequest) ConnType {
			streamid := req.StreamId()

			if streamid == "publish" {
				return PUBLISH
			} else if streamid == "subscribe" {
				return SUBSCRIBE
			}

			return REJECT
		},
	}

	err := server.Listen()
	require.NoError(t, err)

	defer server.Shutdown()

	go func() {
		err := server.Serve()
		if err == ErrServerClosed {
			return
		}
		require.NoError(t, err)
	}()

	config := DefaultConfig()
	config.StreamId = "publish"

	conn, err := Dial("srt", "127.0.0.1:6003", config)
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)

	config = DefaultConfig()
	config.StreamId = "subscribe"

	conn, err = Dial("srt", "127.0.0.1:6003", config)
	require.NoError(t, err)

	err = conn.Close()
	require.NoError(t, err)

	config = DefaultConfig()
	config.StreamId = "nothing"

	_, err = Dial("srt", "127.0.0.1:6003", config)
	require.Error(t, err)
}
````

## File: server.go
````go
package srt

import (
	"errors"
)

// Server is a framework for a SRT server
type Server struct {
	// The address the SRT server should listen on, e.g. ":6001".
	Addr string

	// Config is the configuration for a SRT listener.
	Config *Config

	// HandleConnect will be called for each incoming connection. This
	// allows you to implement your own interpretation of the streamid
	// and authorization. If this is nil, all connections will be
	// rejected.
	HandleConnect AcceptFunc

	// HandlePublish will be called for a publishing connection.
	HandlePublish func(conn Conn)

	// HandlePublish will be called for a subscribing connection.
	HandleSubscribe func(conn Conn)

	ln Listener
}

// ErrServerClosed is returned when the server is about to shutdown.
var ErrServerClosed = errors.New("srt: server closed")

// ListenAndServe starts the SRT server. It blocks until an error happens.
// If the error is ErrServerClosed the server has shutdown normally.
func (s *Server) ListenAndServe() error {
	err := s.Listen()
	if err != nil {
		return err
	}

	return s.Serve()
}

// Listen opens the server listener.
// It returns immediately after the listener is ready.
func (s *Server) Listen() error {
	// Set some defaults if required.
	if s.HandlePublish == nil {
		s.HandlePublish = s.defaultHandler
	}

	if s.HandleSubscribe == nil {
		s.HandleSubscribe = s.defaultHandler
	}

	if s.Config == nil {
		config := DefaultConfig()
		s.Config = &config
	}

	// Start listening for incoming connections.
	ln, err := Listen("srt", s.Addr, *s.Config)
	if err != nil {
		return err
	}

	s.ln = ln

	return err
}

// Serve starts accepting connections. It must be called after Listen().
// It blocks until an error happens.
// If the error is ErrServerClosed the server has shutdown normally.
func (s *Server) Serve() error {
	for {
		// Wait for connections.
		req, err := s.ln.Accept2()
		if err != nil {
			if err == ErrListenerClosed {
				return ErrServerClosed
			}

			return err
		}

		if s.HandleConnect == nil {
			req.Reject(REJ_PEER)
			continue
		}

		go func(req ConnRequest) {
			mode := s.HandleConnect(req)
			if mode == REJECT {
				req.Reject(REJ_PEER)
				return
			}

			conn, err := req.Accept()
			if err != nil {
				// rejected connection, ignore
				return
			}

			if mode == PUBLISH {
				s.HandlePublish(conn)
			} else {
				s.HandleSubscribe(conn)
			}
		}(req)
	}
}

// Shutdown will shutdown the server. ListenAndServe will return a ErrServerClosed
func (s *Server) Shutdown() {
	if s.ln == nil {
		return
	}

	// Close the listener
	s.ln.Close()
}

func (s *Server) defaultHandler(conn Conn) {
	// Close the incoming connection
	conn.Close()
}
````

## File: statistics.go
````go
// https://github.com/Haivision/srt/blob/master/docs/API/statistics.md

package srt

// Statistics represents the statistics for a connection
type Statistics struct {
	MsTimeStamp uint64 // The time elapsed, in milliseconds, since the SRT socket has been created

	// Accumulated
	Accumulated StatisticsAccumulated

	// Interval
	Interval StatisticsInterval

	// Instantaneous
	Instantaneous StatisticsInstantaneous
}

type StatisticsAccumulated struct {
	PktSent          uint64 // The total number of sent DATA packets, including retransmitted packets
	PktRecv          uint64 // The total number of received DATA packets, including retransmitted packets
	PktSentUnique    uint64 // The total number of unique DATA packets sent by the SRT sender
	PktRecvUnique    uint64 // The total number of unique original, retransmitted or recovered by the packet filter DATA packets received in time, decrypted without errors and, as a result, scheduled for delivery to the upstream application by the SRT receiver.
	PktSendLoss      uint64 // The total number of data packets considered or reported as lost at the sender side. Does not correspond to the packets detected as lost at the receiver side.
	PktRecvLoss      uint64 // The total number of SRT DATA packets detected as presently missing (either reordered or lost) at the receiver side
	PktRetrans       uint64 // The total number of retransmitted packets sent by the SRT sender
	PktRecvRetrans   uint64 // The total number of retransmitted packets registered at the receiver side
	PktSentACK       uint64 // The total number of sent ACK (Acknowledgement) control packets
	PktRecvACK       uint64 // The total number of received ACK (Acknowledgement) control packets
	PktSentNAK       uint64 // The total number of sent NAK (Negative Acknowledgement) control packets
	PktRecvNAK       uint64 // The total number of received NAK (Negative Acknowledgement) control packets
	PktSentKM        uint64 // The total number of sent KM (Key Material) control packets
	PktRecvKM        uint64 // The total number of received KM (Key Material) control packets
	UsSndDuration    uint64 // The total accumulated time in microseconds, during which the SRT sender has some data to transmit, including packets that have been sent, but not yet acknowledged
	PktRecvBelated   uint64
	PktSendDrop      uint64 // The total number of dropped by the SRT sender DATA packets that have no chance to be delivered in time
	PktRecvDrop      uint64 // The total number of dropped by the SRT receiver and, as a result, not delivered to the upstream application DATA packets
	PktRecvUndecrypt uint64 // The total number of packets that failed to be decrypted at the receiver side

	ByteSent          uint64 // Same as pktSent, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecv          uint64 // Same as pktRecv, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteSentUnique    uint64 // Same as pktSentUnique, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvUnique    uint64 // Same as pktRecvUnique, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvLoss      uint64 // Same as pktRecvLoss, but expressed in bytes, including payload and all the headers (IP, TCP, SRT), bytes for the presently missing (either reordered or lost) packets' payloads are estimated based on the average packet size
	ByteRetrans       uint64 // Same as pktRetrans, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvRetrans   uint64 // Same as pktRecvRetrans, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvBelated   uint64
	ByteSendDrop      uint64 // Same as pktSendDrop, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvDrop      uint64 // Same as pktRecvDrop, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvUndecrypt uint64 // Same as pktRecvUndecrypt, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
}

type StatisticsInterval struct {
	MsInterval uint64 // Length of the interval, in milliseconds

	PktSent        uint64 // Number of sent DATA packets, including retransmitted packets
	PktRecv        uint64 // Number of received DATA packets, including retransmitted packets
	PktSentUnique  uint64 // Number of unique DATA packets sent by the SRT sender
	PktRecvUnique  uint64 // Number of unique original, retransmitted or recovered by the packet filter DATA packets received in time, decrypted without errors and, as a result, scheduled for delivery to the upstream application by the SRT receiver.
	PktSendLoss    uint64 // Number of data packets considered or reported as lost at the sender side. Does not correspond to the packets detected as lost at the receiver side.
	PktRecvLoss    uint64 // Number of SRT DATA packets detected as presently missing (either reordered or lost) at the receiver side
	PktRetrans     uint64 // Number of retransmitted packets sent by the SRT sender
	PktRecvRetrans uint64 // Number of retransmitted packets registered at the receiver side
	PktSentACK     uint64 // Number of sent ACK (Acknowledgement) control packets
	PktRecvACK     uint64 // Number of received ACK (Acknowledgement) control packets
	PktSentNAK     uint64 // Number of sent NAK (Negative Acknowledgement) control packets
	PktRecvNAK     uint64 // Number of received NAK (Negative Acknowledgement) control packets

	MbpsSendRate float64 // Sending rate, in Mbps
	MbpsRecvRate float64 // Receiving rate, in Mbps

	UsSndDuration uint64 // Accumulated time in microseconds, during which the SRT sender has some data to transmit, including packets that have been sent, but not yet acknowledged

	PktReorderDistance uint64
	PktRecvBelated     uint64 // Number of packets that arrive too late
	PktSndDrop         uint64 // Number of dropped by the SRT sender DATA packets that have no chance to be delivered in time
	PktRecvDrop        uint64 // Number of dropped by the SRT receiver and, as a result, not delivered to the upstream application DATA packets
	PktRecvUndecrypt   uint64 // Number of packets that failed to be decrypted at the receiver side

	ByteSent          uint64 // Same as pktSent, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecv          uint64 // Same as pktRecv, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteSentUnique    uint64 // Same as pktSentUnique, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvUnique    uint64 // Same as pktRecvUnique, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvLoss      uint64 // Same as pktRecvLoss, but expressed in bytes, including payload and all the headers (IP, TCP, SRT), bytes for the presently missing (either reordered or lost) packets' payloads are estimated based on the average packet size
	ByteRetrans       uint64 // Same as pktRetrans, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvRetrans   uint64 // Same as pktRecvRetrans, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvBelated   uint64 // Same as pktRecvBelated, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteSendDrop      uint64 // Same as pktSendDrop, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvDrop      uint64 // Same as pktRecvDrop, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
	ByteRecvUndecrypt uint64 // Same as pktRecvUndecrypt, but expressed in bytes, including payload and all the headers (IP, TCP, SRT)
}

type StatisticsInstantaneous struct {
	UsPktSendPeriod       float64 // Current minimum time interval between which consecutive packets are sent, in microseconds
	PktFlowWindow         uint64  // The maximum number of packets that can be "in flight"
	PktFlightSize         uint64  // The number of packets in flight
	MsRTT                 float64 // Smoothed round-trip time (SRTT), an exponentially-weighted moving average (EWMA) of an endpoint's RTT samples, in milliseconds
	MbpsSentRate          float64 // Current transmission bandwidth, in Mbps
	MbpsRecvRate          float64 // Current receiving bandwidth, in Mbps
	MbpsLinkCapacity      float64 // Estimated capacity of the network link, in Mbps
	ByteAvailSendBuf      uint64  // The available space in the sender's buffer, in bytes
	ByteAvailRecvBuf      uint64  // The available space in the receiver's buffer, in bytes
	MbpsMaxBW             float64 // Transmission bandwidth limit, in Mbps
	ByteMSS               uint64  // Maximum Segment Size (MSS), in bytes
	PktSendBuf            uint64  // The number of packets in the sender's buffer that are already scheduled for sending or even possibly sent, but not yet acknowledged
	ByteSendBuf           uint64  // Instantaneous (current) value of pktSndBuf, but expressed in bytes, including payload and all headers (IP, TCP, SRT)
	MsSendBuf             uint64  // The timespan (msec) of packets in the sender's buffer (unacknowledged packets)
	MsSendTsbPdDelay      uint64  // Timestamp-based Packet Delivery Delay value of the peer
	PktRecvBuf            uint64  // The number of acknowledged packets in receiver's buffer
	ByteRecvBuf           uint64  // Instantaneous (current) value of pktRcvBuf, expressed in bytes, including payload and all headers (IP, TCP, SRT)
	MsRecvBuf             uint64  // The timespan (msec) of acknowledged packets in the receiver's buffer
	MsRecvTsbPdDelay      uint64  // Timestamp-based Packet Delivery Delay value set on the socket via SRTO_RCVLATENCY or SRTO_LATENCY
	PktReorderTolerance   uint64  // Instant value of the packet reorder tolerance
	PktRecvAvgBelatedTime uint64  // Accumulated difference between the current time and the time-to-play of a packet that is received late
	PktSendLossRate       float64 // Percentage of resent data vs. sent data
	PktRecvLossRate       float64 // Percentage of retransmitted data vs. received data
}
````
