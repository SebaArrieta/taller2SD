package main

import (
	"bufio"
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
	"time"

	pbDataNode "Primary-Node/generated/DataNode"
	pb "Primary-Node/generated/Regionales"
	pbTai "Primary-Node/generated/Tai"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pb.UnimplementedPrimaryNodeServer
	pbTai.UnimplementedTaiServer
	dataNode1Client        pbDataNode.DNodeClient
	dataNode2Client        pbDataNode.DNodeClient
	islaFileClient         pb.PrimaryNodeClient
	continenteFolderClient pb.PrimaryNodeClient
	continenteServerClient pb.PrimaryNodeClient

	mu       sync.Mutex
	counter  int64
	stopChan chan struct{}
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

	digimonInfo := strings.Split(decryptedMsg, ",")

	if !SearchFile(digimonInfo[0]) {
		name := strings.ToLower(digimonInfo[0])
		var numDataNode int

		if int(name[0]) >= 97 && int(name[0]) < 110 {
			numDataNode = 1
		} else if int(name[0]) >= 110 && int(name[0]) < 123 {
			numDataNode = 2
		} else {
			return &pb.Response{Message: "Data received successfully"}, nil
		}

		s.mu.Lock()
		id := s.counter
		s.counter++
		s.mu.Unlock()

		if writeRecord(digimonInfo, numDataNode, id) {
			DataNodeRecord := fmt.Sprintf("%d,%s", id, digimonInfo[1])
			sendToDataNode(numDataNode, DataNodeRecord, s)
		}
	}

	log.Printf("[PRIMARY NODE] Solicitud de %s recibida, mensaje enviado: Data del digimon recibida: %s", digimonInfo[3], decryptedMsg)
	return &pb.Response{Message: fmt.Sprintf("Data del digimon recibida: %s", decryptedMsg)}, nil
}

func (s *server) GetSacrificed(ctx context.Context, req *pbTai.Request) (*pbTai.Response, error) {
	file, err := os.Open("INFO.txt")
	if err != nil {
		log.Printf("Error al abrir el archivo: %v", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sacrificedIDs []string

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")

		if len(parts) == 4 && strings.TrimSpace(parts[3]) == "Sacrificado" {
			sacrificedIDs = append(sacrificedIDs, parts[0])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error leyendo el archivo: %v", err)
		return nil, err
	}

	result := strings.Join(sacrificedIDs, ";")

	data := getDataToDnode(result, s)
	accumulatedData, numData := computeData(data)
	log.Printf("[PRIMARY NODE] Solicitud de Nodo Tai recibida, mensaje enviado: {AccumulatedData: %d, SacrificedDigimons: %d}", float32(accumulatedData), int32(numData))
	return &pbTai.Response{AccumulatedData: float32(accumulatedData), SacrificedDigimons: int32(numData)}, nil
}

func (s *server) FinishTai(ctx context.Context, req *pbTai.FinishTaiRequest) (*pbTai.FinishTaiResponse, error) {
	log.Println("Tai node has finished. Shutting down Primary Node...")
	s.FinishDNodes(ctx, &pbDataNode.FinishDNodesRequest{})
	s.FinishRegionales(ctx, &pb.FinishRegionalesRequest{})
	// Enviar señal para cerrar el servidor
	s.stopChan <- struct{}{}

	return &pbTai.FinishTaiResponse{Resp: 1}, nil
}

func computeData(data string) (total float64, num int) {
	dataParts := strings.Split(data, ";")
	total = 0
	for _, part := range dataParts {
		if part == "Vaccine" {
			total += 3
		} else if part == "Data" {
			total += 1.5
		} else {
			total += 0.8
		}
	}
	return total, len(dataParts)
}

func sendToDataNode(dataNode int, record string, s *server) {
	ctxDataNode, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if dataNode == 1 {
		_, err := s.dataNode1Client.GetAtributo(ctxDataNode, &pbDataNode.Request{Message: record})
		if err != nil {
			log.Printf("Error al comunicarse con el Data Node: %v", err)
			return
		}
	} else if dataNode == 2 {
		_, err := s.dataNode2Client.GetAtributo(ctxDataNode, &pbDataNode.Request{Message: record})
		if err != nil {
			log.Printf("Error al comunicarse con el Data Node: %v", err)
			return
		}
	}
}

func writeRecord(digimonData []string, numDataNode int, id int64) bool {
	file, err := os.OpenFile("INFO.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("error opening file: %v", err)
		return false
	}
	defer file.Close()

	record := fmt.Sprintf("%d,%d,%s,%s\n", id, numDataNode, digimonData[0], digimonData[2])

	if _, err := file.WriteString(record); err != nil {
		log.Printf("error writing to file: %v", err)
		return false
	}

	log.Println("Registro escrito correctamente")
	return true
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

func getDataToDnode(ids string, s *server) (data string) {
	var wg sync.WaitGroup
	wg.Add(2)
	var results []string
	var res1, res2 *pbDataNode.Response
	var err1, err2 error

	// Goroutine para Data Node 1
	go func() {
		defer wg.Done()
		ctxDataNode, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res1, err1 = s.dataNode1Client.SendData(ctxDataNode, &pbDataNode.Request{Message: ids})
		if err1 != nil {
			log.Printf("Error al comunicarse con Data Node 1: %v", err1)
			return
		}
	}()

	// Goroutine para Data Node 2
	go func() {
		defer wg.Done()
		ctxDataNode, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res2, err2 = s.dataNode2Client.SendData(ctxDataNode, &pbDataNode.Request{Message: ids})
		if err2 != nil {
			log.Printf("Error al comunicarse con Data Node 2: %v", err2)
			return
		}
		log.Printf("Respuesta de Data Node 2: %s", res2.GetMessage())
	}()

	// Esperar a que ambas goroutines terminen
	wg.Wait()

	// Verificar respuestas antes de acceder a ellas
	if res1 != nil && res1.GetMessage() != "-1" {
		results = append(results, res1.GetMessage())
	} else {
		log.Println("La respuesta de Data Node 1 es nil o inválida")
	}

	if res2 != nil && res2.GetMessage() != "-1" {
		results = append(results, res2.GetMessage())
	} else {
		log.Println("La respuesta de Data Node 2 es nil o inválida")
	}

	data = strings.Join(results, ";")

	return data
}

func (s *server) FinishDNodes(ctx context.Context, req *pbDataNode.FinishDNodesRequest) (*pbDataNode.FinishDNodesResponse, error) {
	// Crear un contexto con timeout para cada llamada gRPC
	ctxDataNode, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Enviar solicitud de finalización a Data Node 1
	_, err := s.dataNode1Client.FinishDNodes(ctxDataNode, &pbDataNode.FinishDNodesRequest{Req: 1})
	if err != nil {
		log.Printf("Error al finalizar Data Node 1: %v", err)
	} else {
		log.Println("Data Node 1 finalizado correctamente.")
	}

	// Enviar solicitud de finalización a Data Node 2
	_, err = s.dataNode2Client.FinishDNodes(ctxDataNode, &pbDataNode.FinishDNodesRequest{Req: 1})
	if err != nil {
		log.Printf("Error al finalizar Data Node 2: %v", err)
	} else {
		log.Println("Data Node 2 finalizado correctamente.")
	}
	return &pbDataNode.FinishDNodesResponse{Resp: 1}, nil
}

func (s *server) FinishRegionales(ctx context.Context, req *pb.FinishRegionalesRequest) (*pb.FinishRegionalesResponse, error) {
	// Crear un contexto con timeout para cada llamada gRPC
	ctxDataNode, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Enviar solicitud de finalización a regionales

	_, err := s.continenteFolderClient.FinishRegionales(ctxDataNode, &pb.FinishRegionalesRequest{Req: 1})
	if err != nil {
		log.Printf("Error al finalizar Continente Folder: %v", err)
	} else {
		log.Println("Continente Folder finalizado correctamente.")
	}

	_, err = s.continenteServerClient.FinishRegionales(ctxDataNode, &pb.FinishRegionalesRequest{Req: 1})
	if err != nil {
		log.Printf("Error al finalizar Continente Server: %v", err)
	} else {
		log.Println("Continente Server finalizado correctamente.")
	}

	_, err = s.islaFileClient.FinishRegionales(ctxDataNode, &pb.FinishRegionalesRequest{Req: 1})
	if err != nil {
		log.Printf("Error al finalizar Isla File: %v", err)
	} else {
		log.Println("Isla File finalizado correctamente.")
	}

	return &pb.FinishRegionalesResponse{Resp: 1}, nil
}

func createFile() {
	file, err := os.Create("INFO.txt")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()
}

func setupConnections(addr string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error

	// Retry loop for each connection
	for {
		conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			log.Printf("Connected successfully to %s", addr)
			break
		}

		// Log error and retry after 5 seconds
		log.Printf("Failed to connect to %s: %v. Retrying in 5 seconds...", addr, err)
		time.Sleep(5 * time.Second)
	}

	return conn, err
}

func main() {
	createFile()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	
	// Conectar al Data Node 1
	conn1, err := grpc.Dial("dist101:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar al Data Node 1: %v", err)
	}
	defer conn1.Close()

	// Conectar al Data Node 2
	conn2, err := grpc.Dial("dist102:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar al Data Node 2: %v", err)
	}
	defer conn2.Close()

	// Conectar al Data Node 2
	conn3, err := grpc.Dial("dist102:50056", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar a Isla File: %v", err)
	}
	defer conn3.Close()

	// Conectar al Data Node 2
	conn4, err := grpc.Dial("dist104:50057", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar al Continente Server: %v", err)
	}
	defer conn4.Close()

	// Conectar al Data Node 2
	conn5, err := grpc.Dial("dist103:50058", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("No se pudo conectar al Continente Folder: %v", err)
	}
	defer conn5.Close()
	// Crear clientes para cada Data Node
	dataNode1Client := pbDataNode.NewDNodeClient(conn1)
	dataNode2Client := pbDataNode.NewDNodeClient(conn2)
	islaFileClient := pb.NewPrimaryNodeClient(conn3)
	continenteServerClient := pb.NewPrimaryNodeClient(conn4)
	continenteFolderClient := pb.NewPrimaryNodeClient(conn5)

	stopChan := make(chan struct{})

	s := grpc.NewServer()
	primaryServer := &server{
		dataNode1Client:        dataNode1Client,
		dataNode2Client:        dataNode2Client,
		islaFileClient:         islaFileClient,
		continenteFolderClient: continenteFolderClient,
		continenteServerClient: continenteServerClient,
		stopChan:               stopChan,
	}

	pb.RegisterPrimaryNodeServer(s, primaryServer)
	pbTai.RegisterTaiServer(s, primaryServer)

	// Goroutine para detener el servidor cuando se reciba la señal de finalización
	go func() {
		<-stopChan
		log.Println("Deteniendo el servidor gRPC...")
		s.GracefulStop()
	}()

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
