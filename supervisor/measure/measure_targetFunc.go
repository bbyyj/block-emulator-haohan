package measure

import (
	"blockEmulator/message"
	"blockEmulator/params"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"time"
)

// to test average Transaction_Confirm_Latency (TCL)  in this system
type TestUtilityFunction struct {
	epochID                int
	itxNumber              []float64
	relay1Number           []float64
	relay2Number           []float64
	txHashes               []map[string]bool
	relayHash_2_Relay1Time map[string]time.Time // hash of relay transactions -> relay1 time
	relayHash_2_Relay2Time map[string]time.Time // hash of relay transactions -> relay1 time
}

func NewTestUtilityFunction() *TestUtilityFunction {
	return &TestUtilityFunction{
		epochID:                -1,
		itxNumber:              make([]float64, 0),
		relay1Number:           make([]float64, 0),
		relay2Number:           make([]float64, 0),
		txHashes:               make([]map[string]bool, 0),
		relayHash_2_Relay1Time: make(map[string]time.Time),
		relayHash_2_Relay2Time: make(map[string]time.Time),
	}
}

func (tml *TestUtilityFunction) OutputMetricName() string {
	return "UtilityFunction Result"
}

// modified latency
func (tml *TestUtilityFunction) UpdateMeasureRecord(b *message.BlockInfoMsg) {
	if b.BlockBodyLength == 0 { // empty block
		return
	}

	epochid := b.Epoch
	txs := append(b.InnerShardTxs, b.Relay2Txs...)

	// extend
	for tml.epochID < epochid {
		tml.itxNumber = append(tml.itxNumber, 0)
		tml.relay1Number = append(tml.relay1Number, 0)
		tml.relay2Number = append(tml.relay2Number, 0)
		tml.txHashes = append(tml.txHashes, make(map[string]bool))
		tml.epochID++
	}

	// add relay1 number
	tml.relay1Number[epochid] += float64(len(b.Relay1Txs))
	// add relayHash into the pool
	for _, r1tx := range b.Relay1Txs {
		tml.relayHash_2_Relay1Time[string(r1tx.TxHash)] = b.CommitTime
	}

	// add relay2 number & itx number
	for _, tx := range txs {
		if _, ok := tml.relayHash_2_Relay1Time[string(tx.TxHash)]; ok {
			tml.relayHash_2_Relay2Time[string(tx.TxHash)] = b.CommitTime
			tml.relay2Number[epochid] += 1
			tml.txHashes[epochid][string(tx.TxHash)] = true
		} else {
			tml.itxNumber[epochid] += 1
		}
	}
}

func (tml *TestUtilityFunction) HandleExtraMessage([]byte) {}

func ComputeUtilityFunction(itxNum, r1Num, r2Num float64, blockHeightGap int) float64 {
	a, b := 0.5, 2.0
	return itxNum + a*r1Num + b*r2Num - float64(blockHeightGap)
}

func ComputeUtilityFunctionFirst(itxNum, r1Num, r2Num float64) float64 {
	a, b := 0.5, 2.0
	return itxNum + a*r1Num + b*r2Num
}

func ComputeUtilityFunctionSecond(blockHeightGap int) float64 {
	return float64(blockHeightGap)
}

func (tml *TestUtilityFunction) OutputRecord() (perEpochUtility []float64, totUtility float64) {
	perEpochUtility = make([]float64, 0)
	perEpochUtilityFirst := make([]float64, 0)
	perEpochUtilitSecond := make([]float64, 0)
	bhgList := make([]int, 0)
	blockGapMap := make(map[string]float64)
	totalitx, totalr1tx, totalr2tx := 0.0, 0.0, 0.0
	totalBlockGaps := 0
	for epoch := 0; epoch <= tml.epochID; epoch++ {
		// compute blockHeightGap
		bhg := 0
		for txhash := range tml.txHashes[epoch] {
			tx_gapTime := tml.relayHash_2_Relay2Time[txhash].Sub(tml.relayHash_2_Relay1Time[txhash])
			blockGap := tx_gapTime.Seconds()/float64(params.Block_Interval/1000) + 1
			blockGapMap[txhash] = blockGap
			bhg += int(blockGap)
		}
		bhgList = append(bhgList, bhg)
		totalBlockGaps += bhg
		totalitx += tml.itxNumber[epoch]
		totalr1tx += tml.relay1Number[epoch]
		totalr2tx += tml.relay2Number[epoch]

		perEpochUtility = append(perEpochUtility, ComputeUtilityFunction(tml.itxNumber[epoch], tml.relay1Number[epoch], tml.relay2Number[epoch], bhgList[epoch]))
		perEpochUtilityFirst = append(perEpochUtilityFirst, ComputeUtilityFunctionFirst(tml.itxNumber[epoch], tml.relay1Number[epoch], tml.relay2Number[epoch]))
		perEpochUtilitSecond = append(perEpochUtilitSecond, ComputeUtilityFunctionSecond(bhgList[epoch]))
	}
	writeToCSViInUtilty(blockGapMap)
	writeToCSViInUtiltyFirst(perEpochUtilityFirst)
	writeToCSViInUtiltySecond(perEpochUtilitSecond)
	totUtility = ComputeUtilityFunction(totalitx, totalr1tx, totalr2tx, totalBlockGaps)
	return
}

func writeToCSViInUtilty(blockGapMap map[string]float64) {
	dirpath := params.DataWrite_path + "supervisor_measureOutput/"
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	targetPath := dirpath + "blockGap.csv"
	file, er := os.Create(targetPath)
	if er != nil {
		panic(er)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	title := []string{"txhash", "blockGap"}
	w.Write(title)
	w.Flush()
	datanum := 0
	for _, val := range blockGapMap {
		resultStr := []string{strconv.Itoa(int(datanum)), strconv.Itoa(int(val))}
		datanum++
		err = w.Write(resultStr)
		if err != nil {
			log.Panic()
		}
		w.Flush()
	}
}

func writeToCSViInUtiltyFirst(perEpochUtilityFirst []float64) {
	dirpath := params.DataWrite_path + "supervisor_measureOutput/"
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	targetPath := dirpath + "perEpochUtilityFirst.csv"
	file, er := os.Create(targetPath)
	if er != nil {
		panic(er)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	title := []string{"blockheight", "blockGap"}
	w.Write(title)
	w.Flush()
	datanum := 0
	for _, val := range perEpochUtilityFirst {
		resultStr := []string{strconv.Itoa(int(datanum)), strconv.Itoa(int(val))}
		datanum++
		err = w.Write(resultStr)
		if err != nil {
			log.Panic()
		}
		w.Flush()
	}
}

func writeToCSViInUtiltySecond(perEpochUtilitSecond []float64) {
	dirpath := params.DataWrite_path + "supervisor_measureOutput/"
	err := os.MkdirAll(dirpath, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	targetPath := dirpath + "perEpochUtilitySecond.csv"
	file, er := os.Create(targetPath)
	if er != nil {
		panic(er)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	title := []string{"blockheight", "blockGap"}
	w.Write(title)
	w.Flush()
	datanum := 0
	for _, val := range perEpochUtilitSecond {
		resultStr := []string{strconv.Itoa(int(datanum)), strconv.Itoa(int(val))}
		datanum++
		err = w.Write(resultStr)
		if err != nil {
			log.Panic()
		}
		w.Flush()
	}
}
