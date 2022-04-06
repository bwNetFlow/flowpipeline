package pb

import (
	goflowpb "github.com/netsampler/goflow2/pb"
)

func (x *EnrichedFlow) GetType() goflowpb.FlowMessage_FlowType {
	if x != nil {
		return x.Core.Type
	}
	return goflowpb.FlowMessage_FLOWUNKNOWN
}

func (x *EnrichedFlow) GetTimeReceived() uint64 {
	if x != nil {
		return x.Core.TimeReceived
	}
	return 0
}

func (x *EnrichedFlow) GetSequenceNum() uint32 {
	if x != nil {
		return x.Core.SequenceNum
	}
	return 0
}

func (x *EnrichedFlow) GetSamplingRate() uint64 {
	if x != nil {
		return x.Core.SamplingRate
	}
	return 0
}

func (x *EnrichedFlow) GetFlowDirection() uint32 {
	if x != nil {
		return x.Core.FlowDirection
	}
	return 0
}

func (x *EnrichedFlow) GetSamplerAddress() []byte {
	if x != nil {
		return x.Core.SamplerAddress
	}
	return nil
}

func (x *EnrichedFlow) GetTimeFlowStart() uint64 {
	if x != nil {
		return x.Core.TimeFlowStart
	}
	return 0
}

func (x *EnrichedFlow) GetTimeFlowEnd() uint64 {
	if x != nil {
		return x.Core.TimeFlowEnd
	}
	return 0
}

func (x *EnrichedFlow) GetBytes() uint64 {
	if x != nil {
		return x.Core.Bytes
	}
	return 0
}

func (x *EnrichedFlow) GetPackets() uint64 {
	if x != nil {
		return x.Core.Packets
	}
	return 0
}

func (x *EnrichedFlow) GetSrcAddr() []byte {
	if x != nil {
		return x.Core.SrcAddr
	}
	return nil
}

func (x *EnrichedFlow) GetDstAddr() []byte {
	if x != nil {
		return x.Core.DstAddr
	}
	return nil
}

func (x *EnrichedFlow) GetEtype() uint32 {
	if x != nil {
		return x.Core.Etype
	}
	return 0
}

func (x *EnrichedFlow) GetProto() uint32 {
	if x != nil {
		return x.Core.Proto
	}
	return 0
}

func (x *EnrichedFlow) GetSrcPort() uint32 {
	if x != nil {
		return x.Core.SrcPort
	}
	return 0
}

func (x *EnrichedFlow) GetDstPort() uint32 {
	if x != nil {
		return x.Core.DstPort
	}
	return 0
}

func (x *EnrichedFlow) GetInIf() uint32 {
	if x != nil {
		return x.Core.InIf
	}
	return 0
}

func (x *EnrichedFlow) GetOutIf() uint32 {
	if x != nil {
		return x.Core.OutIf
	}
	return 0
}

func (x *EnrichedFlow) GetSrcMac() uint64 {
	if x != nil {
		return x.Core.SrcMac
	}
	return 0
}

func (x *EnrichedFlow) GetDstMac() uint64 {
	if x != nil {
		return x.Core.DstMac
	}
	return 0
}

func (x *EnrichedFlow) GetSrcVlan() uint32 {
	if x != nil {
		return x.Core.SrcVlan
	}
	return 0
}

func (x *EnrichedFlow) GetDstVlan() uint32 {
	if x != nil {
		return x.Core.DstVlan
	}
	return 0
}

func (x *EnrichedFlow) GetVlanId() uint32 {
	if x != nil {
		return x.Core.VlanId
	}
	return 0
}

func (x *EnrichedFlow) GetIngressVrfID() uint32 {
	if x != nil {
		return x.Core.IngressVrfID
	}
	return 0
}

func (x *EnrichedFlow) GetEgressVrfID() uint32 {
	if x != nil {
		return x.Core.EgressVrfID
	}
	return 0
}

func (x *EnrichedFlow) GetIPTos() uint32 {
	if x != nil {
		return x.Core.IPTos
	}
	return 0
}

func (x *EnrichedFlow) GetForwardingStatus() uint32 {
	if x != nil {
		return x.Core.ForwardingStatus
	}
	return 0
}

func (x *EnrichedFlow) GetIPTTL() uint32 {
	if x != nil {
		return x.Core.IPTTL
	}
	return 0
}

func (x *EnrichedFlow) GetTCPFlags() uint32 {
	if x != nil {
		return x.Core.TCPFlags
	}
	return 0
}

func (x *EnrichedFlow) GetIcmpType() uint32 {
	if x != nil {
		return x.Core.IcmpType
	}
	return 0
}

func (x *EnrichedFlow) GetIcmpCode() uint32 {
	if x != nil {
		return x.Core.IcmpCode
	}
	return 0
}

func (x *EnrichedFlow) GetIPv6FlowLabel() uint32 {
	if x != nil {
		return x.Core.IPv6FlowLabel
	}
	return 0
}

func (x *EnrichedFlow) GetFragmentId() uint32 {
	if x != nil {
		return x.Core.FragmentId
	}
	return 0
}

func (x *EnrichedFlow) GetFragmentOffset() uint32 {
	if x != nil {
		return x.Core.FragmentOffset
	}
	return 0
}

func (x *EnrichedFlow) GetBiFlowDirection() uint32 {
	if x != nil {
		return x.Core.BiFlowDirection
	}
	return 0
}

func (x *EnrichedFlow) GetSrcAS() uint32 {
	if x != nil {
		return x.Core.SrcAS
	}
	return 0
}

func (x *EnrichedFlow) GetDstAS() uint32 {
	if x != nil {
		return x.Core.DstAS
	}
	return 0
}

func (x *EnrichedFlow) GetNextHop() []byte {
	if x != nil {
		return x.Core.NextHop
	}
	return nil
}

func (x *EnrichedFlow) GetNextHopAS() uint32 {
	if x != nil {
		return x.Core.NextHopAS
	}
	return 0
}

func (x *EnrichedFlow) GetSrcNet() uint32 {
	if x != nil {
		return x.Core.SrcNet
	}
	return 0
}

func (x *EnrichedFlow) GetDstNet() uint32 {
	if x != nil {
		return x.Core.DstNet
	}
	return 0
}

func (x *EnrichedFlow) GetHasMPLS() bool {
	if x != nil {
		return x.Core.HasMPLS
	}
	return false
}

func (x *EnrichedFlow) GetMPLSCount() uint32 {
	if x != nil {
		return x.Core.MPLSCount
	}
	return 0
}

func (x *EnrichedFlow) GetMPLS1TTL() uint32 {
	if x != nil {
		return x.Core.MPLS1TTL
	}
	return 0
}

func (x *EnrichedFlow) GetMPLS1Label() uint32 {
	if x != nil {
		return x.Core.MPLS1Label
	}
	return 0
}

func (x *EnrichedFlow) GetMPLS2TTL() uint32 {
	if x != nil {
		return x.Core.MPLS2TTL
	}
	return 0
}

func (x *EnrichedFlow) GetMPLS2Label() uint32 {
	if x != nil {
		return x.Core.MPLS2Label
	}
	return 0
}

func (x *EnrichedFlow) GetMPLS3TTL() uint32 {
	if x != nil {
		return x.Core.MPLS3TTL
	}
	return 0
}

func (x *EnrichedFlow) GetMPLS3Label() uint32 {
	if x != nil {
		return x.Core.MPLS3Label
	}
	return 0
}

func (x *EnrichedFlow) GetMPLSLastTTL() uint32 {
	if x != nil {
		return x.Core.MPLSLastTTL
	}
	return 0
}

func (x *EnrichedFlow) GetMPLSLastLabel() uint32 {
	if x != nil {
		return x.Core.MPLSLastLabel
	}
	return 0
}

func (x *EnrichedFlow) GetCustomInteger1() uint64 {
	if x != nil {
		return x.Core.CustomInteger1
	}
	return 0
}

func (x *EnrichedFlow) GetCustomInteger2() uint64 {
	if x != nil {
		return x.Core.CustomInteger2
	}
	return 0
}

// TODO: these are part of goflow2 protobuf, but not part of its precompiled
// protobuf distribution.
// func (x *EnrichedFlow) GetCustomInteger3() uint64 {
// 	if x != nil {
// 		return x.Core.CustomInteger3
// 	}
// 	return 0
// }
//
// func (x *EnrichedFlow) GetCustomInteger4() uint64 {
// 	if x != nil {
// 		return x.Core.CustomInteger4
// 	}
// 	return 0
// }
//
// func (x *EnrichedFlow) GetCustomInteger5() uint64 {
// 	if x != nil {
// 		return x.Core.CustomInteger5
// 	}
// 	return 0
// }

func (x *EnrichedFlow) GetCustomBytes1() []byte {
	if x != nil {
		return x.Core.CustomBytes1
	}
	return nil
}

func (x *EnrichedFlow) GetCustomBytes2() []byte {
	if x != nil {
		return x.Core.CustomBytes2
	}
	return nil
}

// TODO: these are part of goflow2 protobuf, but not part of its precompiled
// protobuf distribution.
// func (x *EnrichedFlow) GetCustomBytes3() []byte {
// 	if x != nil {
// 		return x.Core.CustomBytes3
// 	}
// 	return nil
// }
//
// func (x *EnrichedFlow) GetCustomBytes4() []byte {
// 	if x != nil {
// 		return x.Core.CustomBytes4
// 	}
// 	return nil
// }
//
// func (x *EnrichedFlow) GetCustomBytes5() []byte {
// 	if x != nil {
// 		return x.Core.CustomBytes5
// 	}
// 	return nil
// }
