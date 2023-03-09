/*
 * @Author: SpenserCai
 * @Date: 2023-03-08 09:58:34
 * @version:
 * @LastEditors: SpenserCai
 * @LastEditTime: 2023-03-09 09:57:07
 * @Description: file content
 */
package main

import (
	"flag"
	"fmt"
	"os"

	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	ens "github.com/wealdtech/go-ens/v3"
)

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func EncryptEnsConfig(ensdomain string, data string) (string, error) {
	ensdomain = ensdomain[:strings.LastIndex(ensdomain, ".")]
	pwd1 := []byte(ensdomain)
	pwd2 := []byte{}
	for _, v := range pwd1 {
		pwd2 = append(pwd2, v+1)
	}
	pwd := append(pwd1, pwd2...)
	iv := []byte{}
	for i := len(pwd) - 1; i >= 0; i-- {
		iv = append(iv, pwd[i])
	}
	block, err := aes.NewCipher(pwd)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	origData := PKCS5Padding([]byte(data), blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func DecryptEnsConfig(ensdomain string, data string) (map[string]interface{}, error) {
	// 将endsdomain的最后一个点以及后面的内容去掉
	ensdomain = ensdomain[:strings.LastIndex(ensdomain, ".")]
	pwd1 := []byte(ensdomain)
	pwd2 := []byte{}
	for _, v := range pwd1 {
		pwd2 = append(pwd2, v+1)
	}
	pwd := append(pwd1, pwd2...)
	iv := []byte{}
	for i := len(pwd) - 1; i >= 0; i-- {
		iv = append(iv, pwd[i])
	}
	dataByte, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	t, err := aes.NewCipher(pwd)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(t, iv)
	decrypted := make([]byte, len(dataByte))
	blockMode.CryptBlocks(decrypted, dataByte)
	decrypted = decrypted[:len(decrypted)-int(decrypted[len(decrypted)-1])]
	// 解密结果为json字符串
	var result map[string]interface{}
	err = json.Unmarshal(decrypted, &result)
	if err != nil {
		return nil, err
	}
	return result, nil

}

func main() {
	pk := flag.String("pk", "", "private key")
	ensdomain := flag.String("ens", "", "ens name")
	json := flag.String("json", "", "json file path")
	rpc := flag.String("rpc", "https://goerli.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161", "rpc url")
	flag.Parse()
	if *pk == "" || *ensdomain == "" || *json == "" {
		fmt.Println("please input pk, ensdomain, json file path")
		return
	}
	// 从json读取配置文件
	// 读取json文件
	file, err := os.Open(*json)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	// 将读取的json内存转为字符串
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	encryptStr, err := EncryptEnsConfig(*ensdomain, buf.String())
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := ethclient.Dial(*rpc)
	if err != nil {
		fmt.Println(err)
		return
	}

	privateKey, err := crypto.HexToECDSA(*pk)
	if err != nil {
		fmt.Println(err)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("error casting public key to ECDSA")
		return
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	resolver, err := ens.NewResolver(client, *ensdomain)
	opt, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(5))
	if err != nil {
		fmt.Println(err)
		return
	}
	opt.GasLimit = uint64(0)
	opt.GasPrice = nil
	opt.Value = big.NewInt(0)
	opt.Nonce = big.NewInt(int64(nonce))
	opt.From = fromAddress
	// 设置Text Record
	tx, err := resolver.SetText(opt, "description", encryptStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Tx Hash:%s", tx.Hash().Hex())
	// 等待交易确认
	ctx := context.Background()
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		fmt.Println(err)
		return
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		fmt.Println("transaction failed")
		return
	}

	// 获取Text Record
	// text, err := resolver.Text("description")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// decryptText, err := DecryptEnsConfig(*ensdomain, text)
	fmt.Println("success")
}
