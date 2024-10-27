package main

import (
	pb "Nodo-Diaboromon/generated"
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var PS, TE, TD, CD, VI float64
var flagEnd = false

type server struct {
	pb.UnimplementedDiaboromonServer
}

func readVariables(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error al abrir archivo: %v", err)
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

func (s *server) Attack(ctx context.Context, req *pb.AttackRequest) (*pb.AttackResponse, error) {
	accumulatedData := req.Attackreq
	fmt.Printf("[Diaboromon] Diaboromon recibe un ataque... Tai posee %.2f datos acumulados\n", accumulatedData)

	// Verificar si los datos acumulados son suficientes para vencer a Diaboromon
	if float64(accumulatedData) >= CD {
		fmt.Print("[Diaboromon] Greymon y Garurumon han digievolucionado a Omegamon.\n")
		fmt.Print("[Diaboromon] Diaboromon ha sido derrotado por Omegamon de Tai.")
		flagEnd = true
		return &pb.AttackResponse{Attackresp: -1}, nil
	}

	fmt.Println("[Diaboromon] La cantidad de datos de Tai es insuficiente para invocar a Omegamon.")
	fmt.Println("[Diaboromon] Enviando 10 puntos de daño a Tai.")
	return &pb.AttackResponse{Attackresp: 10}, nil
}

// Función para manejar la derrota de Tai
func (s *server) TaiDefeated(ctx context.Context, req *pb.DefeatRequest) (*pb.DefeatResponse, error) {
	fmt.Println("[Diaboromon] Diaboromon recibe mensaje de que Tai ha sido derrotado.")
	flagEnd = true
	return &pb.DefeatResponse{Defresp: 1}, nil
}

// Enviar ataque a Tai
func sendAttack(client pb.DiaboromonClient) {
	fmt.Print("[Diaboromon] Atacando a nodo Tai...\n")
	req := &pb.AttackRequest{Attackreq: 10}
	resp, err := client.Attack(context.Background(), req)
	if err != nil {
		log.Fatalf("Error al atacar: %v", err)
	}
	if resp.Attackresp == -1 {
		fmt.Print("[Diaboromon] Diaboromon ha derrotado a Tai.")
		os.Exit(0)
	}
}

func connectToTai(addr string) (*grpc.ClientConn, pb.DiaboromonClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo conectar a Tai: %v", err)
	}
	client := pb.NewDiaboromonClient(conn)
	return conn, client, nil
}

func main() {
	readVariables("INPUT.txt")

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterDiaboromonServer(grpcServer, &server{})

	go func() {
		log.Println("[Diaboromon] Servidor Diaboromon corriendo en el puerto 50053...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
		}
	}()

	addr := "host.docker.internal:50054" // addr tai
	var conn *grpc.ClientConn
	var client pb.DiaboromonClient

	conn, client, err = connectToTai(addr)
	if err != nil {
		log.Fatalf("Error al conectar con Tai: %v", err)
	}
	defer conn.Close()

	for {
		time.Sleep(time.Duration(TD) * time.Second)
		if flagEnd {
			break
		}
		sendAttack(client)

	}
}
