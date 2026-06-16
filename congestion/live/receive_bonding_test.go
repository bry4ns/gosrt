package live

import (
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/datarhei/gosrt/circular"
	"github.com/datarhei/gosrt/packet"
	"github.com/stretchr/testify/require"
)

func makePacket(seq uint32, tsbpdTime uint64) packet.Packet {
	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")
	p := packet.NewPacket(addr)
	p.Header().PacketSequenceNumber = circular.New(seq, packet.MAX_SEQUENCENUMBER)
	p.Header().PktTsbpdTime = tsbpdTime
	return p
}

func TestBonding2SIMInterleaved(t *testing.T) {
	var delivered []uint32
	var nakCount int32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnSendNAK: func(list []circular.Number) {
			atomic.AddInt32(&nakCount, 1)
		},
		OnDeliver: func(p packet.Packet) {
			mu.Lock()
			delivered = append(delivered, p.Header().PacketSequenceNumber.Val())
			mu.Unlock()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	for i := uint32(0); i < 20; i++ {
		var tsbpd uint64
		if i%2 == 0 {
			tsbpd = uint64(i/2 + 1)
		} else {
			tsbpd = uint64(i/2 + 1 + 50)
		}
		recv.Push(makePacket(i, tsbpd))
	}

	require.Equal(t, uint32(19), recv.maxSeenSequenceNumber.Val())

	for tick := uint64(1); tick <= 200; tick++ {
		recv.Tick(tick)
	}

	require.Equal(t, 20, len(delivered))
	for i := uint32(0); i < 20; i++ {
		require.Equal(t, i, delivered[i])
	}
	require.Equal(t, int32(0), atomic.LoadInt32(&nakCount))
}

func TestBondingHighReorderRate(t *testing.T) {
	var delivered []uint32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			mu.Lock()
			delivered = append(delivered, p.Header().PacketSequenceNumber.Val())
			mu.Unlock()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	packets := make([]packet.Packet, 100)
	for i := uint32(0); i < 100; i++ {
		packets[i] = makePacket(i, uint64(i+1))
	}
	r := rand.New(rand.NewSource(42))
	r.Shuffle(100, func(i, j int) { packets[i], packets[j] = packets[j], packets[i] })

	for _, p := range packets {
		recv.Push(p)
	}

	recv.Tick(200)

	require.Equal(t, 100, len(delivered))
	for i := uint32(0); i < 100; i++ {
		require.Equal(t, i, delivered[i])
	}
}

func TestBondingNAKWithLossMaxTTL(t *testing.T) {
	var nakRanges [][2]uint32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnSendNAK: func(list []circular.Number) {
			mu.Lock()
			for i := 0; i < len(list); i += 2 {
				nakRanges = append(nakRanges, [2]uint32{list[i].Val(), list[i+1].Val()})
			}
			mu.Unlock()
		},
		LossMaxTTL: 30,
	}).(*receiver)

	for i := uint32(0); i < 10; i++ {
		if i == 5 || i == 6 {
			continue
		}
		recv.Push(makePacket(i, uint64(i+1)))
	}

	recv.Push(makePacket(50, 51))

	mu.Lock()
	require.NotEmpty(t, nakRanges)
	found := false
	for _, r := range nakRanges {
		if r[0] == 10 && r[1] == 49 {
			found = true
			break
		}
	}
	mu.Unlock()
	require.True(t, found, "expected NAK [10,49], got %v", nakRanges)
}

func TestBondingACKThroughGaps(t *testing.T) {
	var lastACK uint32

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnSendACK: func(seq circular.Number, light bool) {
			lastACK = seq.Val()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	for i := uint32(0); i < 10; i++ {
		if i == 5 || i == 6 {
			continue
		}
		recv.Push(makePacket(i, uint64(i+1)))
	}

	recv.Tick(10)
	require.Equal(t, uint32(10), lastACK)
}

func TestBondingDuplicates(t *testing.T) {
	var deliveredCount int32

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			atomic.AddInt32(&deliveredCount, 1)
		},
		LossMaxTTL: 100,
	}).(*receiver)

	recv.Push(makePacket(0, 1))
	recv.Push(makePacket(0, 1))
	recv.Push(makePacket(0, 1))

	recv.Tick(10)
	require.Equal(t, int32(1), atomic.LoadInt32(&deliveredCount))
}

func TestBondingBelatedPacket(t *testing.T) {
	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		LossMaxTTL:            100,
	}).(*receiver)

	recv.Push(makePacket(0, 1))
	recv.Tick(10)
	recv.Push(makePacket(0, 1))

	stats := recv.Stats()
	require.Equal(t, uint64(1), stats.PktBelated)
}

func TestBondingSequenceWrapAround(t *testing.T) {
	var delivered []uint32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(packet.MAX_SEQUENCENUMBER-5, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			mu.Lock()
			delivered = append(delivered, p.Header().PacketSequenceNumber.Val())
			mu.Unlock()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	seqs := []uint32{
		packet.MAX_SEQUENCENUMBER - 3,
		packet.MAX_SEQUENCENUMBER - 2,
		packet.MAX_SEQUENCENUMBER - 1,
		packet.MAX_SEQUENCENUMBER,
		0, 1, 2,
	}
	for i, seq := range seqs {
		recv.Push(makePacket(seq, uint64(i+1)))
	}

	recv.Tick(100)

	require.Equal(t, 7, len(delivered))
	for i := 0; i < len(delivered); i++ {
		require.Equal(t, seqs[i], delivered[i])
	}
}

func TestBondingLossMaxTTLBoundary(t *testing.T) {
	var nakCount int32

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnSendNAK: func(list []circular.Number) {
			atomic.AddInt32(&nakCount, 1)
		},
		LossMaxTTL: 30,
	}).(*receiver)

	for i := uint32(0); i < 5; i++ {
		recv.Push(makePacket(i, uint64(i+1)))
	}

	recv.Push(makePacket(34, 35))
	require.Equal(t, int32(0), atomic.LoadInt32(&nakCount))

	recv.Push(makePacket(65, 66))
	require.Equal(t, int32(1), atomic.LoadInt32(&nakCount))
}

func TestBondingMemoryBounded(t *testing.T) {
	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		LossMaxTTL:            100,
	}).(*receiver)

	for i := uint32(0); i < 200; i++ {
		recv.Push(makePacket(i, uint64(i+10000)))
	}

	recv.lock.Lock()
	require.Equal(t, 200, len(recv.packetBuf))
	recv.lock.Unlock()

	for tick := uint64(1); tick <= 15000; tick++ {
		recv.Tick(tick)
	}

	recv.lock.Lock()
	require.Equal(t, 0, len(recv.packetBuf))
	recv.lock.Unlock()
}

func TestBondingStress10K(t *testing.T) {
	var delivered []uint32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			mu.Lock()
			delivered = append(delivered, p.Header().PacketSequenceNumber.Val())
			mu.Unlock()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	simLatencies := []uint64{20, 120, 350}
	for i := uint32(0); i < 10000; i++ {
		simIdx := i % 3
		recv.Push(makePacket(i, uint64(i)+simLatencies[simIdx]))
	}

	for tick := uint64(1); tick <= 10500; tick++ {
		recv.Tick(tick)
	}

	require.Equal(t, 10000, len(delivered))
	for i := uint32(0); i < 10000; i++ {
		require.Equal(t, i, delivered[i])
	}
}

func TestBondingConcurrentPush(t *testing.T) {
	var deliveredCount int32

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			atomic.AddInt32(&deliveredCount, 1)
		},
		LossMaxTTL: 100,
	}).(*receiver)

	var wg sync.WaitGroup
	ranges := [][2]uint32{{0, 3333}, {3333, 6666}, {6666, 10000}}
	for _, r := range ranges {
		wg.Add(1)
		go func(start, end uint32) {
			defer wg.Done()
			for i := start; i < end; i++ {
				recv.Push(makePacket(i, uint64(i+1)))
			}
		}(r[0], r[1])
	}
	wg.Wait()

	for tick := uint64(1); tick <= 10100; tick++ {
		recv.Tick(tick)
	}

	count := atomic.LoadInt32(&deliveredCount)
	require.GreaterOrEqual(t, int(count), 9990)
}

func TestBondingStatsAccuracy(t *testing.T) {
	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   100,
		PeriodicNAKInterval:   200,
		LossMaxTTL:            100,
	}).(*receiver)

	for i := uint32(0); i < 50; i++ {
		recv.Push(makePacket(i, uint64(i+1)))
	}

	stats := recv.Stats()
	require.Equal(t, uint64(50), stats.Pkt)
	require.Equal(t, uint64(50), stats.PktUnique)
	require.Equal(t, uint64(0), stats.PktDrop)

	for i := uint32(0); i < 5; i++ {
		recv.Push(makePacket(i, uint64(i+1)))
	}

	stats = recv.Stats()
	require.Equal(t, uint64(55), stats.Pkt)
	require.Equal(t, uint64(50), stats.PktUnique)
	require.Equal(t, uint64(5), stats.PktDrop)
}

func TestBondingSortedDelivery(t *testing.T) {
	var delivered []uint32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			mu.Lock()
			delivered = append(delivered, p.Header().PacketSequenceNumber.Val())
			mu.Unlock()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	seqs := make([]uint32, 200)
	for i := range seqs {
		seqs[i] = uint32(i)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(200, func(i, j int) { seqs[i], seqs[j] = seqs[j], seqs[i] })

	for _, seq := range seqs {
		recv.Push(makePacket(seq, uint64(seq+1)))
	}

	for tick := uint64(1); tick <= 300; tick++ {
		recv.Tick(tick)
	}

	require.Equal(t, 200, len(delivered))
	for i := 1; i < len(delivered); i++ {
		require.Greater(t, delivered[i], delivered[i-1])
	}
}

func TestBondingCellularSwitch(t *testing.T) {
	var delivered []uint32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			mu.Lock()
			delivered = append(delivered, p.Header().PacketSequenceNumber.Val())
			mu.Unlock()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	for i := uint32(0); i < 50; i++ {
		recv.Push(makePacket(i, uint64(i+1)))
	}

	for i := uint32(50); i < 100; i++ {
		jitter := uint64(200 + rand.Intn(300))
		recv.Push(makePacket(i, uint64(i)+jitter))
	}

	for i := uint32(100); i < 150; i++ {
		recv.Push(makePacket(i, uint64(i)+500))
	}

	for tick := uint64(1); tick <= 2000; tick++ {
		recv.Tick(tick)
	}

	require.Equal(t, 150, len(delivered))
	for i := uint32(0); i < 150; i++ {
		require.Equal(t, i, delivered[i])
	}
}

func TestBondingFlushDuringStream(t *testing.T) {
	var delivered []uint32
	var mu sync.Mutex

	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   10,
		PeriodicNAKInterval:   20,
		OnDeliver: func(p packet.Packet) {
			mu.Lock()
			delivered = append(delivered, p.Header().PacketSequenceNumber.Val())
			mu.Unlock()
		},
		LossMaxTTL: 100,
	}).(*receiver)

	for i := uint32(0); i < 20; i++ {
		recv.Push(makePacket(i, uint64(i+1)))
	}

	recv.Flush()
	recv.lock.Lock()
	require.Equal(t, 0, len(recv.packetBuf))
	recv.lock.Unlock()

	for i := uint32(100); i < 120; i++ {
		recv.Push(makePacket(i, uint64(i+1)))
	}

	recv.Tick(200)

	require.Equal(t, 20, len(delivered))
	for i := 0; i < 20; i++ {
		require.Equal(t, uint32(100+i), delivered[i])
	}
}

func TestBondingPushThroughput(t *testing.T) {
	recv := NewReceiver(ReceiveConfig{
		InitialSequenceNumber: circular.New(0, packet.MAX_SEQUENCENUMBER),
		PeriodicACKInterval:   100000,
		PeriodicNAKInterval:   200000,
		LossMaxTTL:            100,
	}).(*receiver)

	addr, _ := net.ResolveIPAddr("ip", "127.0.0.1")
	const N = 100000
	pkts := make([]packet.Packet, N)
	for i := range pkts {
		p := packet.NewPacket(addr)
		p.Header().PacketSequenceNumber = circular.New(uint32(i), packet.MAX_SEQUENCENUMBER)
		p.Header().PktTsbpdTime = uint64(i + 1)
		pkts[i] = p
	}

	start := time.Now()
	for _, p := range pkts {
		recv.Push(p)
	}
	elapsed := time.Since(start)

	pps := float64(N) / elapsed.Seconds()
	t.Logf("Push throughput: %.0f pkts/sec (%.2f ns/pkt)", pps, float64(elapsed.Nanoseconds())/float64(N))
	require.Greater(t, pps, 100000.0)
}
