package protocol

import "github.com/yddeng/pmp/net/pb"

const (
	CmdLogin  = 1
	CmdFile   = 2
	CmdStart  = 3
	CmdStop   = 4
	CmdKill   = 5
	CmdSignal = 6
)

func init() {
	pb.Register("pmp_msg", &File{}, CmdFile)

	pb.Register("pmp_req", &LoginReq{}, CmdLogin)
	pb.Register("pmp_resp", &LoginResp{}, CmdLogin)

	pb.Register("pmp_req", &StartReq{}, CmdStart)
	pb.Register("pmp_resp", &StartResp{}, CmdStart)

	pb.Register("pmp_req", &StopReq{}, CmdStop)
	pb.Register("pmp_resp", &StopResp{}, CmdStop)

	pb.Register("pmp_req", &KillReq{}, CmdKill)
	pb.Register("pmp_resp", &KillResp{}, CmdKill)

	pb.Register("pmp_req", &SignalReq{}, CmdSignal)
	pb.Register("pmp_resp", &SignalResp{}, CmdSignal)
}