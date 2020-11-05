package core

import "time"

const FileSyncPath = "SYNC"

const OpArg = "pmp_exec"

const DataPath = "pmp_data"

const DefDuration = time.Second * 5 // 5s 上报一次slave 状态

const RpcTimeout = time.Second * 6
