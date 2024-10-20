package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"log"
	"net"

	pb "Primary-Node/generated"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPrimaryNodeServer
}

func (s *server) SendStatus(ctx context.Context, req *pb.DigimonStatus) (*pb.Response, error) {
	// Decodificar el mensaje en hexadecimal
	decodedMsg, err := hex.DecodeString(req.DigimonEncrypt)
	if err != nil {
		return nil, fmt.Errorf("error decoding message: %v", err)
	}

	// Desencriptar el mensaje
	decryptedMsg, err := decryptMessage(decodedMsg, []byte("keyregionalesprimarynode")) // Usa la misma clave que usaste para encriptar
	if err != nil {
		return nil, fmt.Errorf("error decrypting message: %v", err)
	}

	// Procesa el mensaje aquí
	log.Printf("Received decrypted digimon status: %s", decryptedMsg)
	return &pb.Response{Message: "Data received successfully"}, nil
}

func decryptMessage(ciphertext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCFBDecrypter(block, iv)
	mode.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterPrimaryNodeServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
