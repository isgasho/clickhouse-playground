package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type ExecServer struct {
	dbServer *os.Process
	dbClient *os.Process
	mu       sync.Mutex
}

func NewExecServer() (*ExecServer, error) {
	var server ExecServer

	err := server.startDbServer()
	if err != nil {
		return nil, err
	}
	//all, err := process.Processes()
	//if err != nil {
	//	return nil, err
	//}
	//
	//fmt.Println(all)

	//p, err := process.NewProcess(int32(cmd.Process.Pid))
	//if err != nil {
	//	return nil, err
	//}
	//
	//children, err := p.Children()
	//if err != nil {
	//	return nil, err
	//}
	//
	//fmt.Println(children)
	return &server, nil
}

func (s *ExecServer) exec(query string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("in exec")

	cmd := exec.Command("/usr/bin/clickhouse", "client", "-nm", "-f", "JSON")

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	_, err = io.WriteString(stdin, query)
	if err != nil {
		return "", err
	}

	stdin.Close()
	cmd.Wait()

	return outb.String(), nil
}

func (s *ExecServer) startDbServer() error {
	cmd := exec.Command("/entrypoint.sh")
	err := cmd.Start()
	if err != nil {
		return err
	}
	s.dbServer = cmd.Process

	return nil
}

func (s *ExecServer) stopDbServer() error {
	err := s.dbServer.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	// TODO: maybe handle state?
	_, err = s.dbServer.Wait()
	if err != nil {
		return err
	}

	cmd := exec.Command("rm", "-rf", "/data", "/metadata")
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *ExecServer) restartDbServer() error {
	if err := s.stopDbServer(); err != nil {
		return err
	}

	if err := s.startDbServer(); err != nil {
		return err
	}

	return nil
}

type execInput struct {
	QueryStr string `json:"query"`
}

func (s *ExecServer) handleExec(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var input execInput
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.exec(input.QueryStr)
	if err != nil {
		//log.Fatal(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list := make([]map[string]interface{}, 0)

	data := make(map[string]interface{})
	d := json.NewDecoder(strings.NewReader(result))

	for {
		err = d.Decode(&data)
		if err != nil {
			fmt.Println(err)
			break
			//http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		list = append(list, data)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	if err = encoder.Encode(list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

//func (s *ExecServer) handleRestart(w http.ResponseWriter, r *http.Request) {
//	if err := s.restartDbServer(); err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//	}
//}

func main() {
	server, err := NewExecServer()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/exec", server.handleExec)
	//mux.HandleFunc("/restart", server.handleRestart)

	_ = http.ListenAndServe(":8080", mux)
}
