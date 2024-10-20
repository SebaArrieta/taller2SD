package main

import (
	pb "Continente-Folder/generated"
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var PS, TE, TD, CD, VI float64

// Digimon structure
type Digimon struct {
	Name      string
	Attribute string
	Status    string
}

func shuffleDigimons(slice []Digimon) {
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(slice), func(i, j int) { slice[i], slice[j] = slice[j], slice[i] })
}

func loadDigimons(path string) ([]Digimon, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var digimons []Digimon
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		digimons = append(digimons, Digimon{Name: parts[0], Attribute: parts[1], Status: determineSacrifice()})
	}
	shuffleDigimons(digimons)
	return digimons, scanner.Err()
}

func encryptMessage(message string, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	b := []byte(message)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		panic(err)
	}
	mode := cipher.NewCFBEncrypter(block, iv)
	mode.XORKeyStream(ciphertext[aes.BlockSize:], b)
	return ciphertext, nil
}

func sendDigimonStatus(client pb.PrimaryNodeClient, digimonEncryptMsg []byte) error {
	encodedMsg := hex.EncodeToString(digimonEncryptMsg)
	req := &pb.DigimonStatus{DigimonEncrypt: encodedMsg}
	_, err := client.SendStatus(context.Background(), req)
	return err
}

func determineSacrifice() string {
	if mrand.Float64() < PS {
		return "Sacrificado"
	}
	return "No-Sacrificado"
}

func runRegionalServer(digimons []Digimon, client pb.PrimaryNodeClient, aesKey []byte) []Digimon {
	digimon := digimons[0]
	digimons = digimons[1:]
	message := fmt.Sprintf("%s,%s,%s", digimon.Name, digimon.Attribute, digimon.Status)
	encryptedMessage, err := encryptMessage(message, aesKey)
	if err != nil {
		log.Fatalf("Error encrypting message: %v", err)
	}
	err = sendDigimonStatus(client, encryptedMessage)
	if err != nil {
		log.Fatalf("Error sending Digimon status: %v", err)
	}
	fmt.Printf("[Continente Folder] Estado enviado: %s %s\n", digimon.Name, digimon.Status)
	return digimons
}

func digimonSender(digimons []Digimon, client pb.PrimaryNodeClient, aesKey []byte) {
	for i := 0; i < 6; i++ {
		if len(digimons) == 0 {
			return
		}
		digimons = runRegionalServer(digimons, client, aesKey)
	}
	for i := 0; i < len(digimons); i++ {
		digimons = runRegionalServer(digimons, client, aesKey)
		time.Sleep(time.Duration(TE) * time.Second)
	}
}

func connectToPrimaryNode(addr string) (*grpc.ClientConn, pb.PrimaryNodeClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo conectar: %v", err)
	}
	client := pb.NewPrimaryNodeClient(conn)
	return conn, client, nil
}

func readVariables(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		arr := strings.Split(line, ",")
		PS, _ = strconv.ParseFloat(arr[0], 64)
		TE, _ = strconv.ParseFloat(arr[1], 64)
		TD, _ = strconv.ParseFloat(arr[2], 64)
		CD, _ = strconv.ParseFloat(arr[3], 64)
		VI, _ = strconv.ParseFloat(arr[4], 64)
	}
}

func main() {
	readVariables("INPUT.txt")
	digimons, err := loadDigimons("DIGIDATA.txt")
	if err != nil {
		log.Fatalf("Error loading Digimons: %v", err)
	}
	addr := "localhost:50051" // Replace with actual IP

	// Connect to the Primary Node
	conn, client, err := connectToPrimaryNode(addr)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer conn.Close()

	encryptionKey := []byte("keyregionalesprimarynode") // 32-byte AES key
	digimonSender(digimons, client, encryptionKey)
}
