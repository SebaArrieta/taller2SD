package main

import (
	pb "Isla-File/generated/Regionales"
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
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"errors"

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

type server struct {
	pb.UnimplementedPrimaryNodeServer
	stopChan chan struct{}
}

func (s *server) FinishRegionales(ctx context.Context, req *pb.FinishRegionalesRequest) (*pb.FinishRegionalesResponse, error) {
	log.Println("Finalizando Isla File...")
	// Enviar señal para cerrar el servidor
	s.stopChan <- struct{}{}
	return &pb.FinishRegionalesResponse{Resp: 1}, nil
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
	message := fmt.Sprintf("%s,%s,%s,%s", digimon.Name, digimon.Attribute, digimon.Status, "Isla-File")
	encryptedMessage, err := encryptMessage(message, aesKey)
	if err != nil {
		log.Fatalf("Error encrypting message: %v", err)
	}
	err = sendDigimonStatus(client, encryptedMessage)
	if err != nil {
		log.Fatalf("Error sending Digimon status: %v", err)
	}
	fmt.Printf("[Isla File] Estado enviado: %s %s\n", digimon.Name, digimon.Status)
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

/*func connectToPrimaryNode(addr string) (*grpc.ClientConn, pb.PrimaryNodeClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo conectar: %v", err)
	}
	client := pb.NewPrimaryNodeClient(conn)
	return conn, client, nil
}*/

func connectToPrimaryNode(addr string, maxRetries int) (*grpc.ClientConn, pb.PrimaryNodeClient, error) {
	var conn *grpc.ClientConn
	var err error

	// Retry loop with a maximum number of retries
	for i := 0; i < maxRetries; i++ {
		// Attempt to connect to the Primary Node
		conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			// Connection successful, break the loop
			client := pb.NewPrimaryNodeClient(conn)
			return conn, client, nil
		}

		// Log error and retry after 5 seconds
		log.Printf("Failed to connect to Primary Node at %s: %v. Retrying in 5 seconds... (Attempt %d/%d)", addr, err, i+1, maxRetries)
		time.Sleep(5 * time.Second) // Wait 5 seconds before retrying
	}

	// Return error if unable to connect after max retries
	return nil, nil, errors.New("failed to connect to Primary Node after max retries")
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

	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("Error al iniciar Isla File: %v", err)
	}
	stopChan := make(chan struct{})

	s := grpc.NewServer()
	pb.RegisterPrimaryNodeServer(s, &server{stopChan: stopChan})

	go func() {
		addr := "dist101:50051" // Replace with actual IP

		// Connect to the Primary Node
		conn, client, err := connectToPrimaryNode(addr, 5)
		if err != nil {
			log.Fatalf("%v", err)
		}
		defer conn.Close()

		encryptionKey := []byte("keyregionalesprimarynode") // 32-byte AES key
		digimonSender(digimons, client, encryptionKey)
	}()

	go func() {
		<-stopChan
		log.Println("Deteniendo el servidor gRPC...")
		s.GracefulStop()
		os.Exit(0)
	}()

	log.Println("Isla File corriendo en el puerto :50056")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir Isla File: %v", err)
	}

}
