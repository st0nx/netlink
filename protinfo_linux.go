package netlink

import (
	"fmt"
	"syscall"

	"github.com/monsterzz/netlink/nl"
)

func LinkGetProtinfo(link Link) (Protinfo, error) {
	return pkgHandle.LinkGetProtinfo(link)
}

func (h *Handle) LinkGetProtinfo(link Link) (Protinfo, error) {
	base := link.Attrs()
	h.ensureIndex(base)
	var pi Protinfo
	req := h.newNetlinkRequest(syscall.RTM_GETLINK, syscall.NLM_F_DUMP)
	msg := nl.NewIfInfomsg(syscall.AF_BRIDGE)
	req.AddData(msg)
	msgs, err := req.Execute(syscall.NETLINK_ROUTE, 0)
	if err != nil {
		return pi, err
	}

	for _, m := range msgs {
		ans := nl.DeserializeIfInfomsg(m)
		if int(ans.Index) != base.Index {
			continue
		}
		attrs, err := nl.ParseRouteAttr(m[ans.Len():])
		if err != nil {
			return pi, err
		}
		for _, attr := range attrs {
			if attr.Attr.Type != syscall.IFLA_PROTINFO|syscall.NLA_F_NESTED {
				continue
			}
			infos, err := nl.ParseRouteAttr(attr.Value)
			if err != nil {
				return pi, err
			}
			pi = *parseProtinfo(infos)

			return pi, nil
		}
	}
	return pi, fmt.Errorf("Device with index %d not found", base.Index)
}

func parseProtinfo(infos []syscall.NetlinkRouteAttr) *Protinfo {
	var pi Protinfo
	for _, info := range infos {
		switch info.Attr.Type {
		case nl.IFLA_BRPORT_MODE:
			pi.Hairpin = byteToBool(info.Value[0])
		case nl.IFLA_BRPORT_GUARD:
			pi.Guard = byteToBool(info.Value[0])
		case nl.IFLA_BRPORT_FAST_LEAVE:
			pi.FastLeave = byteToBool(info.Value[0])
		case nl.IFLA_BRPORT_PROTECT:
			pi.RootBlock = byteToBool(info.Value[0])
		case nl.IFLA_BRPORT_LEARNING:
			pi.Learning = byteToBool(info.Value[0])
		case nl.IFLA_BRPORT_UNICAST_FLOOD:
			pi.Flood = byteToBool(info.Value[0])
		}
	}
	return &pi
}
