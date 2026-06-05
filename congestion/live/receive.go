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

	// Reorder buffer: delivers packets in strict sequence order.
	// Handles SRTLA bonding where packets arrive out of order from multiple SIMs.
	nextDeliverySeq     circular.Number           // next sequence to deliver to application
	lastDeliveryAdvance time.Time                 // when nextDeliverySeq last advanced
	deliveryStaleMs     int                       // stale timeout in ms (from LossMaxTTL * 2)
}

// NewReceiver takes a ReceiveConfig and returns a new Receiver
func NewReceiver(config ReceiveConfig) congestion.Receiver {
	staleMs := int(config.LossMaxTTL) * 2
	if staleMs < 100 {
		staleMs = 100
	}
	if staleMs > 500 {
		staleMs = 500
	}

	r := &receiver{
		maxSeenSequenceNumber:       config.InitialSequenceNumber.Dec(),
		lastACKSequenceNumber:       config.InitialSequenceNumber.Dec(),
		lastDeliveredSequenceNumber: config.InitialSequenceNumber.Dec(),
		packetList:                  list.New(),

		periodicACKInterval: config.PeriodicACKInterval,
		periodicNAKInterval: config.PeriodicNAKInterval,
		lossMaxTTL:           config.LossMaxTTL,

		avgPayloadSize: 1456, //  5.1.2. SRT's Default LiveCC Algorithm

		sendACK: config.OnSendACK,
		sendNAK: config.OnSendNAK,
		deliver: config.OnDeliver,

		// Reorder buffer init
		nextDeliverySeq:     config.InitialSequenceNumber.Dec(),
		lastDeliveryAdvance: time.Now(),
		deliveryStaleMs:     staleMs,
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

	if pkt.Header().RetransmittedPacketFlag {
		r.statistics.PktRetrans++
		r.statistics.ByteRetrans += pktLen

		r.rate.bytesRetrans += pktLen
	}

	//  5.1.2. SRT's Default LiveCC Algorithm
	r.avgPayloadSize = 0.875*r.avgPayloadSize + 0.125*float64(pktLen)

	seq := pkt.Header().PacketSequenceNumber

	// REORDER BUFFER: For bonding (lossMaxTTL > 0), buffer out-of-order packets
	// and deliver in strict sequence order. This prevents H.264 corruption from
	// packets delivered ahead of gaps.
	if r.lossMaxTTL > 0 {
		// Drop duplicates (already delivered)
		if seq.Lte(r.lastDeliveredSequenceNumber) {
			// Check if it's within the reorder window (late arrival that fills a gap)
			dist := r.lastDeliveredSequenceNumber.Distance(seq)
			if dist <= r.lossMaxTTL {
				// Try to insert into packetList for potential delivery
				inserted := false
				for e := r.packetList.Front(); e != nil; e = e.Next() {
					p := e.Value.(packet.Packet)
					if p.Header().PacketSequenceNumber == seq {
						r.statistics.PktDrop++
						r.statistics.ByteDrop += pktLen
						inserted = true
						break
					}
					if p.Header().PacketSequenceNumber.Gt(seq) {
						r.packetList.InsertBefore(pkt, e)
						r.statistics.PktBuf++
						r.statistics.PktUnique++
						r.statistics.ByteBuf += pktLen
						r.statistics.ByteUnique += pktLen
						inserted = true
						break
					}
				}
				if !inserted && r.packetList.Len() == 0 {
					// Empty list, append
					r.packetList.PushBack(pkt)
					r.statistics.PktBuf++
					r.statistics.PktUnique++
					r.statistics.ByteBuf += pktLen
					r.statistics.ByteUnique += pktLen
				}
			} else {
				r.statistics.PktBelated++
				r.statistics.ByteBelated += pktLen
				r.statistics.PktDrop++
				r.statistics.ByteDrop += pktLen
			}
			return
		}

		// Buffer out-of-order packets (don't deliver yet, wait for gap to complete)
		if seq.Lt(r.maxSeenSequenceNumber) || seq.Equals(r.maxSeenSequenceNumber) {
			// Out of order: insert into packetList sorted
			inserted := false
			for e := r.packetList.Front(); e != nil; e = e.Next() {
				p := e.Value.(packet.Packet)
				if p.Header().PacketSequenceNumber == seq {
					r.statistics.PktDrop++
					r.statistics.ByteDrop += pktLen
					inserted = true
					break
				}
				if p.Header().PacketSequenceNumber.Gt(seq) {
					r.packetList.InsertBefore(pkt, e)
					r.statistics.PktBuf++
					r.statistics.PktUnique++
					r.statistics.ByteBuf += pktLen
					r.statistics.ByteUnique += pktLen
					inserted = true
					break
				}
			}
			if !inserted {
				r.packetList.PushBack(pkt)
				r.statistics.PktBuf++
				r.statistics.PktUnique++
				r.statistics.ByteBuf += pktLen
				r.statistics.ByteUnique += pktLen
			}
			return
		}

		// Gap detected: packet is ahead of maxSeen
		if seq.Gt(r.maxSeenSequenceNumber.Inc()) {
			// Send NAK if gap exceeds lossMaxTTL
			if uint64(seq.Distance(r.maxSeenSequenceNumber.Inc())) > uint64(r.lossMaxTTL) {
				r.sendNAK([]circular.Number{
					r.maxSeenSequenceNumber.Inc(),
					seq.Dec(),
				})
			}
			r.statistics.PktLoss += uint64(seq.Distance(r.maxSeenSequenceNumber.Inc()))
			r.statistics.ByteLoss += uint64(seq.Distance(r.maxSeenSequenceNumber.Inc())) * uint64(r.avgPayloadSize)
		}

		// In order (or ahead): update maxSeen and add to packetList
		r.maxSeenSequenceNumber = seq
		r.packetList.PushBack(pkt)
		r.statistics.PktBuf++
		r.statistics.PktUnique++
		r.statistics.ByteBuf += pktLen
		r.statistics.ByteUnique += pktLen
		return
	}

	// ORIGINAL BEHAVIOR (lossMaxTTL == 0): no reorder buffer
	if pkt.Header().PacketSequenceNumber.Lte(r.lastDeliveredSequenceNumber) {
		r.statistics.PktBelated++
		r.statistics.ByteBelated += pktLen
		r.statistics.PktDrop++
		r.statistics.ByteDrop += pktLen
		return
	}

	if pkt.Header().PacketSequenceNumber.Lt(r.lastACKSequenceNumber) {
		r.statistics.PktDrop++
		r.statistics.ByteDrop += pktLen
		return
	}

	if pkt.Header().PacketSequenceNumber.Equals(r.maxSeenSequenceNumber.Inc()) {
		r.maxSeenSequenceNumber = pkt.Header().PacketSequenceNumber
	} else if pkt.Header().PacketSequenceNumber.Lte(r.maxSeenSequenceNumber) {
		for e := r.packetList.Front(); e != nil; e = e.Next() {
			p := e.Value.(packet.Packet)
			if p.Header().PacketSequenceNumber == pkt.Header().PacketSequenceNumber {
				r.statistics.PktDrop++
				r.statistics.ByteDrop += pktLen
				break
			} else if p.Header().PacketSequenceNumber.Gt(pkt.Header().PacketSequenceNumber) {
				r.statistics.PktBuf++
				r.statistics.PktUnique++
				r.statistics.ByteBuf += pktLen
				r.statistics.ByteUnique += pktLen
				r.packetList.InsertBefore(pkt, e)
				break
			}
		}
		return
	} else {
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

	// For bonding: ACK can advance past gaps within lossMaxTTL
	// This tells the sender "I got everything up to here" even if there are gaps
	// The reorder buffer in Tick() handles actual delivery order
	for e := r.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		if p.Header().PacketSequenceNumber.Lte(ackSequenceNumber) {
			continue
		}

		if p.Header().PktTsbpdTime <= now {
			ackSequenceNumber = p.Header().PacketSequenceNumber
			continue
		}

		if p.Header().PacketSequenceNumber.Equals(ackSequenceNumber.Inc()) {
			ackSequenceNumber = p.Header().PacketSequenceNumber
			maxPktTsbpdTime = p.Header().PktTsbpdTime
			r.statistics.MsBuf = (maxPktTsbpdTime - minPktTsbpdTime) / 1_000
			continue
		}

		// For bonding: skip gaps within tolerance for ACK advancement
		if r.lossMaxTTL > 0 {
			gapSize := p.Header().PacketSequenceNumber.Distance(ackSequenceNumber.Inc())
			if gapSize <= r.lossMaxTTL {
				ackSequenceNumber = p.Header().PacketSequenceNumber
				maxPktTsbpdTime = p.Header().PktTsbpdTime
				r.statistics.MsBuf = (maxPktTsbpdTime - minPktTsbpdTime) / 1_000
				continue
			}
		}

		break
	}

	ok = true
	sequenceNumber = ackSequenceNumber.Inc()

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

	ackSequenceNumber := r.lastACKSequenceNumber

	for e := r.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		if p.Header().PacketSequenceNumber.Lte(ackSequenceNumber) {
			continue
		}

		if !p.Header().PacketSequenceNumber.Equals(ackSequenceNumber.Inc()) {
			gapStart := ackSequenceNumber.Inc()
			gapEnd := p.Header().PacketSequenceNumber.Dec()
			gapSize := gapEnd.Distance(gapStart)

			// Suppress NAK for gaps within lossMaxTTL (bonding tolerance)
			if r.lossMaxTTL > 0 && gapSize <= r.lossMaxTTL {
				ackSequenceNumber = p.Header().PacketSequenceNumber
				continue
			}

			list = append(list, gapStart)
			list = append(list, gapEnd)
		}

		ackSequenceNumber = p.Header().PacketSequenceNumber
	}

	r.lastPeriodicNAK = now

	return list
}

// tryDeliver attempts to deliver packets from packetList in strict sequence order.
// Only delivers if the next expected sequence is available and its TSBPD time has expired.
// Returns true if any packet was delivered.
func (r *receiver) tryDeliver(now uint64) bool {
	delivered := false

	for {
		// Check if nextDeliverySeq is in packetList
		found := false
		gapDetected := false
		for e := r.packetList.Front(); e != nil; e = e.Next() {
			p := e.Value.(packet.Packet)
			if p.Header().PacketSequenceNumber == r.nextDeliverySeq {
				// Found the next packet to deliver
				found = true
				if p.Header().PktTsbpdTime <= now {
					r.packetList.Remove(e)
					r.statistics.PktBuf--
					r.statistics.ByteBuf -= p.Len()
					r.lastDeliveredSequenceNumber = r.nextDeliverySeq
					r.deliver(p)
					r.nextDeliverySeq = r.nextDeliverySeq.Inc()
					r.lastDeliveryAdvance = time.Now()
					delivered = true
					// Continue trying to deliver consecutive packets
					continue
				}
				// TSBPD time not yet expired, wait
				return delivered
			}
			if p.Header().PacketSequenceNumber.Gt(r.nextDeliverySeq) {
				// Gap: nextDeliverySeq is not in the list
				gapDetected = true
				break
			}
		}

		if !found && gapDetected {
			// nextDeliverySeq not in packetList, but packets exist ahead
			// Check stale timeout: if we haven't advanced in a while, skip the gap
			elapsed := time.Since(r.lastDeliveryAdvance).Milliseconds()
			if int(elapsed) >= r.deliveryStaleMs {
				// Skip the gap: advance nextDeliverySeq to the first available packet
				front := r.packetList.Front().Value.(packet.Packet)
				r.nextDeliverySeq = front.Header().PacketSequenceNumber
				r.lastDeliveryAdvance = time.Now()
				// Don't deliver yet, let the next iteration handle it
				continue
			}
			// Not stale yet, wait for the missing packet
			break
		}

		if !found && !gapDetected {
			// packetList is empty or all packets have seq < nextDeliverySeq
			// Check stale timeout for empty list case
			elapsed := time.Since(r.lastDeliveryAdvance).Milliseconds()
			if int(elapsed) >= r.deliveryStaleMs && r.packetList.Len() == 0 {
				// List is empty and stale, nothing to do
			}
			break
		}
	}

	return delivered
}

func (r *receiver) Tick(now uint64) {
	if ok, sequenceNumber, lite := r.periodicACK(now); ok {
		r.sendACK(sequenceNumber, lite)
	}

	if list := r.periodicNAK(now); len(list) != 0 {
		r.sendNAK(list)
	}

	// REORDER BUFFER DELIVERY: deliver packets in strict sequence order
	if r.lossMaxTTL > 0 {
		r.lock.Lock()
		r.tryDeliver(now)
		r.lock.Unlock()
	} else {
		// ORIGINAL DELIVERY (no bonding): deliver in list order
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
	}

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

	b.WriteString(fmt.Sprintf("maxSeen=%d lastACK=%d lastDelivered=%d nextDelivery=%d\n",
		r.maxSeenSequenceNumber.Val(), r.lastACKSequenceNumber.Val(),
		r.lastDeliveredSequenceNumber.Val(), r.nextDeliverySeq.Val()))

	r.lock.RLock()
	for e := r.packetList.Front(); e != nil; e = e.Next() {
		p := e.Value.(packet.Packet)

		b.WriteString(fmt.Sprintf("   %d @ %d (in %d)\n", p.Header().PacketSequenceNumber.Val(), p.Header().PktTsbpdTime, int64(p.Header().PktTsbpdTime)-int64(t)))
	}
	r.lock.RUnlock()

	return b.String()
}
