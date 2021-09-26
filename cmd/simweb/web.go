package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/wowsims/tbc/api"
	// "github.com/wowsims/tbc/dist"
	"google.golang.org/protobuf/proto"
)

func main() {
	var useFS = flag.Bool("usefs", true, "Use local file system and wasm. Set to true for dev")
	// TODO: usefs for now is set to true until we can solve how to embed the dist.
	var host = flag.String("host", ":3333", "URL to host the interface on.")

	flag.Parse()

	var fs http.Handler
	if *useFS {
		log.Printf("Using local file system for development.")
		fs = http.FileServer(http.Dir("./dist"))
	} else {
		log.Printf("Embedded file server running.")
		// fs = http.FileServer(http.FS(dist.FS))
	}

	http.HandleFunc("/statWeights", handleAPI)
	http.HandleFunc("/computeStats", handleAPI)
	http.HandleFunc("/individualSim", handleAPI)
	http.HandleFunc("/gearList", handleAPI)

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(req.URL.Path, ".wasm") {
			resp.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(resp, req)
	})

	url := fmt.Sprintf("http://localhost%s/elemental_shaman/", *host)
	log.Printf("Launching interface on %s", url)

	go func() {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("explorer", url)
		} else if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", url)
		} else if runtime.GOOS == "linux" {
			cmd = exec.Command("xdg-open", url)
		}
		err := cmd.Start()
		if err != nil {
			log.Printf("Error launching browser: %#v", err.Error())
		}
		log.Printf("Closing: %s", http.ListenAndServe(*host, nil))
	}()

	fmt.Printf("Enter Command... '?' for list\n")
	for {
		fmt.Printf("> ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if len(text) == 0 {
			continue
		}
		command := strings.TrimSpace(text)
		switch command {
		case "profile":
			filename := fmt.Sprintf("profile_%d.cpu", time.Now().Unix())
			fmt.Printf("Running profiling for 15 seconds, output to %s\n", filename)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			go func() {
				time.Sleep(time.Second * 15)
				pprof.StopCPUProfile()
				f.Close()
				fmt.Printf("Profiling complete.\n> ")
			}()
		case "quit":
			os.Exit(1)
		case "?":
			fmt.Printf("Commands:\n\tprofile - start a CPU profile for debugging performance\n\tquit - exits\n\n")
		case "":
			// nothing.
		default:
			fmt.Printf("Unknown command: '%s'", command)
		}
	}
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	endpoint := r.URL.Path

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {

		return
	}
	var msg proto.Message
	switch endpoint {
	case "/individualSim":
		msg = &api.IndividualSimRequest{}
	case "/statWeights":
		msg = &api.StatWeightsRequest{}
	case "/computeStats":
		msg = &api.ComputeStatsRequest{}
	case "/gearList":
		msg = &api.GearListRequest{}
	default:
		log.Printf("Invalid Endpoint: %s", endpoint)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := proto.Unmarshal(body, msg); err != nil {
		log.Printf("Failed to parse request: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var result proto.Message
	switch endpoint {
	case "/individualSim":
		result = api.RunSimulation(msg.(*api.IndividualSimRequest))
	case "/statWeights":
		result = api.StatWeights(msg.(*api.StatWeightsRequest))
	case "/computeStats":
		result = api.ComputeStats(msg.(*api.ComputeStatsRequest))
	case "/gearList":
		result = api.GetGearList(msg.(*api.GearListRequest))
	}

	outbytes, err := proto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/x-protobuf")
	w.Write(outbytes)
}
