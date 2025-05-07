package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Chernovuk/biathlon-competetions/internal/biathlon"
	"github.com/Chernovuk/biathlon-competetions/internal/statistics"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %v EVENTS_FILEPATH CONFIG_FILEPATH\n", os.Args[0])
		os.Exit(1)
	}

	eventsFP := os.Args[1]
	listener, err := biathlon.NewEventListener(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to open specified file: %v\n", eventsFP)
		os.Exit(1)
	}

	listenerLogFile, err := os.OpenFile("listener.log", os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		fmt.Printf("Failed to open listne log file: %v\n", listenerLogFile)
		os.Exit(1)
	}
	defer listenerLogFile.Close()

	listener.SetLogger(log.New(listenerLogFile, "Listener: ", log.Ltime))

	configFP := os.Args[2]
	config, err := biathlon.ParseConfig(configFP)
	if err != nil {
		fmt.Printf("Failed to open specified file: %v\n", configFP)
		os.Exit(1)
	}

	stats := statistics.New(config)
	processor := biathlon.NewProcessor(config, listener.Events())
	handleStats(processor, stats)

	processorLogFile, err := os.OpenFile("processor.log", os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		fmt.Printf("Failed to open listne log file: %v\n", processorLogFile)
		os.Exit(1)
	}
	defer processorLogFile.Close()

	processor.SetLogger(biathlon.NewDefaultLogger(processorLogFile))

	go listener.Start()
	processor.Start()

	table := stats.GetResults()
	showReport(table)
}

func handleStats(processor *biathlon.Processor, stats *statistics.Statistics) {
	processor.Handle(biathlon.Register, stats.OnRegister)
	processor.Handle(biathlon.BeSheduled, stats.OnBeSheduled)
	processor.Handle(biathlon.Start, stats.OnStart)
	processor.Handle(biathlon.HitTarget, stats.OnHitTarget)
	processor.Handle(biathlon.LeaveFiringRange, stats.OnComeToFiringRange)
	processor.Handle(biathlon.EnterPenaltyLap, stats.OnEnterPenaltyLap)
	processor.Handle(biathlon.LeavePenaltyLap, stats.OnLeavePenaltyLap)
	processor.Handle(biathlon.EndMainLap, stats.OnEndMainLap)
	processor.Handle(biathlon.BeUnableToContinue, stats.OnBeUnableToContinue)
	processor.Handle(biathlon.Disqualify, stats.OnDisqualify)
	processor.Handle(biathlon.Finish, stats.OnFinish)
}

func showReport(table []statistics.Result) {
	for _, v := range table {
		fmt.Println(v.String())
	}
}
