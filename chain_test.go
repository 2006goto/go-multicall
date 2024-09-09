package multicall

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chainERC20TestCase struct {
	chain         string
	rpcURL        string
	contract      string
	address       string
	multicallAddr string
}

const deadAddr = "0x000000000000000000000000000000000000dEaD"

var ttCall = []chainERC20TestCase{
	{
		chain:    "MANTLE",
		rpcURL:   "https://5000.rpc.thirdweb.com",
		address:  deadAddr,
		contract: "0x371c7ec6D8039ff7933a2AA28EB827Ffe1F52f07",
	},
	{
		chain:  "ONTOLOGY",
		rpcURL: "https://dappnode3.ont.io:10339",
		// rpcURL:        "https://dappnode4.ont.io:10339",
		address:       "0x3c1edbff210cb10311359ca2fce711d6b8c74073",
		contract:      "0xae834526aa3b70de9b34f81c4bf51bc2c80a5473",
		multicallAddr: "0x381e7fb9671aad4b48d02dbee167b1e4cbf597ae",
	},
}

func TestCall_ChainERC20(t *testing.T) {
	ctx := context.Background()

	abiERC20, err := ParseABI(ERC20ABI)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range ttCall {
		t.Run(tt.chain, func(t *testing.T) {
			var multicallAddresses []string
			if tt.multicallAddr != "" {
				multicallAddresses = []string{tt.multicallAddr}
			}
			caller, err := Dial(ctx, tt.rpcURL, multicallAddresses...)
			if err != nil {
				t.Fatal(err)
			}
			var calls []*Call

			notContract := &Contract{
				ABI:     abiERC20,
				Address: common.HexToAddress(tt.contract),
			}
			notContractBalance := new(BalanceOutput)
			decimals := new(DecimalsOutput)
			calls = append(calls, notContract.NewCall(
				notContractBalance,
				"balanceOf",
				common.HexToAddress(tt.address),
			).AllowFailure())
			calls = append(calls, notContract.NewCall(
				decimals,
				"decimals",
			).AllowFailure())

			res, err := caller.Call(nil, calls...)
			if err != nil {
				t.Error(err)
				return
			}
			if len(res) == 0 {
				t.Error("no results")
				return
			}
			require.False(t, res[0].Failed)
			assert.True(t, res[0].Outputs.(*BalanceOutput).Balance.Cmp(big.NewInt(0)) > 0)
			t.Log(notContractBalance.Balance)
			t.Log(res[0].Failed)
			t.Log(decimals.Decimals)
		})
	}
}

const MulticallABI = `[
  {
    "inputs": [
      {
        "internalType": "address",
        "name": "addr",
        "type": "address"
      }
    ],
    "name": "getEthBalance",
    "outputs": [
      {
        "internalType": "uint256",
        "name": "balance",
        "type": "uint256"
      }
    ],
    "stateMutability": "view",
    "type": "function"
  }
]`

func TestCall_ChainEthBalance(t *testing.T) {
	ctx := context.Background()

	abiERC20, err := ParseABI(MulticallABI)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range ttCall {
		t.Run(tt.chain, func(t *testing.T) {
			var multicallAddresses []string
			if tt.multicallAddr != "" {
				multicallAddresses = []string{tt.multicallAddr}
			}
			caller, err := Dial(ctx, tt.rpcURL, multicallAddresses...)
			if err != nil {
				t.Fatal(err)
			}
			var calls []*Call

			var maddr string
			if tt.multicallAddr != "" {
				maddr = tt.multicallAddr
			} else {
				maddr = DefaultAddress
			}
			mcontract := &Contract{
				ABI:     abiERC20,
				Address: common.HexToAddress(maddr),
			}
			mbalance := new(BalanceOutput)
			calls = append(calls, mcontract.NewCall(
				mbalance,
				"getEthBalance",
				common.HexToAddress(tt.address),
			).AllowFailure())

			res, err := caller.Call(nil, calls...)
			if err != nil {
				t.Error(err)
				return
			}
			if len(res) == 0 {
				t.Error("no results")
				return
			}
			assert.False(t, res[0].Failed)
			assert.True(t, res[0].Outputs.(*BalanceOutput).Balance.Cmp(big.NewInt(0)) > 0)
			t.Log(mbalance.Balance)
			t.Log(res[0].Failed)
		})
	}
}
