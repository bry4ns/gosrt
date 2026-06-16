package live

import (
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
	packetBuf                   map[uint32]packet.Packet // O(1) ring buffer: seq -> packet
	lock                        sync.RWMutex

	nPackets uint

	periodicACKInterval uint64 // config
	periodicNAKInterval uint64 // config
	lossMaxTTL          uint32 // config: reorder tolerance

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
		packetBuf:                   make(map[uint32]packet.Packet),

		periodicACKInterval: config.PeriodicACKInterval,
		periodicNAKInterval: config.PeriodicNAKInterval,
		lossMaxTTL:          config.LossMaxTTL,

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

	r.packetBuf = make(map[uint32]packet.Packet)
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

	if pkt.Header().RetransmittedPacketFlag {
		r.statistics.PktRetrans++
		r.statistics.ByteRetrans += pktLen

		r.rate.bytesRetrans += pktLen
	}

	//  5.1.2. SRT's Default LiveCC Algorithm
	r.avgPayloadSize = 0.875*r.avgPayloadSize + 0.125*float64(pktLen)

	seq := pkt.Header().PacketSequenceNumber

	if seq.Lte(r.lastDeliveredSequenceNumber) {
		// Too old, because up until r.lastDeliveredSequenceNumber, we already delivered
		r.statistics.PktBelated++
		r.statistics.ByteBelated += pktLen

		r.statistics.PktDrop++
		r.statistics.ByteDrop += pktLen

		return
	}

	if seq.Lt(r.lastACKSequenceNumber) {
		// Already acknowledged, ignoring
		r.statistics.PktDrop++
		r.statistics.ByteDrop += pktLen

		return
	}

	if seq.Equals(r.maxSeenSequenceNumber.Inc()) {
		// In order, the packet we expected
		r.maxSeenSequenceNumber = seq
	} else if seq.Lte(r.maxSeenSequenceNumber) {
		// Out of order, is it a missing piece? O(1) map insert
		seqVal := seq.Val()
		if _, exists := r.packetBuf[seqVal]; exists {
			// Already received (has been sent more than once), ignoring
			r.statistics.PktDrop++
			r.statistics.ByteDrop += pktLen
			return
		}

		// Late arrival, this fills a gap
		r.statistics.PktBuf++
		r.statistics.PktUnique++

		r.statistics.ByteBuf += pktLen
		r.statistics.ByteUnique += pktLen

		r.packetBuf[seqVal] = pkt
		return
	} else {
		// Too far ahead, there are some missing sequence numbers, immediate NAK report
		// here we can prevent a possibly unnecessary NAK with SRTO_LOSSMAXTTL
		if r.lossMaxTTL == 0 || uint64(seq.Distance(r.maxSeenSequenceNumber)) > uint64(r.lossMaxTTL) {
			r.sendNAK([]circular.Number{
				r.maxSeenSequenceNumber.Inc(),
				seq.Dec(),
			})
		}

		len := uint64(seq.Distance(r.maxSeenSequenceNumber))
		r.statistics.PktLoss += len
		r.statistics.ByteLoss += len * uint64(r.avgPayloadSize)

		r.maxSeenSequenceNumber = seq
	}

	r.statistics.PktBuf++
	r.statistics.PktUnique++

	r.statistics.ByteBuf += pktLen
	r.statistics.ByteUnique += pktLen

	r.packetBuf[seq.Val()] = pkt
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

	ackSequenceNumber := r.lastACKSequenceNumber

	// Scan forward from lastACK+1, find consecutive ripe packets in map
	// (skip gaps — equivalent to linked list iteration which only sees existing packets)
	maxIter := uint32(1000)
	if r.lossMaxTTL > 0 {
		maxIter = r.lossMaxTTL + 100
	}

	checkSeq := ackSequenceNumber.Inc()
	for i := uint32(0); i < maxIter; i++ {
		// Stop if we've gone past maxSeen
		if checkSeq.Gt(r.maxSeenSequenceNumber) {
			break
		}

		pkt, exists := r.packetBuf[checkSeq.Val()]
		if !exists {
			// Gap — skip and continue
			checkSeq = checkSeq.Inc()
			continue
		}

		// If there are packets that should have been delivered by now, move forward.
		if pkt.Header().PktTsbpdTime <= now {
			ackSequenceNumber = checkSeq
			checkSeq = checkSeq.Inc()
			continue
		}

		// Check if the packet is the next in the row.
		if checkSeq.Equals(ackSequenceNumber.Inc()) {
			ackSequenceNumber = checkSeq
			checkSeq = checkSeq.Inc()
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

	// Scan forward from lastACK+1, find gaps in map
	maxIter := uint32(1000)
	if r.lossMaxTTL > 0 {
		maxIter = r.lossMaxTTL + 100
	}

	checkSeq := ackSequenceNumber.Inc()
	for i := uint32(0); i < maxIter; i++ {
		// Stop if we've gone past maxSeen
		if checkSeq.Gt(r.maxSeenSequenceNumber) {
			break
		}

		_, exists := r.packetBuf[checkSeq.Val()]
		if exists {
			ackSequenceNumber = checkSeq
			checkSeq = checkSeq.Inc()
			continue
		}

		// Gap found — find the end of the gap
		nackStart := checkSeq
		nackEnd := checkSeq

		// Scan forward to find end of gap
		for j := uint32(0); j < maxIter; j++ {
			nackEnd = checkSeq
			checkSeq = checkSeq.Inc()

			if checkSeq.Gt(r.maxSeenSequenceNumber) {
				break
			}

			_, exists := r.packetBuf[checkSeq.Val()]
			if exists {
				break
			}
		}

		if r.lossMaxTTL > 0 {
			if nackEnd.Gte(nakLimit) {
				nackEnd = nakLimit.Dec()
			}
		}

		if nackStart.Lte(nackEnd) {
			list = append(list, nackStart)
			list = append(list, nackEnd)
		}

		ackSequenceNumber = nackEnd
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

	// Scan from lastDelivered+1 forward, deliver ripe packets (skip gaps)
	checkSeq := r.lastDeliveredSequenceNumber.Inc()
	maxIter := uint32(1000)
	if r.lossMaxTTL > 0 {
		maxIter = r.lossMaxTTL + 100
	}

	for i := uint32(0); i < maxIter; i++ {
		// Stop if we've gone past lastACK
		if checkSeq.Gt(r.lastACKSequenceNumber) {
			break
		}

		pkt, exists := r.packetBuf[checkSeq.Val()]
		if !exists {
			// Gap — skip this sequence and continue
			checkSeq = checkSeq.Inc()
			continue
		}

		// Only deliver if ripe
		if pkt.Header().PktTsbpdTime <= now {
			r.statistics.PktBuf--
			r.statistics.ByteBuf -= pkt.Len()
			r.lastDeliveredSequenceNumber = checkSeq

			r.deliver(pkt)
			delete(r.packetBuf, checkSeq.Val())
			checkSeq = checkSeq.Inc()
		} else {
			break
		}
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
	for seq, pkt := range r.packetBuf {
		b.WriteString(fmt.Sprintf("   %d @ %d (in %d)\n", seq, pkt.Header().PktTsbpdTime, int64(pkt.Header().PktTsbpdTime)-int64(t)))
	}
	r.lock.RUnlock()

	return b.String()
}
