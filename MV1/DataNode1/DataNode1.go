package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"strings"
	"fmt"

	pbDataNode "DataNode1/generated/DataNode" // Ruta generada para el servicio storeAtributo

	"google.golang.org/grpc"
)

// Servidor que implementa el servicio storeAtributo
type server struct {
	pbDataNode.UnimplementedDNodeServer
	stopChan chan struct{}
}

// Implementación del método getAtributo
func (s *server) GetAtributo(ctx context.Context, req *pbDataNode.Request) (*pbDataNode.Response, error) {

	// Guardar el atributo en el archivo INFO_1.txt
	file, err := os.OpenFile("INFO_1.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error al abrir el archivo: %v", err)
		return nil, err
	}
	defer file.Close()

	if _, err := file.WriteString(req.Message + "\n"); err != nil {
		log.Printf("Error al escribir en el archivo: %v", err)
		return nil, err
	}

	log.Printf("[<DATA NODE 2>] Solicitud de Primary Node recibida, mensaje enviado: Data %s almacenada en DataNode2", req.Message)
	return &pbDataNode.Response{Message: fmt.Sprintf("Data %s almacenada en DataNode1", req.Message)}, nil
}

func (s *server) SendData(ctx context.Context, req *pbDataNode.Request) (*pbDataNode.Response, error) {
	// Separar las IDs por ';'
	ids := strings.Split(req.Message, ";")

	file, err := os.Open("INFO_1.txt")
	if err != nil {
		log.Printf("Error al abrir el archivo: %v", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var foundAttributes []string

	// Buscar cada ID en el archivo
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",") // Asume que la ID está en la primera columna y el atributo en la segunda

		for _, id := range ids {
			if strings.TrimSpace(parts[0]) == strings.TrimSpace(id) { // Coincidir con la ID
				foundAttributes = append(foundAttributes, parts[1]) // Agregar el atributo correspondiente
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error leyendo el archivo: %v", err)
		return nil, err
	}

	if len(foundAttributes) > 0 {
		result := strings.Join(foundAttributes, ";")
		log.Printf("Atributos encontrados: %s", result)
		log.Printf("[<DATA NODE 1>] Solicitud de Primary Node recibida, mensaje enviado: %s", result)
		return &pbDataNode.Response{Message: result}, nil
	}

	log.Printf("[<DATA NODE 1>] Solicitud de Primary Node recibida, mensaje enviado: %s", "-1")
	log.Println("No se encontraron atributos para las IDs proporcionadas")
	return &pbDataNode.Response{Message: "-1"}, nil
}

func (s *server) FinishDNodes(ctx context.Context, req *pbDataNode.FinishDNodesRequest) (*pbDataNode.FinishDNodesResponse, error) {
	log.Println("Finalizando DataNode1...")
	// Enviar señal para cerrar el servidor
	s.stopChan <- struct{}{}
	return &pbDataNode.FinishDNodesResponse{Resp: 1}, nil
}

func createFile() {
	file, err := os.Create("INFO_1.txt")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()
}

func main() {
	createFile()
	lis, err := net.Listen("tcp", ":50052") // Puerto para DataNode1
	if err != nil {
		log.Fatalf("Error al iniciar DataNode1: %v", err)
	}
	stopChan := make(chan struct{})

	s := grpc.NewServer()
	pbDataNode.RegisterDNodeServer(s, &server{stopChan: stopChan})

	go func() {
		<-stopChan
		log.Println("Deteniendo el servidor gRPC...")
		s.GracefulStop()
	}()

	log.Println("DataNode1 corriendo en el puerto :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir DataNode1: %v", err)
	}
}
