package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"bufio"

	pb "Primary-Node/generated"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedPrimaryNodeServer
	mu      sync.Mutex
    counter int64
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

	digimonInfo := strings.Split(decryptedMsg, ",")

	if(!SearchFile(digimonInfo[0])){
		name := strings.ToLower(digimonInfo[0])
		var numDataNode int

		if int(name[0]) >= 97 && int(name[0]) < 110 {
			numDataNode = 1
		} else if (int(name[0]) >= 110 && int(name[0]) < 123){
			numDataNode = 2
		}else{
			return &pb.Response{Message: "Data received successfully"}, nil
		}

		s.mu.Lock()
		id := s.counter
		s.counter++
		s.mu.Unlock()

		writeRecord(digimonInfo, numDataNode, id)
	}
	return &pb.Response{Message: "Data received successfully"}, nil
}

func writeRecord(DigimonData []string, NumDataNode int, id int64){
	file, err := os.OpenFile("INFO.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Printf("error opening file: %v", err)
		return 
    }
    defer file.Close()

	record := fmt.Sprintf("%d,%d,%s,%s\n", id, NumDataNode, DigimonData[0], DigimonData[2])

	if _, err := file.WriteString(record); err != nil {
        log.Printf("error writing to file: %v", err)
        return
    }

    log.Println("Registro escrito correctamente")
}

func SearchFile(name string) bool {
    filename := "INFO.txt"

    // Abrir el archivo en modo lectura
    file, err := os.Open(filename)
    if err != nil {
        log.Fatalf("Error opening file: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    // Recorrer el archivo línea por línea
    for scanner.Scan() {
        line := scanner.Text()
        fields := strings.Split(line, ",")

        // Verificar si el nombre es igual al buscado
        if len(fields) > 2 && strings.TrimSpace(fields[2]) == name {
            return true
        }
    }

    // Si no se encontró el nombre, retornar false
    if err := scanner.Err(); err != nil {
        log.Fatalf("Error reading file: %v", err)
    }

    return false
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

func writeToFile(filename string, data string) error {
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    // Escribir los datos en el archivo
    if _, err := file.WriteString(data + "\n"); err != nil {
        return err
    }
    return nil
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
