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
	"flag"
	"log"
	"math"
	"math/big"
)

func main() {

	// as input, we provide how many parts of 10000 should be assigned
	// this corresponds to hundredths of percentages
	share := flag.Int("share", 10000, "pertenmille for distribution")
	flag.Parse()
	assignedPertenmille := big.NewInt(int64(*share))

	// the total supply is the maximum managable in an unsigned 64-bit integer
	tenThousand := big.NewInt(10000)
	totalSupply := big.NewInt(0).SetUint64(math.MaxUint64)

	// the number of tokens assigned is the total multiplied by the assigned
	// parts and then divided by ten thousand
	assignedTokens := big.NewInt(0)
	assignedTokens = assignedTokens.Mul(totalSupply, assignedPertenmille).Quo(assignedTokens, tenThousand)

	// to find the rounding error, we do the opposite, multiplying by ten thousand
	// and then dividing by the assigned parts
	reverseSupply := big.NewInt(0)
	reverseSupply = reverseSupply.Mul(assignedTokens, tenThousand).Quo(reverseSupply, assignedPertenmille)

	// the error margin then corresponds to the difference between total supply
	// and rounded back up supply
	errorMargin := big.NewInt(0)
	errorMargin = errorMargin.Sub(totalSupply, reverseSupply)

	log.Printf("Total supply: %v\n", totalSupply)
	log.Printf("Assigned pertenmille: %v\n", assignedPertenmille)
	log.Printf("Assigned tokens: %v\n", assignedTokens)
	log.Printf("Error margin: %v\n", errorMargin)
	log.Printf("(this should be positive)\n")
}
