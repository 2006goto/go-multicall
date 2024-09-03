package multicall

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCall_BadABI(t *testing.T) {
	r := require.New(t)

	const oneValueABI = `[
		{
			"constant":true,
			"inputs": [
				{
					"name":"val1",
					"type":"bool"
				}
			],
			"name":"testFunc",
			"outputs": [
				{
					"name":"val1",
					"type":"bool"
				}
			],
			"payable":false,
			"stateMutability":"view",
			"type":"function"
		}
	` // missing closing ] at the end

	_, err := NewContract(oneValueABI, "0x")
	r.Error(err)
	r.ErrorContains(err, "unexpected EOF")
}

const ERC20ABI = `[
		{
			"constant":true,
			"inputs":[
					{
						"name":"tokenOwner",
						"type":"address"
					}
			],
			"name":"balanceOf",
			"outputs":[
					{
						"name":"balance",
						"type":"uint256"
					}
			],
			"payable":false,
			"stateMutability":"view",
			"type":"function"
		},
                {
                        "constant": true,
                        "inputs": [],
                        "name": "decimals",
                        "outputs": [
                                       {
                                                "name": "",
                                                "type": "uint8"
                                       }
                        ],
                        "payable": false,
                        "stateMutability": "view",
                        "type": "function"
                }
	]`

type BalanceOutput struct {
	Balance *big.Int
}

func TestCall_BadContract(t *testing.T) {
	ctx := context.Background()
	rpcURL := "https://cloudflare-eth.com"
	abiERC20, err := ParseABI(ERC20ABI)
	if err != nil {
		t.Fatal(err)
	}
	caller, err := Dial(ctx, rpcURL)
	if err != nil {
		t.Fatal(err)
	}
	var calls []*Call

	notContract := &Contract{
		ABI:     abiERC20,
		Address: common.HexToAddress("0x0000000000000000000000000000000000000000"),
	}
	notContractBalance := new(BalanceOutput)
	calls = append(calls, notContract.NewCall(
		notContractBalance,
		"balanceOf",
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
	).AllowFailure())

	invalidContract := &Contract{
		ABI:     abiERC20,
		Address: common.HexToAddress("0x93A4009DD030C34d9aa45e00a29990c5352918B1"),
	}
	invalidBalance := new(BalanceOutput)
	calls = append(calls, invalidContract.NewCall(
		invalidBalance,
		"balanceOf",
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
	).AllowFailure())

	validContract := &Contract{
		ABI:     abiERC20,
		Address: common.HexToAddress("0xfd1450a131599ff34f3be1775d8c8bf79e353d8c"),
	}
	validBalance := new(BalanceOutput)
	calls = append(calls, validContract.NewCall(
		validBalance,
		"balanceOf",
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
	).AllowFailure())

	res, err := caller.Call(nil, calls...)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, true, res[0].Failed)
	require.Equal(t, true, res[1].Failed)
	require.Equal(t, false, res[2].Failed)
}

func TestCall_ExoticChains(t *testing.T) {
	ctx := context.Background()
	rpcURL := "https://1rpc.io/zksync2-era"
	abiERC20, err := ParseABI(ERC20ABI)
	if err != nil {
		t.Fatal(err)
	}
	multicallAddr := "0x47898B2C52C957663aE9AB46922dCec150a2272c"
	caller, err := Dial(ctx, rpcURL, multicallAddr)
	// caller, err := Dial(ctx, rpcURL)
	if err != nil {
		t.Fatal(err)
	}
	var calls []*Call

	notContract := &Contract{
		ABI:     abiERC20,
		Address: common.HexToAddress("0x1d17CBcF0D6D143135aE902365D2E5e2A16538D4"),
	}
	notContractBalance := new(BalanceOutput)
	calls = append(calls, notContract.NewCall(
		notContractBalance,
		"balanceOf",
		common.HexToAddress("0x000000000000000000000000000000000000dEaD"),
	).AllowFailure())

	res, err := caller.Call(nil, calls...)
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, res[0].Failed)
	assert.True(t, res[0].Outputs.(*BalanceOutput).Balance.Cmp(big.NewInt(0)) > 0)
	t.Log(notContractBalance.Balance)
	t.Log(res[0].Failed)
}
