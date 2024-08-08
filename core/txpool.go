// the define and some operation of txpool

package core

import (
	"blockEmulator/params"
	"blockEmulator/utils"
	"fmt"
	"sync"
	"time"
)

type TxPool struct {
	TxQueue   []*Transaction            // transaction Queue
	RelayPool map[uint64][]*Transaction //designed for sharded blockchain, from Monoxide
	lock      sync.Mutex
	// The pending list is ignored
}

func NewTxPool() *TxPool {
	return &TxPool{
		TxQueue:   make([]*Transaction, 0),
		RelayPool: make(map[uint64][]*Transaction),
	}
}

// Add a transaction to the pool (consider the queue only)
func (txpool *TxPool) AddTx2Pool(tx *Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	if tx.Time.IsZero() {
		tx.Time = time.Now()
	}
	txpool.TxQueue = append(txpool.TxQueue, tx)
}

// Add a list of transactions to the pool
func (txpool *TxPool) AddTxs2Pool(txs []*Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	for _, tx := range txs {
		if tx.Time.IsZero() {
			tx.Time = time.Now()
		}
		txpool.TxQueue = append(txpool.TxQueue, tx)
	}
}

// 对交易等待时间排序
func bubbleSort(txpool *TxPool) {
	length := len(txpool.TxQueue)
	for i := 0; i < length; i++ {
		for j := 0; j < length-1-i; j++ {
			if txpool.TxQueue[j+1].Time.Before(txpool.TxQueue[j].Time) {
				txpool.TxQueue[j], txpool.TxQueue[j+1] = txpool.TxQueue[j+1], txpool.TxQueue[j]
			}
		}
	}
}

// add transactions into the pool head
/* func (txpool *TxPool) AddTxs2Pool_Head(tx []*Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txpool.TxQueue = append(tx, txpool.TxQueue...)
} */
//v2
func (txpool *TxPool) AddTxs2Pool_Head(Txs []*Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	for _, tx := range Txs {
		if tx.Time.IsZero() {
			tx.Time = time.Now()
		}
	}
	txpool.TxQueue = append(Txs, txpool.TxQueue...)
}

//v1
/* func (txpool *TxPool) AddTxs2Pool_Head(Txs []*Transaction) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	for _, tx := range Txs {
		if tx.Time.IsZero() {
			tx.Time = time.Now()
		}
		txpool.TxQueue = append(txpool.TxQueue, tx)
		idex := 0
		copy(txpool.TxQueue[idex+1:], txpool.TxQueue[idex:])
		txpool.TxQueue[idex] = tx
	}
} */

// Pack transactions for a proposal

func (txpool *TxPool) PackTxs(max_txs uint64) []*Transaction {
	fmt.Println("params.Algorithm：", params.Algorithm)
	if params.Algorithm == "monoxide" {
		txpool.lock.Lock()
		defer txpool.lock.Unlock()
		//bubbleSort(txpool) // 按照时间先后顺序优先打包等待时间久的交易
		txNum := max_txs
		if uint64(len(txpool.TxQueue)) < txNum {
			txNum = uint64(len(txpool.TxQueue))
		}
		txs_Packed := txpool.TxQueue[:txNum]
		txpool.TxQueue = txpool.TxQueue[txNum:]
		return txs_Packed
	} else if params.Algorithm == "delayfirst" {
		txpool.lock.Lock()
		defer txpool.lock.Unlock()
		bubbleSort(txpool) // 按照时间先后顺序优先打包等待时间久的交易
		txNum := max_txs
		if uint64(len(txpool.TxQueue)) < txNum {
			txNum = uint64(len(txpool.TxQueue))
		}
		txs_Packed := txpool.TxQueue[:txNum]
		txpool.TxQueue = txpool.TxQueue[txNum:]
		return txs_Packed
	} else {
		txpool.lock.Lock()
		defer txpool.lock.Unlock()
		size_TxQueue := uint64(len(txpool.TxQueue))
		txNum := max_txs
		if size_TxQueue < txNum {
			//txs_Packed := make([]*Transaction, size_TxQueue)
			txs_Packed := []*Transaction{}
			txNum = size_TxQueue
			txs_Packed = txpool.TxQueue[:txNum]
			txpool.TxQueue = txpool.TxQueue[txNum:]
			return txs_Packed
		} else {
			var left uint64 = 0
			var right, middle uint64 = size_TxQueue - 2, size_TxQueue - 2 // 定义了target在左闭右闭的区间内，[left, right]
			var target uint64 = 0
			//txs_Packed := make([]*Transaction, txNum)
			txs_Packed := []*Transaction{}
			for {
				if middle == 0 {
					break
				}
				if left > right {
					break //当left == right时，区间[left, right]仍然有效
				}
				middle = left + ((right - left) / 2) //等同于 (left + right) / 2，防止溢出
				if txpool.TxQueue[middle].Relayed == true && txpool.TxQueue[middle+1].Relayed == true {
					left = middle + 1 //target在右区间，所以[middle + 1, right]
				} else if txpool.TxQueue[middle].Relayed == false && txpool.TxQueue[middle+1].Relayed == false {
					right = middle - 1 //target在左区间，所以[left, middle - 1]
				} else if txpool.TxQueue[middle].Relayed == true && txpool.TxQueue[middle+1].Relayed == false {
					target = middle
					break
				}
			}
			//优先处理后半和前半
			if target <= txNum {
				txs_Packing := []*Transaction{}
				txs_Packing = txpool.TxQueue[:target]
				txpool.TxQueue = txpool.TxQueue[target:]
				i := 0
				for {
					if txpool.TxQueue[i].Sender != txpool.TxQueue[i].Recipient {
						txs_Packing = append(txs_Packing, txpool.TxQueue[i])
						txpool.TxQueue = append(txpool.TxQueue[:i], txpool.TxQueue[i+1:]...)
						i--
					}
					i++
					if uint64(len(txs_Packing)) == txNum {
						break
					}
					if i == len(txpool.TxQueue) {
						txs_Packing = append(txs_Packing, txpool.TxQueue[:txNum-uint64(len(txs_Packing))]...)
						txpool.TxQueue = txpool.TxQueue[txNum-uint64(len(txs_Packing)):]
					}
				}
				txs_Packed = txs_Packing
			} else {
				copy(txs_Packed, txpool.TxQueue[target-txNum:target])
				txpool.TxQueue = append(txpool.TxQueue[:target-txNum], txpool.TxQueue[target:]...)
			}
			return txs_Packed

		}
	}

}

//func (txpool *TxPool) PackTxs(max_txs uint64) []*Transaction {
//	txpool.lock.Lock()
//	defer txpool.lock.Unlock()
//	size_TxQueue := uint64(len(txpool.TxQueue))
//	txNum := max_txs
//	if size_TxQueue < txNum {
//		//txs_Packed := make([]*Transaction, size_TxQueue)
//		txs_Packed := []*Transaction{}
//		txNum = size_TxQueue
//		txs_Packed = txpool.TxQueue[:txNum]
//		txpool.TxQueue = txpool.TxQueue[txNum:]
//		return txs_Packed
//	} else {
//		var left uint64 = 0
//		var right, middle uint64 = size_TxQueue - 2, size_TxQueue - 2 // 定义了target在左闭右闭的区间内，[left, right]
//		var target uint64 = 0
//		//txs_Packed := make([]*Transaction, txNum)
//		txs_Packed := []*Transaction{}
//		for {
//			if middle == 0 {
//				break
//			}
//			if left > right {
//				break //当left == right时，区间[left, right]仍然有效
//			}
//			middle = left + ((right - left) / 2) //等同于 (left + right) / 2，防止溢出
//			if txpool.TxQueue[middle].Relayed == true && txpool.TxQueue[middle+1].Relayed == true {
//				left = middle + 1 //target在右区间，所以[middle + 1, right]
//			} else if txpool.TxQueue[middle].Relayed == false && txpool.TxQueue[middle+1].Relayed == false {
//				right = middle - 1 //target在左区间，所以[left, middle - 1]
//			} else if txpool.TxQueue[middle].Relayed == true && txpool.TxQueue[middle+1].Relayed == false {
//				target = middle
//				break
//			}
//		}
//
//		//fmt.Println("target", target)
//		//v3
//		if target < txNum {
//			txs_Packed = txpool.TxQueue[:txNum]
//			txpool.TxQueue = txpool.TxQueue[txNum:]
//		} else {
//			copy(txs_Packed, txpool.TxQueue[target-txNum:target])
//			txpool.TxQueue = append(txpool.TxQueue[:target-txNum], txpool.TxQueue[target:]...)
//		}
//
//		//v4
//
//		if (size_TxQueue - target) < txNum/13 {
//			//copy(txs_Packed, txpool.TxQueue[size_TxQueue-txNum:size_TxQueue])
//			txs_Packed = append(txs_Packed, txpool.TxQueue[size_TxQueue-txNum:size_TxQueue]...)
//			txpool.TxQueue = txpool.TxQueue[:size_TxQueue-txNum]
//		} else {
//			//copy(txs_Packed, txpool.TxQueue[size_TxQueue-txNum/10:size_TxQueue])
//			txs_Packed = append(txs_Packed, txpool.TxQueue[size_TxQueue-txNum/13:size_TxQueue]...)
//			txpool.TxQueue = txpool.TxQueue[:size_TxQueue-txNum/13]
//			if target < txNum-txNum/13 {
//				//copy(txs_Packed, txpool.TxQueue[:txNum-txNum/10])
//				txs_Packed = append(txs_Packed, txpool.TxQueue[:txNum-txNum/10]...)
//				txpool.TxQueue = txpool.TxQueue[txNum-txNum/13:]
//			} else {
//				//copy(txs_Packed, txpool.TxQueue[target+txNum/10-txNum:target])
//				txs_Packed = append(txs_Packed, txpool.TxQueue[target+txNum/13-txNum+1:target+1]...)
//				txpool.TxQueue = append(txpool.TxQueue[:target+txNum/13+1-txNum], txpool.TxQueue[target+1:]...)
//			}
//
//		}
//
//		//优先处理后半和前半
//		if target <= txNum {
//			txs_Packing := []*Transaction{}
//			txs_Packing = txpool.TxQueue[:target]
//			txpool.TxQueue = txpool.TxQueue[target:]
//			i := 0
//			for {
//				if txpool.TxQueue[i].Sender != txpool.TxQueue[i].Recipient {
//					txs_Packing = append(txs_Packing, txpool.TxQueue[i])
//					txpool.TxQueue = append(txpool.TxQueue[:i], txpool.TxQueue[i+1:]...)
//					i--
//				}
//				i++
//				if uint64(len(txs_Packing)) == txNum {
//					break
//				}
//				if i == len(txpool.TxQueue) {
//					txs_Packing = append(txs_Packing, txpool.TxQueue[:txNum-uint64(len(txs_Packing))]...)
//					txpool.TxQueue = txpool.TxQueue[txNum-uint64(len(txs_Packing)):]
//				}
//			}
//			txs_Packed = txs_Packing
//		} else {
//			copy(txs_Packed, txpool.TxQueue[target-txNum:target])
//			txpool.TxQueue = append(txpool.TxQueue[:target-txNum], txpool.TxQueue[target:]...)
//		}
//
//		//v5
//		if (size_TxQueue - target) <= 5*txNum/8 {
//			//copy(txs_Packed, TxQueue[size_TxQueue-txNum:size_TxQueue])
//			txs_Packed = append(txs_Packed, txpool.TxQueue[size_TxQueue-txNum:size_TxQueue]...)
//			txpool.TxQueue = txpool.TxQueue[:size_TxQueue-txNum]
//		} else {
//			if (target + 5*txNum/8) < txNum {
//				txs_Packed = txpool.TxQueue[:txNum]
//				txpool.TxQueue = txpool.TxQueue[txNum:]
//			} else {
//				//copy(txs_Packed, TxQueue[target+txNum/10-txNum:target+txNum/10])
//				txs_Packed = append(txs_Packed, txpool.TxQueue[target+5*txNum/8-txNum+1:target+5*txNum/8+1]...)
//				txpool.TxQueue = append(txpool.TxQueue[:target+5*txNum/8-txNum+1], txpool.TxQueue[target+5*txNum/8+1:]...)
//			}
//		}
//
//		return txs_Packed
//	}
//}

// Relay transactions
func (txpool *TxPool) AddRelayTx(tx *Transaction, shardID uint64) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	_, ok := txpool.RelayPool[shardID]
	if !ok {
		txpool.RelayPool[shardID] = make([]*Transaction, 0)
	}
	txpool.RelayPool[shardID] = append(txpool.RelayPool[shardID], tx)
}

// txpool get locked
func (txpool *TxPool) GetLocked() {
	txpool.lock.Lock()
}

// txpool get unlocked
func (txpool *TxPool) GetUnlocked() {
	txpool.lock.Unlock()
}

// get the length of tx queue
func (txpool *TxPool) GetTxQueueLen() int {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	return len(txpool.TxQueue)
}

// get the length of ClearRelayPool
func (txpool *TxPool) ClearRelayPool() {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txpool.RelayPool = nil
}

// abort ! Pack relay transactions from relay pool
func (txpool *TxPool) PackRelayTxs(shardID, minRelaySize, maxRelaySize uint64) ([]*Transaction, bool) {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	if _, ok := txpool.RelayPool[shardID]; !ok {
		return nil, false
	}
	if len(txpool.RelayPool[shardID]) < int(minRelaySize) {
		return nil, false
	}
	txNum := maxRelaySize
	if uint64(len(txpool.RelayPool[shardID])) < txNum {
		txNum = uint64(len(txpool.RelayPool[shardID]))
	}
	relayTxPacked := txpool.RelayPool[shardID][:txNum]
	txpool.RelayPool[shardID] = txpool.RelayPool[shardID][txNum:]
	return relayTxPacked, true
}

// abort ! Transfer transactions when re-sharding
func (txpool *TxPool) TransferTxs(addr utils.Address) []*Transaction {
	txpool.lock.Lock()
	defer txpool.lock.Unlock()
	txTransfered := make([]*Transaction, 0)
	newTxQueue := make([]*Transaction, 0)
	for _, tx := range txpool.TxQueue {
		if tx.Sender == addr {
			txTransfered = append(txTransfered, tx)
		} else {
			newTxQueue = append(newTxQueue, tx)
		}
	}
	newRelayPool := make(map[uint64][]*Transaction)
	for shardID, shardPool := range txpool.RelayPool {
		for _, tx := range shardPool {
			if tx.Sender == addr {
				txTransfered = append(txTransfered, tx)
			} else {
				if _, ok := newRelayPool[shardID]; !ok {
					newRelayPool[shardID] = make([]*Transaction, 0)
				}
				newRelayPool[shardID] = append(newRelayPool[shardID], tx)
			}
		}
	}
	txpool.TxQueue = newTxQueue
	txpool.RelayPool = newRelayPool
	return txTransfered
}
