package core

import "time"

const SharedPath = "shared"

const OpArg = "pmp_exec"

const DataPath = "pmp_data"

const DefDuration = time.Second * 5 // 5s 上报一次slave 状态

const RpcTimeout = time.Second * 6

const TimeFormat = "2006-01-02 15:04:05"
