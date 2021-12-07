package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	pb "github.com/aprilsanchez/SleepTracker/sleep_tracker/proto"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedSleepTrackerServiceServer
}

var db map[string]string

var (
	port = flag.Int("port", 50051, "The server port")
)

func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.ServerResponse, error) {
	// Key is in the format user:start_date:end_Date
	var err error
	var value string

	key := strings.Split(in.GetKey(), ":")
	user := key[0]
	if key[1] == key[2] {
		key := user + ":" + key[1]
		value = db[key]
		log.Println("Getting " + key + ": " + value)
		return generateResponse(value, err)
	}

	var f float64
	var init_day int
	var end_day int

	// Dates are in the format MM/DD/YYYY
	start_date := strings.Split(key[1], "/")
	end_date := strings.Split(key[2], "/")
	total_hours := 0.0

	log.Println("Getting " + user + ":" + key[1] + ":" + key[2])
	// We assume the dates are within the same month
	init_day, err = strconv.Atoi(start_date[1])
	if err != nil {
		return generateResponse(value, err)
	}
	end_day, err = strconv.Atoi(end_date[1])
	for i := init_day; i <= end_day; i++ {
		day := strconv.Itoa(i)
		if i < 10 {
			day = "0" + day
		}
		key := user + ":" + start_date[0] + "/" + day + "/" + start_date[2]
		value = db[key]
		log.Println("db[" + key + "] = " + value)
		f, err = strconv.ParseFloat(value, 64)
		total_hours += f
	}
	log.Println("Total hours: ", total_hours)

	log.Println("init day: ", init_day)
	log.Println("end day: ", end_day)
	log.Println("end - init + 1: ", end_day-init_day+1)
	num_days := end_day - init_day + 1
	value = fmt.Sprint(total_hours / (float64(num_days)))
	log.Println("Average: ", value)
	return generateResponse(value, err)
}

func (s *server) Set(ctx context.Context, in *pb.SetRequest) (*pb.ServerResponse, error) {
	log.Println("Setting " + in.GetKey() + " = " + in.GetValue())
	db[in.GetKey()] = in.GetValue()
	return generateResponse("Set Key", nil)
}

func (s *server) Delete(ctx context.Context, in *pb.DeleteRequest) (*pb.ServerResponse, error) {
	log.Println("Deleting key " + in.GetKey())
	delete(db, in.GetKey())
	return generateResponse("Deleted", nil)
}

func generateResponse(value string, err error) (*pb.ServerResponse, error) {
	if err != nil {
		return &pb.ServerResponse{Success: false, Value: value, Error: err.Error()}, nil
	}
	return &pb.ServerResponse{Success: true, Value: value, Error: ""}, nil
}

func main() {

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	db = make(map[string]string)

	pb.RegisterSleepTrackerServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
