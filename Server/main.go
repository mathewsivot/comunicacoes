package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var MANUAL_KEY = []byte("COMP")

const windowSize = 5

type ClientState struct {
	total       int
	received    map[int]string
	expectedSeq int
	modo        string
}

func checksum(s string) int {
	sum := 0
	for _, c := range s {
		sum += int(c)
	}
	return sum % 256
}

func manualDecrypt(encryptedHex string) (string, error) {

	encryptedBytes, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}
	if len(encryptedBytes) != 4 {
		return "", fmt.Errorf("payload criptografado deve ter 4 bytes")
	}

	decryptedSub := make([]byte, 4)

	decryptedSub[0] = encryptedBytes[2]
	decryptedSub[1] = encryptedBytes[3]
	decryptedSub[2] = encryptedBytes[0]
	decryptedSub[3] = encryptedBytes[1]

	original := make([]byte, 4)

	for i := 0; i < 4; i++ {
		original[i] = decryptedSub[i] ^ MANUAL_KEY[i]
	}

	return string(original), nil
}

func main() {

	addr, _ := net.ResolveUDPAddr("udp", ":10000")
	conn, _ := net.ListenUDP("udp", addr)

	fmt.Println("[SERVIDOR] ouvindo em 0.0.0.0:10000")

	clients := make(map[string]*ClientState)

	buffer := make([]byte, 4096)

	for {

		n, clientAddr, _ := conn.ReadFromUDP(buffer)

		arrival := time.Now().Format("15:04:05")
		msg := string(buffer[:n])
		parts := strings.Split(msg, "|")

		key := clientAddr.String()

		switch parts[0] {

		case "HELLO":
			if len(parts) < 3 {
				conn.WriteToUDP([]byte("NAK|0"), clientAddr)
				continue
			}

			maxlen := parts[1]
			modo := parts[2]

			fmt.Printf("[%s] HELLO %s MAXLEN=%s MODO=%s\n",
				arrival, key, maxlen, modo)

			conn.WriteToUDP(
				[]byte(fmt.Sprintf("HELLO_ACK|%d", windowSize)),
				clientAddr,
			)

			clients[key] = &ClientState{
				total:       -1,
				received:    map[int]string{},
				expectedSeq: 0,
				modo:        modo,
			}

		case "DATA":

			if len(parts) < 5 {
				conn.WriteToUDP([]byte("NAK|0"), clientAddr)
				continue
			}

			seq, err := strconv.Atoi(parts[1])
			if err != nil {
				conn.WriteToUDP([]byte("NAK|0"), clientAddr)
				continue
			}

			total, err := strconv.Atoi(parts[2])
			if err != nil || total <= 0 {
				conn.WriteToUDP([]byte(fmt.Sprintf("NAK|%d", seq)), clientAddr)
				continue
			}

			payloadEnc := parts[3]
			recvCS, err := strconv.Atoi(parts[4])
			if err != nil {
				conn.WriteToUDP([]byte(fmt.Sprintf("NAK|%d", seq)), clientAddr)
				continue
			}

			state := clients[key]
			if state == nil {
				fmt.Printf("[%s] DATA rejeitado de %s: HELLO nao realizado\n", arrival, key)
				conn.WriteToUDP([]byte("NAK|0"), clientAddr)
				continue
			}

			payloadPad, err := manualDecrypt(payloadEnc)

			if err != nil {
				conn.WriteToUDP(
					[]byte(fmt.Sprintf("NAK|%d", seq)),
					clientAddr,
				)
				continue
			}

			if checksum(payloadPad) != recvCS {
				fmt.Println("Erro checksum", seq)
				conn.WriteToUDP(
					[]byte(fmt.Sprintf("NAK|%d", seq)),
					clientAddr,
				)
				continue
			}

			payload := strings.TrimRight(payloadPad, " ")

			if state.total == -1 {
				state.total = total
			}

			if state.modo == "gobackn" {

				if seq == state.expectedSeq {

					state.received[seq] = payload
					state.expectedSeq++

					conn.WriteToUDP(
						[]byte(fmt.Sprintf("ACK|%d", seq)),
						clientAddr,
					)

				} else if seq < state.expectedSeq {

					conn.WriteToUDP(
						[]byte(fmt.Sprintf("ACK|%d", seq)),
						clientAddr,
					)

				} else {

					conn.WriteToUDP(
						[]byte(fmt.Sprintf("NAK|%d", state.expectedSeq)),
						clientAddr,
					)
				}

			} else { // Selective Repeat

				state.received[seq] = payload

				conn.WriteToUDP(
					[]byte(fmt.Sprintf("ACK|%d", seq)),
					clientAddr,
				)
			}

			if len(state.received) == state.total {

				full := ""

				for i := 0; i < state.total; i++ {
					full += state.received[i]
				}

				fmt.Println("================================")
				fmt.Println("MENSAGEM COMPLETA:")
				fmt.Println(full)
				fmt.Println("================================")

				clients[key] = &ClientState{
					total:       -1,
					received:    map[int]string{},
					expectedSeq: 0,
					modo:        state.modo,
				}
			}
		}
	}
}
