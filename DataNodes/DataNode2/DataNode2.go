package main

import (
    "context"
    "log"
    "net"
    "os"

    "google.golang.org/grpc"
    pbDataNode "DataNode2/generated/DataNode" // Ruta generada para el servicio storeAtributo
)

// Servidor que implementa el servicio storeAtributo
type server struct {
    pbDataNode.UnimplementedStoreAtributoServer
}

// Implementación del método getAtributo
func (s *server) GetAtributo(ctx context.Context, req *pbDataNode.Request) (*pbDataNode.Response, error) {
    log.Printf("DataNode2 recibió: %s", req.DataAtributo)

    // Guardar el atributo en el archivo INFO_2.txt
    file, err := os.OpenFile("INFO_2.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Printf("Error al abrir el archivo: %v", err)
        return nil, err
    }
    defer file.Close()

    if _, err := file.WriteString(req.DataAtributo + "\n"); err != nil {
        log.Printf("Error al escribir en el archivo: %v", err)
        return nil, err
    }

    log.Println("Datos almacenados en INFO_2.txt")

    return &pbDataNode.Response{Message: "Data almacenada en DataNode2"}, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50053") // Puerto para DataNode2
    if err != nil {
        log.Fatalf("Error al iniciar DataNode2: %v", err)
    }

    s := grpc.NewServer()
    pbDataNode.RegisterStoreAtributoServer(s, &server{})

    log.Println("DataNode2 corriendo en el puerto :50053")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("Error al servir DataNode2: %v", err)
    }
}
