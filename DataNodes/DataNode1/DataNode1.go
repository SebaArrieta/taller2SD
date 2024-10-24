package main

import (
    "context"
    "log"
    "net"
    "os"

    "google.golang.org/grpc"
    pbDataNode "DataNode1/generated/DataNode" // Ruta generada para el servicio storeAtributo
)

// Servidor que implementa el servicio storeAtributo
type server struct {
    pbDataNode.UnimplementedStoreAtributoServer
}

// Implementación del método getAtributo
func (s *server) GetAtributo(ctx context.Context, req *pbDataNode.Request) (*pbDataNode.Response, error) {
    log.Printf("DataNode1 recibió: %s", req.DataAtributo)

    // Guardar el atributo en el archivo INFO_1.txt
    file, err := os.OpenFile("INFO_1.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Printf("Error al abrir el archivo: %v", err)
        return nil, err
    }
    defer file.Close()

    if _, err := file.WriteString(req.DataAtributo + "\n"); err != nil {
        log.Printf("Error al escribir en el archivo: %v", err)
        return nil, err
    }

    log.Println("Datos almacenados en INFO_1.txt")

    return &pbDataNode.Response{Message: "Data almacenada en DataNode1"}, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50052") // Puerto para DataNode1
    if err != nil {
        log.Fatalf("Error al iniciar DataNode1: %v", err)
    }

    s := grpc.NewServer()
    pbDataNode.RegisterStoreAtributoServer(s, &server{})

    log.Println("DataNode1 corriendo en el puerto :50052")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Error al servir DataNode1: %v", err)
    }
}
