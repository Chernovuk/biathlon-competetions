package main

import (
	"fmt"
	"os"

	"github.com/Chernovuk/biathlon-competetions/internal/biathlon"
	"github.com/Chernovuk/biathlon-competetions/internal/statistics"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %v EVENTS_FILEPATH CONFIG_FILEPATH", os.Args[0])
		os.Exit(1)
	}

	eventsFP := os.Args[1]
	listener, err := biathlon.NewEventListener(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to open specified file: %v", eventsFP)
		os.Exit(1)
	}

	configFP := os.Args[2]
	config, err := biathlon.ParseConfig(configFP)
	if err != nil {
		fmt.Printf("Failed to open specified file: %v", configFP)
		os.Exit(1)
	}

	stats := statistics.New(config)
	processor := biathlon.NewProcessor(config, listener.Events())

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

	go listener.Start()
	processor.Start()

	fmt.Println()
	table := stats.GetResults()
	for _, v := range table {
		fmt.Println(v.String())
	}
}
