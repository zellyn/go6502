package visual

import (
	icpu "github.com/zellyn/go6502/cpu" // Just need the interface
)

type cpu struct {
	m              icpu.Memory
	cycle          uint64
	nodeValues     []byte   // Bitmask of node values (see const VAL_* below)
	nodeGates      [][]uint // the list of transistor indexes attached to a node
	nodeC1C2s      [][]uint // the list of transistor c1/c2s attached to a node
	nodeDependants [][]uint // all C1 and C2 nodes of transistors attached to a node

	transistorValues []bool

	listIn  []uint
	listOut []uint

	groupList  []uint               // list of node group membership
	groupSet   [NODES/32 + 1]uint32 // quick check for node group membership
	groupValue byte                 // presence of vss/vcc/pulldown/pullup/hi in group
}

// Bitfield for node values.
const (
	VAL_HI = 1 << iota // We count on this being bit 0, so we can mask it out for 0 or 1.
	VAL_PULLUP
	VAL_PULLDOWN
	VAL_VCC
	VAL_VSS
)

// The lookup table for the group value. If vss is in the group, it's 0, vcc makes it 1, etc.
// vss, vcc, pulldown, pullup, hi
var GroupValues = [32]byte{
	0, // 00000 - nothing
	1, // 00001 - contains at least one hi node
	1, // 00010 - contains at least one pullup
	1, // 00011 - contains at least one pullup
	0, // 00100 - contains at least one pulldown
	0, // 00101 - contains at least one pulldown
	0, // 00110 - contains at least one pulldown
	0, // 00111 - contains at least one pulldown
	1, // 01000 - contains vcc
	1, // 01001 - contains vcc
	1, // 01010 - contains vcc
	1, // 01011 - contains vcc
	1, // 01100 - contains vcc
	1, // 01101 - contains vcc
	1, // 01110 - contains vcc
	1, // 01111 - contains vcc
	0, // 10000- contains vss
	0, // 10001- contains vss
	0, // 10010- contains vss
	0, // 10011- contains vss
	0, // 10100- contains vss
	0, // 10101- contains vss
	0, // 10110- contains vss
	0, // 10111- contains vss
	0, // 11000- contains vss
	0, // 11001- contains vss
	0, // 11010- contains vss
	0, // 11011- contains vss
	0, // 11100- contains vss
	0, // 11101- contains vss
	0, // 11110- contains vss
	0, // 11111- contains vss
}

func NewCPU(memory icpu.Memory) icpu.Cpu {
	c := cpu{m: memory}
	c.setupNodesAndTransistors()
	return &c
}

// Needed for the interface. Not really practical. I guess we could try changing the nodes directly.
func (c *cpu) SetPC(uint16) {
	panic("Not implemented")
}

// --------------------------------
// Interfacing and extracting state
// --------------------------------

func (c *cpu) Read8(n0, n1, n2, n3, n4, n5, n6, n7 uint) byte {
	return (c.nodeBit(n0) | c.nodeBit(n1)<<1 | c.nodeBit(n2)<<2 | c.nodeBit(n3)<<3 |
		c.nodeBit(n4)<<4 | c.nodeBit(n5)<<5 | c.nodeBit(n6)<<6 | c.nodeBit(n7)<<7)
}

func (c *cpu) AddressBus() uint16 {
	abl := uint16(c.Read8(NODE_ab0, NODE_ab1, NODE_ab2, NODE_ab3, NODE_ab4, NODE_ab5, NODE_ab6, NODE_ab7))
	abh := uint16(c.Read8(NODE_ab8, NODE_ab9, NODE_ab10, NODE_ab11, NODE_ab12, NODE_ab13, NODE_ab14, NODE_ab15))
	return abl + abh<<8
}

func (c *cpu) DataBus() byte {
	return c.Read8(NODE_db0, NODE_db1, NODE_db2, NODE_db3, NODE_db4, NODE_db5, NODE_db6, NODE_db7)
}

func (c *cpu) A() byte {
	return c.Read8(NODE_a0, NODE_a1, NODE_a2, NODE_a3, NODE_a4, NODE_a5, NODE_a6, NODE_a7)
}

func (c *cpu) X() byte {
	return c.Read8(NODE_x0, NODE_x1, NODE_x2, NODE_x3, NODE_x4, NODE_x5, NODE_x6, NODE_x7)
}

func (c *cpu) Y() byte {
	return c.Read8(NODE_y0, NODE_y1, NODE_y2, NODE_y3, NODE_y4, NODE_y5, NODE_y6, NODE_y7)
}

func (c *cpu) P() byte {
	return c.Read8(NODE_p0, NODE_p1, NODE_p2, NODE_p3, NODE_p4, NODE_p5, NODE_p6, NODE_p7)
}

func (c *cpu) SP() byte {
	return c.Read8(NODE_s0, NODE_s1, NODE_s2, NODE_s3, NODE_s4, NODE_s5, NODE_s6, NODE_s7)
}

func (c *cpu) IR() byte {
	return c.Read8(NODE_notir0, NODE_notir1, NODE_notir2, NODE_notir3, NODE_notir4,
		NODE_notir5, NODE_notir6, NODE_notir7) ^ 0xFF
}

func (c *cpu) PCL() byte {
	return c.Read8(NODE_pcl0, NODE_pcl1, NODE_pcl2, NODE_pcl3, NODE_pcl4, NODE_pcl5, NODE_pcl6, NODE_pcl7)
}

func (c *cpu) PCH() byte {
	return c.Read8(NODE_pch0, NODE_pch1, NODE_pch2, NODE_pch3, NODE_pch4, NODE_pch5, NODE_pch6, NODE_pch7)
}

func (c *cpu) PC() uint16 {
	return uint16(c.PCH())<<8 + uint16(c.PCL())
}

func (c *cpu) nodeBit(n uint) byte {
	return c.nodeValues[n] & VAL_HI // 1
}

func (c *cpu) writeDataBus(d byte) {
	for i := 0; i < 8; i++ {
		c.setNode(DataBusNodes[i], d&1 == 1)
		d >>= 1
	}
}

func (c *cpu) Reset() {
	// All nodes down
	for i := range c.nodeValues {
		c.nodeValues[i] &^= VAL_HI
	}

	// All transistors off

	for i := range c.transistorValues {
		c.transistorValues[i] = false
	}

	c.setNode(NODE_res, false)
	c.setNode(NODE_clk0, true)
	c.setNode(NODE_rdy, true)
	c.setNode(NODE_so, false)
	c.setNode(NODE_irq, true)
	c.setNode(NODE_nmi, true)

	c.recalcAllNodes()

	// Hold RESET for 8 cycles
	for i := 0; i < 8; i++ {
		c.Step()
	}

	c.setNode(NODE_res, true)

	c.cycle = 0
}

func (c *cpu) switchLists() {
	c.listIn, c.listOut = c.listOut, c.listIn
}

func (c *cpu) addNodeToGroup(n uint) {
	index := n >> 5
	mask := uint32(1 << (n & 0x1f))
	if c.groupSet[index]&mask > 0 {
		return
	}

	c.groupSet[index] |= mask
	c.groupList = append(c.groupList, n)

	c.groupValue |= c.nodeValues[n]
	if n == NODE_vss || n == NODE_vcc {
		return
	}

	/* revisit all transistors that are controlled by this node */
	for _, tn := range c.nodeC1C2s[n] {
		if c.transistorValues[tn] {
			if TransDefs[tn].c1 == n {
				c.addNodeToGroup(TransDefs[tn].c2)
			} else {
				c.addNodeToGroup(TransDefs[tn].c1)
			}
		}
	}
}

func (c *cpu) addAllNodesToGroup(node uint) {
	c.groupList = c.groupList[0:0]
	c.groupValue = 0

	c.addNodeToGroup(node)
}

func (c *cpu) recalcNode(node uint) {
	/*
	 * get all nodes that are connected through
	 * transistors, starting with this one
	 */
	c.addAllNodesToGroup(node)

	/* get the state of the group */
	newv := GroupValues[c.groupValue]

	/*
	 * - set all nodes to the group state
	 * - check all transistors switched by nodes of the group
	 * - collect all nodes behind toggled transistors
	 *   for the next run
	 */
	for _, nn := range c.groupList {
		c.groupSet[nn>>5] = 0 // Clear as we go
		if c.nodeValues[nn]&VAL_HI != newv {
			c.nodeValues[nn] ^= VAL_HI
			for _, tn := range c.nodeGates[nn] {
				c.transistorValues[tn] = !c.transistorValues[tn]
			}
			c.listOut = append(c.listOut, nn)
		}
	}
}

func (c *cpu) recalcNodeList(nodes []uint) {
	c.listOut = c.listOut[0:0]

	for _, n := range nodes {
		c.recalcNode(n)
	}

	c.switchLists()

	for j := 0; j < 100; j++ { /* loop limiter */
		if len(c.listIn) == 0 {
			break
		}
		c.listOut = c.listOut[0:0]

		/*
		 * for all nodes, follow their paths through
		 * turned-on transistors, find the state of the
		 * path and assign it to all nodes, and re-evaluate
		 * all transistors controlled by this path, collecting
		 * all nodes that changed because of it for the next run
		 */
		for _, n := range c.listIn {
			for _, d := range c.nodeDependants[n] {
				c.recalcNode(d)
			}
		}
		/*
		 * make the secondary list our primary list, use
		 * the data storage of the primary list as the
		 * secondary list
		 */
		c.switchLists()
	}
}

func (c *cpu) recalcAllNodes() {
	temp := make([]uint, NODES)
	for i := uint(0); i < NODES; i++ {
		temp[i] = i
	}
	c.recalcNodeList(temp)
}

/**************/
/* Node State */
/**************/

// So we don't have to keep re-allocating
var oneNode = []uint{0}

func (c *cpu) setNode(nn uint, state bool) {
	oldState := c.nodeValues[nn]
	newState := oldState
	if state {
		newState &^= VAL_PULLDOWN
		newState |= VAL_PULLUP
	} else {
		newState &^= VAL_PULLUP
		newState |= VAL_PULLDOWN
	}
	if newState != oldState {
		c.nodeValues[nn] = newState
		oneNode[0] = nn
		c.recalcNodeList(oneNode)
	}
}

func (c *cpu) isNodeHigh(n uint) bool {
	return c.nodeValues[n]&VAL_HI > 0
}

// handleMemory is called when clk0 is low, and either reads from or
// writes to memory, depending on rw.
func (c *cpu) handleMemory() {
	if c.isNodeHigh(NODE_rw) {
		c.writeDataBus(c.m.Read(c.AddressBus()))
	} else {
		c.m.Write(c.AddressBus(), c.DataBus())
	}
}

// HalfStep is the main clock loop, and takes a half clock step.
func (c *cpu) HalfStep() {
	clk := c.isNodeHigh(NODE_clk0)
	c.setNode(NODE_clk0, !clk)

	if !clk {
		c.handleMemory()
	}
	c.cycle++
}

// Step takes two half steps.
func (c *cpu) Step() error {
	c.HalfStep()
	c.HalfStep()
	return nil
}

/******************/
/* Initialization */
/******************/

func (c *cpu) addNodeDependant(a, b uint) {
	for _, d := range c.nodeDependants[a] {
		if b == d {
			return
		}
	}
	c.nodeDependants[a] = append(c.nodeDependants[a], b)
}

func (c *cpu) setupNodesAndTransistors() {

	// Zero out bitsets
	c.transistorValues = make([]bool, TRANSISTORS)
	c.groupList = make([]uint, 0, NODES)
	c.nodeValues = make([]byte, NODES)
	c.nodeGates = make([][]uint, NODES)
	c.nodeC1C2s = make([][]uint, NODES)
	c.nodeDependants = make([][]uint, NODES)

	// Copy node data from SegDefs into r/w data structures
	for i := uint(0); i < NODES; i++ {
		if SegDefs[i] {
			c.nodeValues[i] = VAL_PULLUP
		}
		if i == NODE_vss {
			c.nodeValues[i] |= VAL_VSS
		}
		if i == NODE_vcc {
			c.nodeValues[i] |= VAL_VCC
		}
	}

	// Cross-reference transistors in nodes data structures
	for j, t := range TransDefs {
		i := uint(j)
		c.nodeGates[t.gate] = append(c.nodeGates[t.gate], i)
		c.nodeC1C2s[t.c1] = append(c.nodeC1C2s[t.c1], i)
		c.nodeC1C2s[t.c2] = append(c.nodeC1C2s[t.c2], i)
	}

	for i := uint(0); i < NODES; i++ {
		for _, t := range c.nodeGates[i] {
			c.addNodeDependant(i, TransDefs[t].c1)
			c.addNodeDependant(i, TransDefs[t].c2)
		}
	}
}
