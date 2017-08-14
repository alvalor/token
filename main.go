// Copyright (c) 2017 The Alvalor Authors
//
// This file is part of Alvalor.
//
// Alvalor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Alvalor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

type stats struct {
	total   uint64
	empty   uint64
	balance *big.Int
}

func (s stats) String() string {
	bal := big.NewFloat(0).SetInt(s.balance)
	con, _ := big.NewFloat(0).SetString("1000000000000000000")
	eth := big.NewFloat(0).Quo(bal, con)
	return fmt.Sprintf("%v total (%v empty) - %v ETH", s.total, s.empty, eth)
}

func main() {

	log.Println("startup")

	// make sure to catch signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	log.Println("opening database")

	// initialize DB connection
	path := "/Users/awishformore/Library/Ethereum/geth/chaindata"
	db, err := ethdb.NewLDBDatabase(path, 16, 16)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("initializing state trie")

	// initialize the state trie
	hash := core.GetHeadBlockHash(db)
	number := core.GetBlockNumber(db, hash)
	block := core.GetBlock(db, hash, number)
	t, err := trie.New(block.Root(), db)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	log.Println("iterating over state trie...")

	// initialize iterator and walk through DB
	contracts := &stats{balance: big.NewInt(0)}
	accounts := &stats{balance: big.NewInt(0)}
	weird := &stats{balance: big.NewInt(0)}
	it := trie.NewIterator(t.NodeIterator(nil))
	var s *stats
	var last uint64
Loop:
	for it.Next() {

		// if we received a signal, abort the loop
		select {
		case <-c:
			break Loop
		default:
		}

		// decode the data
		// addr := common.BytesToAddress(it.Key)
		var account state.Account
		err = rlp.DecodeBytes(it.Value, &account)
		if err != nil {
			log.Println(err)
			continue
		}

		// check what type of account we have
		zeroHash, _ := hex.DecodeString("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
		weirdHash, _ := hex.DecodeString("1e4366082444073b4bea23568c108c4288f2293bf126f5cd6359c41ec52fc7fe")
		switch {
		case bytes.Equal(account.CodeHash, zeroHash):
			s = accounts
		case bytes.Equal(account.CodeHash, weirdHash):
			s = weird
		default:
			s = contracts
		}

		// add the statistics
		s.balance.Add(s.balance, account.Balance)
		s.total++
		if s.balance.Uint64() == 0 {
			s.empty++
		}

		// progress
		if accounts.total%100000 == 0 && accounts.total != last {
			last = accounts.total
			log.Printf("processed: %v accounts, %v weird, %v contracts", accounts.total, weird.total, contracts.total)
		}
	}

	log.Printf("accounts: %v", accounts)
	log.Printf("weird: %v", weird)
	log.Printf("contracts: %v", contracts)

	log.Println("closing database")

	db.Close()

	log.Println("shutdown")

	os.Exit(0)
}
