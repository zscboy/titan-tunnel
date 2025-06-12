package wallet

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestWallet(t *testing.T) {
	wallet, err := NewWallet(
		WithChainID("titan"),
		WithAccountPrefix("titan"),
		WithKeyringBackend(keyring.BackendOS),
		WithKeyDirectory("D://codes//titan-agent-private//common"),
	)
	if err != nil {
		log.Fatalf("Create wallet failed: %v", err)
	}

	ks, err := wallet.ListKeys()
	if err != nil {
		log.Fatalf("list keys: %v", err)
	}

	for _, key := range ks {
		// fmt.Printf("%s:%s\n", key.GetName(), key.GetAddress().String())
		wallet.DeleteKey(key.GetName())
	}

	key, err := wallet.AddKey("abc", sdk.CoinType)
	if err != nil {
		log.Fatalf("AddKey: %v", err)
	}

	fmt.Printf("Wallet Address:%s\n", key.Address)
	fmt.Printf("Mnemonic:%s\n", key.Mnemonic)

	testData := []byte("hello, cosmos")
	sig, err := wallet.Sign("abc", testData)
	if err != nil {
		log.Fatalf("Sign failed: %v", err)
	}

	fmt.Printf("sign: %s\n", hex.EncodeToString(sig))

	pubKey, err := wallet.GetPubKey("abc")
	if err != nil {
		log.Fatalf("GetPubKey: %v", err)
	}

	fmt.Println("Public key hex:", hex.EncodeToString(pubKey.Bytes()))
	pubKeyBytes := pubKey.Bytes()
	var newPubKey = secp256k1.PubKey(pubKeyBytes)

	if newPubKey.VerifySignature(testData, sig) {
		fmt.Println("verify sign success")
	} else {
		fmt.Println("verify sign failed")
	}

	address := sdk.AccAddress(newPubKey.Address())
	_, err = sdk.Bech32ifyAddressBytes("titan", address)
	if err != nil {
		log.Fatalf("Bech32ifyAddressBytes: %v", err)
	}

	encryptedKey, err := wallet.ExportPrivKeyArmor("abc")
	if err != nil {
		log.Fatalf("Bech32ifyAddressBytes: %v", err)
	}

	privKey, algo, err := crypto.UnarmorDecryptPrivKey(encryptedKey, "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("算法: %v\n", algo)                    // 输出: secp256k1
	fmt.Printf("原始私钥 (HEX): %X\n", privKey.Bytes()) // 32字节私钥

}

func TestPubkeyToAddress(t *testing.T) {
	pubkey := "023d1515c5e35228bef24d09cc8fbc165b911b14c5ced27b64315efca04e1a2ce1"
	pubkeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		t.Errorf("err:%s", err.Error())
		return
	}
	pubKey := secp256k1.PubKey(pubkeyBytes)
	address := types.AccAddress(pubKey.Address().Bytes())
	t.Logf("addrss:%s", address.String())
}
