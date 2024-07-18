package build

import (
	"blockEmulator/consensus_shard/pbft_all"
	"blockEmulator/params"
	"blockEmulator/supervisor"
	"fmt"
	"strconv"
	"time"
)

func initConfig(nid, nnm, sid, snm uint64) *params.ChainConfig {
	params.ShardNum = int(snm)
	nowIpPort := 3
	shardsInMechine := params.ShardNum / 8
	for gap := 0; gap < int(snm); gap += shardsInMechine {
		for i := gap; i < gap+shardsInMechine; i++ {

			if _, ok := params.IPmap_nodeTable[uint64(i)]; !ok {
				params.IPmap_nodeTable[uint64(i)] = make(map[uint64]string)
			}
			if nowIpPort == 5 {
				for j := uint64(0); j < nnm; j++ {
					params.IPmap_nodeTable[uint64(i)][j] = "192.168.3.11:" + strconv.Itoa(31800+(i-gap)*100+int(j))
				}
			} else {
				for j := uint64(0); j < nnm; j++ {
					params.IPmap_nodeTable[uint64(i)][j] = "192.168.3." + strconv.Itoa(nowIpPort) + ":" + strconv.Itoa(31800+(i-gap)*100+int(j))
				}
			}

		}
		nowIpPort++
	}

	params.IPmap_nodeTable[params.DeciderShard] = make(map[uint64]string)
	params.IPmap_nodeTable[params.DeciderShard][0] = params.SupervisorAddr
	params.NodesInShard = int(nnm)
	params.ShardNum = int(snm)
	pcc := &params.ChainConfig{
		ChainID:        sid,
		NodeID:         nid,
		ShardID:        sid,
		Nodes_perShard: uint64(params.NodesInShard),
		ShardNums:      snm,
		BlockSize:      uint64(params.MaxBlockSize_global),
		BlockInterval:  uint64(params.Block_Interval),
		InjectSpeed:    uint64(params.InjectSpeed),
	}
	return pcc
}

func BuildSupervisor(nnm, snm, mod uint64) {
	var measureMod []string
	if mod == 0 || mod == 2 {
		measureMod = params.MeasureBrokerMod
	} else {
		measureMod = params.MeasureRelayMod
	}
	measureMod = append(measureMod, "Tx_Details")

	lsn := new(supervisor.Supervisor)
	fmt.Println("123")
	lsn.NewSupervisor(params.SupervisorAddr, initConfig(123, nnm, 123, snm), params.CommitteeMethod[mod], measureMod...)
	time.Sleep(10000 * time.Millisecond)
	go lsn.SupervisorTxHandling()
	lsn.TcpListen()
}

func BuildNewPbftNode(nid, nnm, sid, snm, mod uint64) {
	worker := pbft_all.NewPbftNode(sid, nid, initConfig(nid, nnm, sid, snm), params.CommitteeMethod[mod])
	go worker.TcpListen()
	worker.Propose()
}
