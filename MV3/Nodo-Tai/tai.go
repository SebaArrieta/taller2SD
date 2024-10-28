package main

import (
	pbDiaboromon "Nodo-Tai/generated/Diaboromon"

	pbTai "Nodo-Tai/generated/Tai"
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var PS, TE, TD, CD, VI float64
var flagEnd = false

type server struct {
	pbDiaboromon.UnimplementedDiaboromonServer
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

// Solicitar los datos de los Digimons sacrificados al Primary Node
func requestDigimonData(client pbTai.TaiClient) (float32, error) {
	req := &pbTai.Request{RequestMessage: "Solicitar datos de Digimons sacrificados"}
	resp, err := client.GetSacrificed(context.Background(), req)
	if err != nil {
		return 0, fmt.Errorf("error solicitando datos: %v", err)
	}
	fmt.Printf("Digimons sacrificados: %d, Datos acumulados: %.2f\n", resp.SacrificedDigimons, resp.AccumulatedData)
	return resp.AccumulatedData, nil
}

func connectToPrimaryNode(addr string) (*grpc.ClientConn, pbTai.TaiClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("no se pudo conectar: %v", err)
	}
	client := pbTai.NewTaiClient(conn)
	return conn, client, nil
}

func (s *server) Attack(ctx context.Context, req *pbDiaboromon.AttackRequest) (*pbDiaboromon.AttackResponse, error) {

	VI -= float64(req.Attackreq)
	fmt.Printf("[Nodo Tai] Diaboromon ha atacado. Greymon y Garurumon reciben 10 puntos de daño. Vida restante: %.2f\n", VI)

	if VI <= 0 {
		fmt.Println("[Nodo Tai] Tai ha sido derrotado.")
		fmt.Println("[Nodo Tai] Pulsa cualquier tecla para salir.")
		flagEnd = true
		return &pbDiaboromon.AttackResponse{Attackresp: -1}, nil
	}
	return &pbDiaboromon.AttackResponse{Attackresp: 1}, nil
}

// Función para atacar a Diaboromon
func attackDiaboromon(client pbDiaboromon.DiaboromonClient, accumulatedData float32, clientPrimary pbTai.TaiClient) {
	req := &pbDiaboromon.AttackRequest{Attackreq: accumulatedData}
	resp, err := client.Attack(context.Background(), req)
	if err != nil {
		log.Fatalf("[Nodo Tai] Error atacando a Diaboromon: %v", err)
	}

	if resp.Attackresp == -1 {
		fmt.Print("[Nodo Tai] Greymon y Garurumon han digievolucionado a Omegamon.\n")
		fmt.Printf("[Nodo Tai] Diaboromon ha sido derrotado por Tai.\n")
		_, err := clientPrimary.FinishTai(context.Background(), &pbTai.FinishTaiRequest{})
		if err != nil {
			log.Fatalf("Error notificando al Primary Node que Tai ha terminado: %v", err)
		}
		os.Exit(0)
	}

	fmt.Printf("[Nodo Tai] La cantidad de datos es insuficiente para digievolucionar a Omegamon.\n")

	VI -= 10
	fmt.Printf("[Nodo Tai] Greymon y Garurumon reciben 10 puntos de daño. Vida restante: %.2f\n", VI)

	if VI <= 0 {
		fmt.Println("[Nodo Tai] Tai ha sido derrotado.")
		notifyDiaboromonDefeat(client)
		_, err := clientPrimary.FinishTai(context.Background(), &pbTai.FinishTaiRequest{})
		if err != nil {
			log.Fatalf("Error notificando al Primary Node que Tai ha terminado: %v", err)
		}
		os.Exit(0)
	}

}

// Función para notificar la derrota de Tai a Diaboromon
func notifyDiaboromonDefeat(client pbDiaboromon.DiaboromonClient) {
	req := &pbDiaboromon.DefeatRequest{Defreq: 1}
	_, err := client.TaiDefeated(context.Background(), req)
	if err != nil {
		log.Fatalf("[Nodo Tai] Error notificando la derrota de Tai: %v", err)
	}

}

func connectToDiaboromon(addr string) (*grpc.ClientConn, pbDiaboromon.DiaboromonClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	fmt.Printf("[Nodo Tai] no se pudo conectar a Diaboromon %v", err)
	if err != nil {
		return nil, nil, fmt.Errorf("[Nodo Tai] no se pudo conectar a Diaboromon: %v", err)
	}
	client := pbDiaboromon.NewDiaboromonClient(conn)
	return conn, client, nil
}

func main() {
	readVariables("INPUT.txt")
	var accumulatedData float32 = 0.0
	addrPrimaryNode := "dist101:50051"
	addrDiaboromon := "dist104:50055"
	var connDiaboromon *grpc.ClientConn
	var clientDiaboromon pbDiaboromon.DiaboromonClient

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("[Nodo Tai] Error al iniciar el servidor: %v", err)
	}

	grpcServer := grpc.NewServer()
	pbDiaboromon.RegisterDiaboromonServer(grpcServer, &server{})

	fmt.Println("[Nodo Tai] Servidor Tai corriendo en el puerto 50054...")
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("[Nodo Tai] Error al iniciar el servidor gRPC: %v", err)
		}
	}()

	// Conectar al Primary Node
	connPrimary, clientPrimary, err := connectToPrimaryNode(addrPrimaryNode)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer connPrimary.Close()

	connDiaboromon, clientDiaboromon, err = connectToDiaboromon(addrDiaboromon)
	if err != nil {
		log.Fatalf("%v", err)
	}

	defer connDiaboromon.Close()

	for {
		fmt.Println("Seleccione una opción:")
		fmt.Println("1. Pedir datos sacrificados")
		fmt.Println("2. Atacar a Diaboromon")
		fmt.Println("0. Salir")

		var option int

		fmt.Scan(&option)
		if flagEnd {
			_, err := clientPrimary.FinishTai(context.Background(), &pbTai.FinishTaiRequest{})
			if err != nil {
				log.Fatalf("Error notificando al Primary Node que Tai ha terminado: %v", err)
			}
			break
		}

		switch option {
		case 1:
			// Solicitar datos al Primary Node
			accumulatedData, err = requestDigimonData(clientPrimary)
			if err != nil {
				log.Fatalf("%v", err)
			}
		case 2:
			attackDiaboromon(clientDiaboromon, accumulatedData, clientPrimary)

		case 0:
			fmt.Println("Saliendo del programa...")

			// Notificar al Primary Node que Tai ha terminado
			_, err := clientPrimary.FinishTai(context.Background(), &pbTai.FinishTaiRequest{})
			if err != nil {
				log.Fatalf("Error notificando al Primary Node que Tai ha terminado: %v", err)
			}
			notifyDiaboromonDefeat(clientDiaboromon)

			return

		default:
			fmt.Println("Opción no válida. Intente de nuevo.")
		}
	}

}
