package visual

import (
	icpu "github.com/zellyn/go6502/cpu" // Just need the interface
)

type cpu struct {
	m     icpu.Memory
	cycle uint64

	nodes       uint // number of nodes
	transistors uint // number of transistors

	vss uint
	vcc uint

	nodesPullup        bitmap
	nodesPulldown      bitmap
	nodesValue         bitmap
	nodeGates          [][]uint // the list of transistor indexes attached to a node
	nodeC1C2s          [][]uint // the list of transistor c1/c2s attached to a node
	nodeDependents     [][]uint // all C1 and C2 nodes of transistors attached to a node
	nodeLeftDependents [][]uint // TODO(zellyn): doc

	transistorsGate []uint
	transistorsC1   []uint
	transistorsC2   []uint
	transistorsOn   bitmap

	listIn        []uint // the nodes we are working with
	listOut       []uint // the indirect nodes we are collecting for the next run
	listOutBitmap bitmap

	group              []uint
	groupBitmap        bitmap
	groupContainsValue groupContains
}

type groupContains uint8

const (
	CONTAINS_NOTHING groupContains = iota
	CONTAINS_HI
	CONTAINS_PULLUP
	CONTAINS_PULLDOWN
	CONTAINS_VCC
	CONTAINS_VSS
)

// // The lookup table for the group value. If vss is in the group, it's 0, vcc makes it 1, etc.
// // vss, vcc, pulldown, pullup, hi
// var GroupValues = [32]byte{
// 	0, // 00000 - nothing
// 	1, // 00001 - contains at least one hi node
// 	1, // 00010 - contains at least one pullup
// 	1, // 00011 - contains at least one pullup
// 	0, // 00100 - contains at least one pulldown
// 	0, // 00101 - contains at least one pulldown
// 	0, // 00110 - contains at least one pulldown
// 	0, // 00111 - contains at least one pulldown
// 	1, // 01000 - contains vcc
// 	1, // 01001 - contains vcc
// 	1, // 01010 - contains vcc
// 	1, // 01011 - contains vcc
// 	1, // 01100 - contains vcc
// 	1, // 01101 - contains vcc
// 	1, // 01110 - contains vcc
// 	1, // 01111 - contains vcc
// 	0, // 10000- contains vss
// 	0, // 10001- contains vss
// 	0, // 10010- contains vss
// 	0, // 10011- contains vss
// 	0, // 10100- contains vss
// 	0, // 10101- contains vss
// 	0, // 10110- contains vss
// 	0, // 10111- contains vss
// 	0, // 11000- contains vss
// 	0, // 11001- contains vss
// 	0, // 11010- contains vss
// 	0, // 11011- contains vss
// 	0, // 11100- contains vss
// 	0, // 11101- contains vss
// 	0, // 11110- contains vss
// 	0, // 11111- contains vss
// }

func NewCPU(memory icpu.Memory) icpu.Cpu {
	c := cpu{m: memory}
	c.setupNodesAndTransistors(TransDefs, SegDefs, NODE_vss, NODE_vcc)
	return &c
}

// Needed for the interface. Not really practical. I guess we could try changing the nodes directly.
func (c *cpu) SetPC(uint16) {
	panic("Not implemented")
}

/************************************/
/* Interfacing and extracting state */
/************************************/

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
	if c.getNodeValue(n) {
		return 1
	}
	return 0
}

func (c *cpu) writeDataBus(d byte) {
	for i := 0; i < 8; i++ {
		c.setNode(DataBusNodes[i], d&1 == 1)
		d >>= 1
	}
}

func (c *cpu) Reset() {
	c.setNode(NODE_res, false)
	c.setNode(NODE_clk0, true)
	c.setNode(NODE_rdy, true)
	c.setNode(NODE_so, false)
	c.setNode(NODE_irq, true)
	c.setNode(NODE_nmi, true)

	c.stabilizeChip()

	// Hold RESET for 8 cycles
	for i := 0; i < 8; i++ {
		c.Step()
	}

	c.setNode(NODE_res, true)
	c.recalcNodeList()

	c.cycle = 0
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
	/* invert clock */
	c.setNode(NODE_clk0, !clk)
	c.recalcNodeList()

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

/************************/
/* Algorithms for Nodes */
/************************/

/*
 * The "value" propertiy of VCC and GND is never evaluated in the code,
 * so we don't bother initializing it properly or special-casing writes.
 */

func (c *cpu) setNodePullup(t uint, s bool) {
	c.nodesPullup.set(t, s)
}

func (c *cpu) getNodePullup(t uint) bool {
	return c.nodesPullup.get(t)
}

func (c *cpu) setNodePulldown(t uint, s bool) {
	c.nodesPulldown.set(t, s)
}

func (c *cpu) getNodePulldown(t uint) bool {
	return c.nodesPulldown.get(t)
}

func (c *cpu) setNodeValue(t uint, s bool) {
	c.nodesValue.set(t, s)
}

func (c *cpu) getNodeValue(t uint) bool {
	return c.nodesValue.get(t)
}

/******************************/
/* Algorithms for Transistors */
/******************************/

func (c *cpu) setTransistorOn(t uint, s bool) {
	c.transistorsOn.set(t, s)
}

func (c *cpu) getTransistorOn(t uint) bool {
	return c.transistorsOn.get(t)
}

/************************/
/* Algorithms for Lists */
/************************/

func (c *cpu) switchLists() {
	c.listIn, c.listOut = c.listOut, c.listIn
}

func (c *cpu) clearListOut() {
	c.listOut = c.listOut[:0]
	c.listOutBitmap.clear()
}

func (c *cpu) listOutAdd(i uint) {
	if !c.listOutBitmap.get(i) {
		c.listOut = append(c.listOut, i)
		c.listOutBitmap.set(i, true)
	}
}

/**********************************/
/* Algorithms for Groups of Nodes */
/**********************************/

/*
 * a group is a set of connected nodes, which consequently
 * share the same value
 *
 * we use an array and a count for O(1) insert and
 * iteration, and a redundant bitmap for O(1) lookup
 */

func (c *cpu) groupClear() {
	c.group = c.group[:0]
	c.groupBitmap.clear()
}

func (c *cpu) groupAdd(i uint) {
	c.group = append(c.group, i)
	c.groupBitmap.set(i, true)
}

func (c *cpu) groupContains(el uint) bool {
	return c.groupBitmap.get(el)
}

func (c *cpu) groupCount() uint {
	return uint(len(c.group))
}

/*********************************/
/* Node and Transistor Emulation */
/*********************************/

func (c *cpu) addNodeToGroup(n uint) {
	/*
	 * We need to stop at vss and vcc, otherwise we'll revisit other groups
	 * with the same value - just because they all derive their value from
	 * the fact that they are connected to vcc or vss.
	 */

	if n == c.vss {
		c.groupContainsValue = CONTAINS_VSS
		return
	}

	if n == c.vcc {
		if c.groupContainsValue != CONTAINS_VSS {
			c.groupContainsValue = CONTAINS_VCC
		}
		return
	}

	if c.groupContains(n) {
		return
	}

	c.groupAdd(n)

	if c.groupContainsValue < CONTAINS_PULLDOWN && c.getNodePulldown(n) {
		c.groupContainsValue = CONTAINS_PULLDOWN
	}

	if c.groupContainsValue < CONTAINS_PULLUP && c.getNodePullup(n) {
		c.groupContainsValue = CONTAINS_PULLUP
	}

	if c.groupContainsValue < CONTAINS_HI && c.getNodeValue(n) {
		c.groupContainsValue = CONTAINS_HI
	}

	/* revisit all transistors that control this node */
	for _, tn := range c.nodeC1C2s[n] {
		/* if the transistor connects c1 and c2... */
		if c.getTransistorOn(uint(tn)) {
			/* if original node was connected to c1, continue with c2 */
			if c.transistorsC1[tn] == n {
				c.addNodeToGroup(c.transistorsC2[tn])
			} else {
				c.addNodeToGroup(c.transistorsC1[tn])
			}
		}
	}
}

func (c *cpu) addAllNodesToGroup(node uint) {
	c.groupClear()
	c.groupContainsValue = CONTAINS_NOTHING
	c.addNodeToGroup(node)
}

func (c *cpu) getGroupValue() bool {
	switch c.groupContainsValue {
	case CONTAINS_VCC, CONTAINS_PULLUP, CONTAINS_HI:
		return true
	case CONTAINS_VSS, CONTAINS_PULLDOWN, CONTAINS_NOTHING:
		return false
	}
	panic("cannot get here")
}

func (c *cpu) recalcNode(node uint) {
	/*
	 * get all nodes that are connected through
	 * transistors, starting with this one
	 */
	c.addAllNodesToGroup(node)

	/* get the state of the group */
	newv := c.getGroupValue()

	/*
	 * - set all nodes to the group state
	 * - check all transistors switched by nodes of the group
	 * - collect all nodes behind toggled transistors
	 *   for the next run
	 */
	for _, nn := range c.group {
		if c.getNodeValue(nn) != newv {
			c.setNodeValue(nn, newv)
			for _, tn := range c.nodeGates[nn] {
				c.setTransistorOn(tn, newv)
			}

			if newv {
				for _, dp := range c.nodeLeftDependents[nn] {
					c.listOutAdd(dp)
				}
			} else {
				for _, dp := range c.nodeDependents[nn] {
					c.listOutAdd(dp)
				}
			}
		}
	}
}

func (c *cpu) recalcNodeList() {
	for j := 0; j < 100; j++ { /* loop limiter */
		/*
		 * make the secondary list our primary list, use
		 * the data storage of the primary list as the
		 * secondary list
		 */
		c.switchLists()

		if len(c.listIn) == 0 {
			break
		}

		c.clearListOut()

		/*
		 * for all nodes, follow their paths through
		 * turned-on transistors, find the state of the
		 * path and assign it to all nodes, and re-evaluate
		 * all transistors controlled by this path, collecting
		 * all nodes that changed because of it for the next run
		 */
		for _, n := range c.listIn {
			c.recalcNode(n)
		}
	}
	c.clearListOut()
}

/******************/
/* Initialization */
/******************/

func (c *cpu) addNodeDependent(a uint, b uint) {
	for _, dp := range c.nodeDependents[a] {
		if dp == b {
			return
		}
	}
	c.nodeDependents[a] = append(c.nodeDependents[a], b)
}

func (c *cpu) addNodeLeftDependent(a uint, b uint) {
	for _, dp := range c.nodeLeftDependents[a] {
		if dp == b {
			return
		}
	}
	c.nodeLeftDependents[a] = append(c.nodeLeftDependents[a], b)
}

func (c *cpu) setupNodesAndTransistors(transDefs []TransDef, nodeIsPullup []bool, vss uint, vcc uint) {
	c.nodes = uint(len(nodeIsPullup))
	c.transistors = uint(len(transDefs))
	c.vss = vss
	c.vcc = vcc
	c.nodesPullup = newBitmap(c.nodes)
	c.nodesPulldown = newBitmap(c.nodes)
	c.nodesValue = newBitmap(c.nodes)

	c.transistorsOn = newBitmap(c.transistors)
	c.listOutBitmap = newBitmap(c.nodes)
	c.groupBitmap = newBitmap(c.nodes)

	c.nodeGates = make([][]uint, c.transistors)
	c.nodeC1C2s = make([][]uint, c.transistors)
	c.nodeDependents = make([][]uint, c.nodes, c.nodes)
	c.nodeLeftDependents = make([][]uint, c.nodes, c.nodes)

	/* copy nodes into r/w data structure */
	for i, isPullup := range nodeIsPullup {
		c.setNodePullup(uint(i), isPullup)
	}

	/* copy transistors into r/w data structure */
	for _, transDef := range transDefs {
		found := false
		for j2, gate := range c.transistorsGate {
			if gate == transDef.gate &&
				((c.transistorsC1[j2] == transDef.c1 && c.transistorsC2[j2] == transDef.c2) ||
					(c.transistorsC1[j2] == transDef.c2 && c.transistorsC2[j2] == transDef.c1)) {
				found = true
				break
			}
		}
		if !found {
			c.transistorsGate = append(c.transistorsGate, transDef.gate)
			c.transistorsC1 = append(c.transistorsC1, transDef.c1)
			c.transistorsC2 = append(c.transistorsC2, transDef.c2)
		}
	}

	/* cross reference transistors in nodes data structures */
	for i, gate := range c.transistorsGate {
		c1 := c.transistorsC1[i]
		c2 := c.transistorsC2[i]
		c.nodeGates[gate] = append(c.nodeGates[gate], uint(i))
		c.nodeC1C2s[c1] = append(c.nodeC1C2s[c1], uint(i))
		c.nodeC1C2s[c2] = append(c.nodeC1C2s[c2], uint(i))
	}

	for i := uint(0); i < c.nodes; i++ {
		for _, t := range c.nodeGates[i] {
			c1 := c.transistorsC1[t]
			if c1 != vss && c1 != vcc {
				c.addNodeDependent(i, c1)
			}
			c2 := c.transistorsC2[t]
			if c2 != vss && c2 != vcc {
				c.addNodeDependent(i, c2)
			}
			if c1 != vss && c1 != vcc {
				c.addNodeLeftDependent(i, c1)
			} else {
				c.addNodeLeftDependent(i, c2)
			}
		}
	}
}

func (c *cpu) Print(bool) {
	panic("not implemented")
}

func (c *cpu) stabilizeChip() {
	for i := uint(0); i < c.nodes; i++ {
		c.listOutAdd(i)
	}
	c.recalcNodeList()
}

/**************/
/* Node State */
/**************/

func (c *cpu) setNode(nn uint, s bool) {
	c.setNodePullup(nn, s)
	c.setNodePulldown(nn, !s)
	c.listOutAdd(nn)
	c.recalcNodeList()
}

func (c *cpu) isNodeHigh(nn uint) bool {
	return c.getNodeValue(nn)
}
