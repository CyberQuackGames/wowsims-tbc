// +build wasm

package main

import (
	"log"
	"syscall/js"
	"time"

	"github.com/wowsims/tbc/sim"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func init() {
	sim.RegisterAll()
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("computeStats", js.FuncOf(computeStats))
	js.Global().Set("gearList", js.FuncOf(gearList))
	js.Global().Set("individualSim", js.FuncOf(individualSim))
	js.Global().Set("raidSim", js.FuncOf(raidSim))
	js.Global().Set("raidSimAsync", js.FuncOf(raidSimAsync))
	js.Global().Set("statWeights", js.FuncOf(statWeights))
	js.Global().Call("wasmready")
	<-c
}

func computeStats(this js.Value, args []js.Value) interface{} {
	// Assumes args[0] is a Uint8Array
	data := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(data, args[0])

	csr := &proto.ComputeStatsRequest{}
	if err := googleProto.Unmarshal(data, csr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.ComputeStats(csr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func gearList(this js.Value, args []js.Value) interface{} {
	// Assumes args[0] is a Uint8Array
	data := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(data, args[0])

	glr := &proto.GearListRequest{}
	if err := googleProto.Unmarshal(data, glr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.GetGearList(glr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func individualSim(this js.Value, args []js.Value) interface{} {
	// Assumes args[0] is a Uint8Array
	data := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(data, args[0])

	isr := &proto.IndividualSimRequest{}
	if err := googleProto.Unmarshal(data, isr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.RunIndividualSim(isr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func raidSim(this js.Value, args []js.Value) interface{} {
	// Assumes args[0] is a Uint8Array
	data := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(data, args[0])

	rsr := &proto.RaidSimRequest{}
	if err := googleProto.Unmarshal(data, rsr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.RunRaidSim(rsr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}

func raidSimAsync(this js.Value, args []js.Value) interface{} {
	log.Printf("Started async sim")
	// Assumes args[0] is a Uint8Array
	data := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(data, args[0])

	rsr := &proto.RaidSimRequest{}
	if err := googleProto.Unmarshal(data, rsr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	reporter := make(chan *proto.ProgressMetrics, 100)
	core.RunRaidSimAsync(rsr, reporter)

reader:
	for {
		// TODO: cleanup so we dont collect these
		select {
		case progMetric, ok := <-reporter:
			if !ok {
				break reader
			}

			outbytes, err := googleProto.Marshal(progMetric)
			if err != nil {
				log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
				return nil
			}

			outArray := js.Global().Get("Uint8Array").New(len(outbytes))
			js.CopyBytesToJS(outArray, outbytes)

			args[1].Invoke(outArray)

			if progMetric.FinalResult != nil {
				close(reporter)
				return outArray
			}
			time.Sleep(time.Millisecond * 100)
		}
	}

	return nil
}

func statWeights(this js.Value, args []js.Value) interface{} {
	// Assumes args[0] is a Uint8Array
	data := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(data, args[0])

	swr := &proto.StatWeightsRequest{}
	if err := googleProto.Unmarshal(data, swr); err != nil {
		log.Printf("Failed to parse request: %s", err)
		return nil
	}
	result := core.StatWeights(swr)

	outbytes, err := googleProto.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal result: %s", err.Error())
		return nil
	}

	outArray := js.Global().Get("Uint8Array").New(len(outbytes))
	js.CopyBytesToJS(outArray, outbytes)

	return outArray
}
