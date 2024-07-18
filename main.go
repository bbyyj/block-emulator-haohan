package main

import (
	"blockEmulator/build"
	"blockEmulator/params"
	"fmt"
	"runtime"

	"github.com/spf13/pflag"
)

var (
	// network config
	shardNum int
	nodeNum  int
	shardID  int
	nodeID   int

	// consensus module
	modID int

	// the location for an experiment
	dataRootDir string

	// supervisor or not
	isSupervisor bool

	// batch running config
	isGen                bool
	isGenerateForExeFile bool

	// 下面是新增的
	algorithm           int
	totalDataSize       int
	injectSpeed         int
	maxBlockSize_global int
	block_Interval      int
)

func main() {
	// Start a node.
	pflag.IntVarP(&shardNum, "shardNum", "S", params.ShardNum, "shardNum is an Integer, which indicates that how many shards are deployed. ")
	pflag.IntVarP(&nodeNum, "nodeNum", "N", params.NodesInShard, "nodeNum is an Integer, which indicates how many nodes of each shard are deployed. ")
	pflag.IntVarP(&shardID, "shardID", "s", 0, "shardID is an Integer, which indicates the ID of the shard to which this node belongs. Value range: [0, shardNum). ")
	pflag.IntVarP(&nodeID, "nodeID", "n", 0, "nodeID is an Integer, which indicates the ID of this node. Value range: [0, nodeNum).")
	pflag.IntVarP(&modID, "modID", "m", 3, "modID is an Integer, which indicates the choice ID of methods / consensuses. Value range: [0, 4), representing [CLPA_Broker, CLPA, Broker, Relay]")
	pflag.StringVar(&dataRootDir, "dataRootDir", params.ExpDataRootDir, "dataRootDir is a string, which defines the RootDir of the experimental data, including ./log, ./record and ./result")
	pflag.BoolVarP(&isSupervisor, "supervisor", "c", false, "isSupervisor is a bool value, which indicates whether this node is a supervisor.")
	pflag.BoolVarP(&isGen, "gen", "g", false, "isGen is a bool value, which indicates whether to generate a batch file")
	pflag.BoolVarP(&isGenerateForExeFile, "shellForExe", "f", false, "isGenerateForExeFile is a bool value, which is effective only if 'isGen' is true; True to generate for an executable, False for 'go run'. ")

	// 下面是新增的
	pflag.IntVarP(&algorithm, "algorithm", "a", 0, "using of algorithm")
	pflag.IntVarP(&totalDataSize, "totalDataSize", "d", 900000, "the total number of txs")
	pflag.IntVarP(&injectSpeed, "injectSpeed", "i", 5000, "the transaction inject speed")
	pflag.IntVarP(&maxBlockSize_global, "maxBlockSize_global", "b", 2000, "the block contains the maximum number of transactions")
	pflag.IntVarP(&block_Interval, "block_Interval", "t", 6000, "generate new block interval")

	pflag.Parse()

	params.Block_Interval = block_Interval
	params.InjectSpeed = injectSpeed
	params.MaxBlockSize_global = maxBlockSize_global
	params.TotalDataSize = totalDataSize

	method := []string{"monoxide", "delayfirst", "proposed"}
	params.Algorithm = method[algorithm]

	fmt.Println("algorithm：", algorithm)
	fmt.Println("totalDataSize：", totalDataSize)
	fmt.Println("injectSpeed：", injectSpeed)
	fmt.Println("maxBlockSize_global：", maxBlockSize_global)
	fmt.Println("block_Interval：", block_Interval)

	if isGen {
		if isGenerateForExeFile {
			// Determine the current operating system.
			// Generate the corresponding .bat file or .sh file based on the detected operating system.
			os := runtime.GOOS
			switch os {
			case "windows":
				build.Exebat_Windows_GenerateBatFile(nodeNum, shardNum, modID, dataRootDir)
			case "darwin":
				build.Exebat_MacOS_GenerateShellFile(nodeNum, shardNum, modID, dataRootDir)
			case "linux":
				build.Exebat_Linux_GenerateShellFile(nodeNum, shardNum, modID, dataRootDir)
			}
		} else {
			// Without determining the operating system.
			// Generate a .bat file or .sh file for running `go run`.
			build.GenerateBatFile(nodeNum, shardNum, modID, dataRootDir)
			build.GenerateShellFile(nodeNum, shardNum, modID, dataRootDir)
		}

		return
	}

	// // set the experimental data root dir
	// params.DataWrite_path = dataRootDir + "/result/"
	// params.LogWrite_path = dataRootDir + "/log/"
	// params.DatabaseWrite_path = dataRootDir + "/database/"

	if isSupervisor {
		build.BuildSupervisor(uint64(nodeNum), uint64(shardNum), uint64(modID))
	} else {
		build.BuildNewPbftNode(uint64(nodeID), uint64(nodeNum), uint64(shardID), uint64(shardNum), uint64(modID))
	}
}
