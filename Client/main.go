package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var MANUAL_KEY = []byte("COMP")

const (
	lossProbabilityPercent  = 20
	errorProbabilityPercent = 15
)

func checksum(s string) int {
	sum := 0
	for _, c := range s {
		sum += int(c)
	}
	return sum % 256
}

func manualEncrypt(payload string) string {

	p := []byte(payload)

	encrypted := make([]byte, 4)

	for i := 0; i < 4; i++ {
		encrypted[i] = p[i] ^ MANUAL_KEY[i]
	}

	final := make([]byte, 4)

	final[0] = encrypted[2]
	final[1] = encrypted[3]
	final[2] = encrypted[0]
	final[3] = encrypted[1]

	return hex.EncodeToString(final)
}

func buildPacket(seq int, total int, fragment string, corrupt bool) string {
	payload := fmt.Sprintf("%-4s", fragment)
	cs := checksum(payload)

	if corrupt {
		cs = (cs + 1) % 256
	}

	return fmt.Sprintf(
		"DATA|%d|%d|%s|%d",
		seq,
		total,
		manualEncrypt(payload),
		cs,
	)
}

func sendPacket(conn *net.UDPConn, rng *rand.Rand, seq int, total int, fragment string, reason string) {
	if rng.Intn(100) < lossProbabilityPercent {
		fmt.Printf("PERDA SIMULADA seq %d (%s)\n", seq, reason)
		return
	}

	corrupt := rng.Intn(100) < errorProbabilityPercent
	packet := buildPacket(seq, total, fragment, corrupt)

	conn.Write([]byte(packet))

	if corrupt {
		fmt.Printf("ERRO SIMULADO seq %d (%s)\n", seq, reason)
	} else {
		fmt.Printf("Enviado seq %d (%s)\n", seq, reason)
	}
}

func main() {

	server, _ := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	conn, _ := net.DialUDP("udp", nil, server)

	reader := bufio.NewReader(os.Stdin)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Print("Modo (gobackn/selecionado): ")
	modo, _ := reader.ReadString('\n')
	modo = strings.TrimSpace(modo)

	conn.Write([]byte("HELLO|30|" + modo))

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)

	fmt.Println("Servidor:", string(buffer[:n]))

	fmt.Print("Mensagem (>=30): ")
	msg, _ := reader.ReadString('\n')
	msg = strings.TrimSpace(msg)

	fragSize := 4
	var fragments []string

	for i := 0; i < len(msg); i += fragSize {

		end := i + fragSize
		if end > len(msg) {
			end = len(msg)
		}

		fragments = append(fragments, msg[i:end])
	}

	total := len(fragments)

	base := 0
	nextSeq := 0
	window := 5

	acks := make(map[int]bool)

	for base < total {

		for nextSeq < total && nextSeq < base+window {
			sendPacket(conn, rng, nextSeq, total, fragments[nextSeq], "envio")
			nextSeq++
		}

		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, err := conn.Read(buffer)

		if err != nil {
			if modo == "gobackn" {
				fmt.Println("Timeout -> Go-Back-N retransmitindo a partir da base", base)
				nextSeq = base
			} else {
				fmt.Println("Timeout -> Selective Repeat retransmitindo apenas pendentes")
				for seq := base; seq < nextSeq; seq++ {
					if !acks[seq] {
						sendPacket(conn, rng, seq, total, fragments[seq], "retransmissao")
					}
				}
			}
			continue
		}

		resp := strings.Split(string(buffer[:n]), "|")

		if resp[0] == "ACK" {

			ack, _ := strconv.Atoi(resp[1])
			fmt.Println("ACK", ack)

			acks[ack] = true

			for acks[base] {
				base++
			}

		} else if resp[0] == "NAK" {

			nak, _ := strconv.Atoi(resp[1])
			fmt.Println("NAK", nak)

			if modo == "gobackn" {
				base = nak
				nextSeq = nak
			} else if nak >= 0 && nak < total && !acks[nak] {
				sendPacket(conn, rng, nak, total, fragments[nak], "retransmissao por NAK")
			}
		}
	}

	fmt.Println("Mensagem entregue com sucesso!")
}
