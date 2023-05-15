package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/torquem-ch/mdbx-go/mdbx"
)

func main() {
	fmt.Printf("Hello this is the console, please write a call to interact with the database: \n")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		length := len(text)
		if bytes.Equal([]byte(text[:length-1]), []byte("eth.getBlockHeaderByNumber(0)")) {
			getBlockHeaderByNumber()
		} else if length > 26 && bytes.Equal([]byte(text[:27]), []byte("eth.getBlockHeaderByNumber(")) {
			fmt.Println("Block Doesn't Exists")
		} else if bytes.Equal([]byte(text[:length-1]), []byte("exit")) {
			os.Exit(0)
		} else {
			fmt.Println("Invalid API")
		}
	}

}

func getBlockHeaderByNumber() {
	env, err1 := mdbx.NewEnv()
	fmt.Println("Environment Created ", env, err1)

	if err1 != nil {
		fmt.Println("Cannot Open Environment")
	}

	err := env.Open("/media/kavania2002/VolumeE1/College Stuff/blockchain/project/devnet/gethdata/geth/chaindata", 0, 0664)
	defer env.Close()
	if err != nil {
		fmt.Println("Cannot Use Open function")
	}
	var hash common.Hash
	err = env.View(func(txn *mdbx.Txn) error {
		// fmt.Println("GET INITIATED")
		dbi, err := txn.OpenRoot(0)
		if err != nil {
			return err
		}
		data, _ := txn.Get(dbi, headerHashKey(0))
		readCan := common.BytesToHash(data)

		hash = readCan

		if readCan == (common.Hash{}) {
			fmt.Println("ReadCan")
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(hash)
	header := new(types.Header)
	err = env.View(func(txn *mdbx.Txn) error {
		// fmt.Println("GET INITIATED")
		dbi, err := txn.OpenRoot(0)
		if err != nil {
			return err
		}
		data, _ := txn.Get(dbi, headerKey(0, hash))

		if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
			fmt.Println("Invalid block header RLP")
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(header)
	// body := new(types.Body)
	// err = env.View(func(txn *mdbx.Txn) error {
	// 	// fmt.Println("GET INITIATED")
	// 	dbi, err := txn.OpenRoot(0)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	fmt.Println(blockBodyKey(0, hash))
	// 	data2, _ := txn.Get(dbi, blockBodyKey(0, hash))
	// 	fmt.Println(data2)
	// 	if err := rlp.Decode(bytes.NewReader(data2), body); err != nil {
	// 		fmt.Println("Invalid block body RLP", hash)
	// 		return err
	// 	}

	// 	return nil
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// //fmt.Println(types.NewBlockWithHeader(header).WithBody(body.Transactions, body.Uncles).WithWithdrawals(body.Withdrawals))
	fmt.Println(types.NewBlockWithHeader(header))
}

func headerHashKey(number uint64) []byte {
	return append(append([]byte("h"), encodeBlockNumber(number)...), []byte("n")...)
}
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// func blockBodyKey(number uint64, hash common.Hash) []byte {
// return append(append([]byte("r"), encodeBlockNumber(number)...), hash.Bytes()...)
// }
func headerKey(number uint64, hash common.Hash) []byte {
	return append(append([]byte("h"), encodeBlockNumber(number)...), hash.Bytes()...)
}

// func GetBlock(hash common.Hash, number uint64) *types.Block{
// 	block  :=
// 	if block == nil{
// 		return nil
// 	}
// }
