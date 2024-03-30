package routers

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"database/sql"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	_ "github.com/go-sql-driver/mysql"
)

var nilData = "[{\"data\": \"nil\"}]"

func GetPublicKey(X, Y string) *ecdsa.PublicKey {
	InitiatorPublicKey := new(ecdsa.PublicKey)
	InitiatorPublicKey.X, _ = new(big.Int).SetString(X, 10)
	InitiatorPublicKey.Y, _ = new(big.Int).SetString(Y, 10)
	InitiatorPublicKey.Curve = elliptic.P256()
	return InitiatorPublicKey
}

func generateRandomMessage() string {
	b := make([]byte, 10)
	rand.Read(b)
	randomMessage := fmt.Sprintf("%x", b)
	return randomMessage
}

func getOuterIP() (ipv4 string, err error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(ip)), nil
}

func SignMessage(message string, signer crypto.Signer) (string, error) {
	hash := sha256.Sum256([]byte(message))

	signature, err := signer.Sign(rand.Reader, hash[:], crypto.SHA256)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	signatureStr := base64.StdEncoding.EncodeToString(signature)
	return signatureStr, nil
}

func verifySignature(message, signature string, publicKey *ecdsa.PublicKey) error {
	decodedSignature, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	hash := sha256.Sum256([]byte(message))

	// 将 *ecdsa.PublicKey 转换为 crypto.PublicKey
	if !ecdsa.VerifyASN1(publicKey, hash[:], decodedSignature) {
		return fmt.Errorf("invalid signature")
	}
	fmt.Println("VerifySignature successfully verify a signature")
	return nil
}

// Login by verifying the signature
func (r *Routers) execVerify(referer string, InitiatorPublicKey *ecdsa.PublicKey) (bool, error) {
	// 1. 生成随机message
	randomMessage := generateRandomMessage()
	fmt.Println("execVerify generate random message: ", randomMessage)

	// 2. 将随机message发送给申请方的receive_message接口
	// 2.1 获取对方的url
	newUrl := referer + "/receive_message"
	// 2.2 构造payload
	payload := map[string]interface{}{
		"message":      randomMessage,
		"PublisherUrl": r.MyURL,
	}
	fmt.Println("execVerify generate payload: ", payload)
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("execVerify json.Marshal err: ", err)
		return false, err
	}
	// 2.3 send message
	req, err := http.NewRequest(http.MethodPost, newUrl, bytes.NewReader(body))
	if err != nil {
		fmt.Println("execVerify http.NewRequest err: ", err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("execVerify http.DefaultClient.Do() err:", err)
		return false, err
	}
	defer resp.Body.Close()

	// 4.等待签名后的结果
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("execVerify ioutil.ReadAll() err: ", err)
		return false, err
	}

	var respData map[string]interface{}
	err = json.Unmarshal(respBody, &respData)
	fmt.Println("respData:", respData)
	if err != nil {
		fmt.Println("execVerify unmarshal response err: ", err)
		return false, err
	}

	signedMessage, ok := respData["signature"].(string)
	fmt.Println("execVerify receive signature: ", signedMessage)
	if !ok {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "invalid response format"})
		return false, err
	}

	// 5. 验证签名
	fmt.Println("execVerify initiator pub key:", InitiatorPublicKey)
	// InitPubKey, err := parseECDSAPublicKey(InitiatorPublicKey)
	if err != nil {
		fmt.Println("execVerify parseECDSAPublicKey err: ", err)
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false, err
	}
	verifiedErr := verifySignature(randomMessage, signedMessage, InitiatorPublicKey)
	if verifiedErr == nil {
		return true, nil
	} else {
		return false, err
	}
}

func (r *Routers) dataBase(usrname string, passwd string, ip string, port string, databaseName string, tableName string) string {
	dsn := usrname + ":" + passwd + "@tcp(" + ip + ":" + port + ")/" + databaseName
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("dsn:%s invalid,err:%v\n", dsn, err)
		return nilData
	}
	defer db.Close()
	err = db.Ping() //尝试连接数据库
	if err != nil {
		fmt.Printf("open %s faild,err:%v\n", dsn, err)
		return nilData
	}
	sqlStr := "select * from " + tableName + ";"
	rows, err := db.Query(sqlStr)
	if err != nil {
		fmt.Println(err)
	}
	// defer close result set
	defer rows.Close()

	cols, _ := rows.Columns()
	var ret []map[string]interface{}
	if len(cols) > 0 {
		for rows.Next() {
			buff := make([]interface{}, len(cols))
			data := make([][]byte, len(cols)) //数据库中的NULL值可以扫描到字节中
			for i := range buff {
				buff[i] = &data[i]
			}
			rows.Scan(buff...) //扫描到buff接口中，实际是字符串类型data中
			//将每一行数据存放到数组中
			dataKv := make(map[string]interface{}, len(cols))
			for k, col := range data { //k是index，col是对应的值
				//fmt.Printf("%30s:\t%s\n", cols[k], col)
				dataKv[cols[k]] = string(col)
			}
			ret = append(ret, dataKv)
		}
		retjson, _ := json.Marshal(ret)
		return string(retjson)
	} else {
		return nilData
	}
}

// ECIES 公钥数据加密
func (r *Routers) EnCryptByEcies(srcData string, public_key *ecdsa.PublicKey) (cryptData string, err error) {
	//获取公钥数据
	publicKey := ecies.ImportECDSAPublic(public_key)
	if err != nil {
		return "", err
	}

	//公钥加密数据
	encryptBytes, err := ecies.Encrypt(rand.Reader, publicKey, []byte(srcData), nil, nil)
	if err != nil {
		return "", err
	}

	cryptData = hex.EncodeToString(encryptBytes)

	return
}

// ECIES 私钥数据解密
func (r *Routers) DeCryptByEcies(cryptData string, private_Key *ecdsa.PrivateKey) (srcData string, err error) {
	//获取私钥信息
	privateKey := ecies.ImportECDSA(private_Key)

	//私钥解密数据
	cryptBytes, err := hex.DecodeString(cryptData)
	if err != nil {
		fmt.Println("解密错误：", err)
		return "", err
	}
	srcByte, err := privateKey.Decrypt(cryptBytes, nil, nil)
	if err != nil {
		fmt.Println("解密错误：", err)
		return "", err
	}
	srcData = string(srcByte)

	return
}
