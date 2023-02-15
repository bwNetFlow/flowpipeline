package pb

import (
	"net"

	oldpb "github.com/bwNetFlow/protobuf/go"
	goflowpb "github.com/netsampler/goflow2/pb"
)

func NewFromOld(msg *oldpb.FlowMessage) *EnrichedFlow {
	return &EnrichedFlow{
		Type:             EnrichedFlow_FlowType(msg.Type),
		TimeReceived:     msg.TimeReceived,
		SequenceNum:      msg.SequenceNum,
		SamplingRate:     msg.SamplingRate,
		FlowDirection:    msg.FlowDirection,
		SamplerAddress:   msg.SamplerAddress,
		TimeFlowStart:    msg.TimeFlowStart,
		TimeFlowEnd:      msg.TimeFlowEnd,
		Bytes:            msg.Bytes,
		Packets:          msg.Packets,
		SrcAddr:          msg.SrcAddr,
		DstAddr:          msg.DstAddr,
		Etype:            msg.Etype,
		Proto:            msg.Proto,
		SrcPort:          msg.SrcPort,
		DstPort:          msg.DstPort,
		InIf:             msg.InIf,
		OutIf:            msg.OutIf,
		SrcMac:           msg.SrcMac,
		DstMac:           msg.DstMac,
		SrcVlan:          msg.SrcVlan,
		DstVlan:          msg.DstVlan,
		VlanId:           msg.VlanId,
		IngressVrfID:     msg.IngressVrfID,
		EgressVrfID:      msg.EgressVrfID,
		IPTos:            msg.IPTos,
		ForwardingStatus: msg.ForwardingStatus,
		IPTTL:            msg.IPTTL,
		TCPFlags:         msg.TCPFlags,
		IcmpType:         msg.IcmpType,
		IcmpCode:         msg.IcmpCode,
		IPv6FlowLabel:    msg.IPv6FlowLabel,
		FragmentId:       msg.FragmentId,
		FragmentOffset:   msg.FragmentOffset,
		BiFlowDirection:  msg.BiFlowDirection,
		SrcAS:            msg.SrcAS,
		DstAS:            msg.DstAS,
		NextHop:          msg.NextHop,
		NextHopAS:        msg.NextHopAS,
		SrcNet:           msg.SrcNet,
		DstNet:           msg.DstNet,
		HasMPLS:          msg.HasMPLS,
		MPLSCount:        msg.MPLSCount,
		MPLS1TTL:         msg.MPLS1TTL,
		MPLS1Label:       msg.MPLS1Label,
		MPLS2TTL:         msg.MPLS2TTL,
		MPLS2Label:       msg.MPLS2Label,
		MPLS3TTL:         msg.MPLS3TTL,
		MPLS3Label:       msg.MPLS3Label,
		MPLSLastTTL:      msg.MPLSLastTTL,
		MPLSLastLabel:    msg.MPLSLastLabel,
		Cid:              msg.Cid,
		CidString:        msg.CidString,
		SrcCid:           msg.SrcCid,
		DstCid:           msg.DstCid,
		Normalized:       EnrichedFlow_NormalizedType(msg.Normalized),
		SrcIfName:        msg.SrcIfName,
		SrcIfDesc:        msg.SrcIfDesc,
		SrcIfSpeed:       msg.SrcIfSpeed,
		DstIfName:        msg.DstIfName,
		DstIfDesc:        msg.DstIfDesc,
		DstIfSpeed:       msg.DstIfSpeed,
		ProtoName:        msg.ProtoName,
		RemoteCountry:    msg.RemoteCountry,
		SrcCountry:       msg.SrcCountry,
		DstCountry:       msg.DstCountry,
		RemoteAddr:       EnrichedFlow_RemoteAddrType(msg.RemoteAddr),
		Note:             msg.Note,
	}
}

func NewFromGoflow(msg *goflowpb.FlowMessage) *EnrichedFlow {
	return &EnrichedFlow{
		Type:             EnrichedFlow_FlowType(msg.Type),
		TimeReceived:     msg.TimeReceived,
		SequenceNum:      msg.SequenceNum,
		SamplingRate:     msg.SamplingRate,
		FlowDirection:    msg.FlowDirection,
		SamplerAddress:   msg.SamplerAddress,
		TimeFlowStart:    msg.TimeFlowStart,
		TimeFlowEnd:      msg.TimeFlowEnd,
		Bytes:            msg.Bytes,
		Packets:          msg.Packets,
		SrcAddr:          msg.SrcAddr,
		DstAddr:          msg.DstAddr,
		Etype:            msg.Etype,
		Proto:            msg.Proto,
		SrcPort:          msg.SrcPort,
		DstPort:          msg.DstPort,
		InIf:             msg.InIf,
		OutIf:            msg.OutIf,
		SrcMac:           msg.SrcMac,
		DstMac:           msg.DstMac,
		SrcVlan:          msg.SrcVlan,
		DstVlan:          msg.DstVlan,
		VlanId:           msg.VlanId,
		IngressVrfID:     msg.IngressVrfID,
		EgressVrfID:      msg.EgressVrfID,
		IPTos:            msg.IPTos,
		ForwardingStatus: msg.ForwardingStatus,
		IPTTL:            msg.IPTTL,
		TCPFlags:         msg.TCPFlags,
		IcmpType:         msg.IcmpType,
		IcmpCode:         msg.IcmpCode,
		IPv6FlowLabel:    msg.IPv6FlowLabel,
		FragmentId:       msg.FragmentId,
		FragmentOffset:   msg.FragmentOffset,
		BiFlowDirection:  msg.BiFlowDirection,
		SrcAS:            msg.SrcAS,
		DstAS:            msg.DstAS,
		NextHop:          msg.NextHop,
		NextHopAS:        msg.NextHopAS,
		SrcNet:           msg.SrcNet,
		DstNet:           msg.DstNet,
		HasMPLS:          msg.HasMPLS,
		MPLSCount:        msg.MPLSCount,
		MPLS1TTL:         msg.MPLS1TTL,
		MPLS1Label:       msg.MPLS1Label,
		MPLS2TTL:         msg.MPLS2TTL,
		MPLS2Label:       msg.MPLS2Label,
		MPLS3TTL:         msg.MPLS3TTL,
		MPLS3Label:       msg.MPLS3Label,
		MPLSLastTTL:      msg.MPLSLastTTL,
		MPLSLastLabel:    msg.MPLSLastLabel,
		// leave the 'Custom*' fields out of this conversion
	}
}

var (
	FlowDirectionMap = map[uint32]string{
		0: "Incoming",
		1: "Outgoing"}
	EtypeMap = map[uint32]string{
		0x0800: "IPv4",
		0x0806: "ARP",
		0x0842: "Wake-on-LAN",
		0x86DD: "IPv6",
		0x8809: "Ethernet Slow Protocols (LACP)",
		0x8847: "MPLS unicast",
		0x8848: "MPLS multicast",
		0x8863: "PPPoE Discovery Stage",
		0x8864: "PPPoE Session Stage",
		0x889A: "HyperSCSI (SCSI over Ethernet)",
		0x88A2: "ATA over Ethernet",
		0x88A4: "EtherCAT Protocol",
		0x88CC: "LLDP",
		0x88E5: "MAC Security",
		0x8906: "Fibre Channel over Ethernet (FCoE)",
		0x8914: "FCoE Initialization Protocol",
		0x9000: "Ethernet Configuration Testing Protocol"}
	ForwardingStatusMap = map[uint32]string{
		0:   "Unknown",
		64:  "Forwarded (Unknown)",
		65:  "Forwarded (Fragmented)",
		66:  "Forwarded (Not Fragmented)",
		128: "Dropped (Unknown)",
		129: "Dropped (ACL Deny)",
		130: "Dropped (ACL Drop)",
		131: "Dropped (Unroutable)",
		132: "Dropped (Adjacency)",
		133: "Dropped (Fragmented and DF set)",
		134: "Dropped (Bad Header Checksum)",
		135: "Dropped (Bad Total Length)",
		136: "Dropped (Bad Header Length)",
		137: "Dropped (Bad TTL)",
		138: "Dropped (Policer)",
		139: "Dropped (WRED)",
		140: "Dropped (RPF)",
		141: "Dropped (For Us)",
		142: "Dropped (Bad Output Interface)",
		143: "Dropped (Hardware)",
		192: "Consumed (Unknown)",
		193: "Consumed (Terminate Punt Adjacency)",
		194: "Consumed (Terminate Incomplete Adjacency)",
		195: "Consumed (Terminate For Us)"}
)

func (flow *EnrichedFlow) FlowDirectionString() string {
	return FlowDirectionMap[flow.GetFlowDirection()]
}

func (flow *EnrichedFlow) IsIncoming() bool {
	return flow.GetFlowDirection() == 0
}

func (flow *EnrichedFlow) IsOutgoing() bool {
	return flow.GetFlowDirection() == 1
}

func (flow *EnrichedFlow) Peer() string {
	switch flow.GetFlowDirection() {
	case 0:
		return flow.GetSrcIfDesc()
	case 1:
		return flow.GetDstIfDesc()
	default:
		return ""
	}
}

func (flow *EnrichedFlow) EtypeString() string {
	return EtypeMap[flow.GetEtype()]
}

func (flow *EnrichedFlow) IPVersion() uint8 {
	switch flow.GetEtype() {
	case 0x0800:
		return 4
	case 0x86dd:
		return 6
	default:
		return 0
	}
}

func (flow *EnrichedFlow) IPVersionString() string {
	if flow.GetEtype() == 0x800 || flow.GetEtype() == 0x86dd {
		return EtypeMap[flow.GetEtype()]
	} else {
		return ""
	}
}

func (flow *EnrichedFlow) IsIPv4() bool {
	return flow.GetEtype() == 0x0800
}

func (flow *EnrichedFlow) IsIPv6() bool {
	return flow.GetEtype() == 0x86dd
}

func (flow *EnrichedFlow) ForwardingStatusString() string {
	return ForwardingStatusMap[flow.GetForwardingStatus()]
}

func (flow *EnrichedFlow) IsConsumed() bool {
	return 192 <= flow.GetForwardingStatus() // && < 256
}

func (flow *EnrichedFlow) IsDropped() bool {
	return 128 <= flow.GetForwardingStatus() && flow.GetForwardingStatus() < 192
}

func (flow *EnrichedFlow) IsForwarded() bool {
	return 64 <= flow.GetForwardingStatus() && flow.GetForwardingStatus() < 128
}

func (flow *EnrichedFlow) IsUnknownForwardingStatus() bool {
	return flow.GetForwardingStatus() < 64
}

func (flow *EnrichedFlow) SrcAddrObj() net.IP {
	return net.IP(flow.SrcAddr)
}

func (flow *EnrichedFlow) DstAddrObj() net.IP {
	return net.IP(flow.DstAddr)
}

func (flow *EnrichedFlow) SamplerAddressObj() net.IP {
	return net.IP(flow.SamplerAddress)
}

func (flow *EnrichedFlow) NextHopObj() net.IP {
	return net.IP(flow.NextHop)
}

func (flow *EnrichedFlow) GetBps() uint64 {
	duration := flow.TimeFlowEnd - flow.TimeFlowStart
	if duration == 0 {
		duration = 1
	}
	return flow.Bytes / duration
}

func (flow *EnrichedFlow) GetPps() uint64 {
	duration := flow.TimeFlowEnd - flow.TimeFlowStart
	if duration == 0 {
		duration = 1
	}
	return flow.Packets / duration
}
