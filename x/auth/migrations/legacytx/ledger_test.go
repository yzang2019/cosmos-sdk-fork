package legacytx_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/stretchr/testify/require"
)

type txJsonTest struct {
	Name       string
	Proto      json.RawMessage
	SignerData signing.SignerData `json:"signer_data"`
	Metadata   *bankv1beta1.Metadata
	Error      bool
}

type ledgerTestCase struct {
	Name string
	Tx   json.RawMessage
}

func TestLedger(t *testing.T) {
	encCfg := testutil.MakeTestEncodingConfig(bank.AppModuleBasic{}, authz.AppModuleBasic{}, gov.AppModuleBasic{})
	// Register the dummy types used in the tx extension options.
	encCfg.InterfaceRegistry.RegisterImplementations((*tx.ExtensionOptionI)(nil), &sdk.Coin{}, &authtypes.Params{})

	rawTxs, err := os.ReadFile("../../../../tx/textual/internal/testdata/tx.json")
	require.NoError(t, err)
	var txTestcases []txJsonTest
	err = json.Unmarshal(rawTxs, &txTestcases)
	require.NoError(t, err)

	raw, err := os.ReadFile("./ledger.json")
	require.NoError(t, err)
	var testcases []ledgerTestCase
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			// Find the matching proto Tx from tx.json
			var txTestCase txJsonTest
			for _, t := range txTestcases {
				if strings.HasPrefix(tc.Name, t.Name) {
					txTestCase = t
				}
			}

			tx, err := encCfg.TxConfig.TxJSONDecoder()(txTestCase.Proto)
			require.NoError(t, err)

			sigTx := tx.(signing.Tx)

			signDoc := legacytx.StdSignBytes("my-chain", 1, 2, sigTx.GetTimeoutHeight(), legacytx.StdFee{
				Amount:  sigTx.GetFee(),
				Gas:     sigTx.GetGas(),
				Payer:   sigTx.FeeGranter().String(),
				Granter: sigTx.FeeGranter().String(),
			}, sigTx.GetMsgs(), sigTx.GetMemo(), sigTx.GetTip())
			buffer := new(bytes.Buffer)
			err = json.Compact(buffer, tc.Tx)
			require.NoError(t, err)

			require.Equal(t, buffer.Bytes(), signDoc)
		})
	}
}
