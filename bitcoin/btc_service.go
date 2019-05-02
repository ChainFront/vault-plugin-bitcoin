package bitcoin

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
)

// GetCurrentFee gets the current fee in bitcoin
func GetCurrentFee() (float64, error) {
	req := struct {
		ID     int    `json:"id"`
		Method string `json:"method"`
		Params []int  `json:"params"`
	}{
		ID:     1,
		Method: "blockchain.estimatefee",
		Params: []int{2},
	}

	msg := struct {
		JSONRPC string  `json:"jsonrpc,omitempty"`
		ID      int     `json:"id"`
		Result  float64 `json:"result"`
	}{}

	var fee float64
	var MaxTries = 5
	for try := 0; try < MaxTries; try++ {
		sendMsg(req, &msg)

		if msg.Result == -1.0 || msg.Result == 0 {
			log.Printf("expected result > 0; received: %f", msg.Result)
			continue
		}

		fee = msg.Result
		// sanity check
		if fee > 0.05 {
			fee = 0.1
		} else if fee < 0 {
			fee = 0
		}

		break
	}

	fmt.Printf("fee: %f\n", fee)

	if fee == 0 {
		log.Print("could not get fees")
		return fee, errors.New("could not get fees")
	}

	return fee, nil
}

// GetCurrentFeeRate gets the current fee in satoshis per kb
func GetCurrentFeeRate() (*big.Int, error) {
	fee, err := GetCurrentFee()
	if err != nil {
		return nil, err
	}

	// convert to satoshis to kb
	feeRate := big.NewInt(int64(fee * 1.0E5))

	fmt.Printf("fee rate: %s\n", feeRate)

	return feeRate, nil
}

func sendMsg(req, res interface{}) {
	//serverAddr := "localhost:3001"
	serverAddr := "testnet.hsmiths.com:53012"

	fmt.Printf("dialing to server: %s\n", serverAddr)
	conn, err := tls.Dial("tcp", serverAddr, &tls.Config{
		//Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	fmt.Printf("client connected to: %s\n", conn.RemoteAddr())

	reqMsgBytes, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	reqMsg := fmt.Sprintf("%s\n", string(reqMsgBytes))
	fmt.Printf("writing message: %s", reqMsg)
	_, err = io.WriteString(conn, reqMsg)
	if err != nil {
		log.Fatal(err)
	}

	var (
		i        int
		readSize int = 1024
		respData []byte
	)

	for {
		fmt.Println("reading response...")
		respBytes := make([]byte, readSize)
		n, err := conn.Read(respBytes)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
		}

		fmt.Printf("reading: %q (%d bytes)\n", string(respBytes[:n]), n)

		respData = append(respData, respBytes[:n]...)
		i += n

		if n < readSize {
			break
		}
	}

	json.Unmarshal(respData[:i], &res)
}
