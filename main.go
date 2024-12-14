package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	dem "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)

type BombDefuseInfo struct {
	Round        int
	EnemiesAlive int
	MapName      string
	Date         string
}

const maxParallelism = 25
const defuseOutputFile = "defuse_results.txt"

func processDemoFile(filePath string, playerID uint64, fileDate time.Time, wg *sync.WaitGroup, semaphore chan struct{}, processed *int32, total int, outputFile *os.File) {
	defer wg.Done()
	<-semaphore
	defer func() { semaphore <- struct{}{} }()

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("failed to open demo file %s: %v", filePath, err)
		return
	}
	defer f.Close()

	p := dem.NewParser(f)
	defer p.Close()

	var bombDefuses []BombDefuseInfo
	finalTScore := 0
	finalCTScore := 0

	p.RegisterEventHandler(func(e events.BombDefused) {
		if e.Player.SteamID64 == playerID {
			enemiesAlive := 0
			players := p.GameState().TeamTerrorists().Members()
			for _, player := range players {
				if player.IsAlive() {
					enemiesAlive++
				}
			}
			if enemiesAlive > 0 {
				bombDefuses = append(bombDefuses, BombDefuseInfo{
					Round:        p.GameState().TotalRoundsPlayed(),
					EnemiesAlive: enemiesAlive,
					MapName:      p.Header().MapName,
					Date:         fileDate.Format("2006-01-02"),
				})
			}
		}
	})

	p.RegisterEventHandler(func(e events.AnnouncementWinPanelMatch) {
		finalTScore = p.GameState().TeamTerrorists().Score()
		finalCTScore = p.GameState().TeamCounterTerrorists().Score()
	})

	err = p.ParseToEnd()
	if err != nil {
		log.Printf("failed to parse demo file %s: %v", filePath, err)
		return
	}

	for _, defuse := range bombDefuses {
		output := fmt.Sprintf("File: %s, Date: %s, Map: %s, Final Score: CT %d - T %d, Round: %d, Enemies Alive: %d\n",
			filepath.Base(filePath), defuse.Date, defuse.MapName, finalCTScore, finalTScore, defuse.Round, defuse.EnemiesAlive)
		fmt.Print(output)
		if _, err := outputFile.WriteString(output); err != nil {
			log.Printf("failed to write defuse result for file %s: %v", filePath, err)
		}
	}

	processedCount := atomic.AddInt32(processed, 1)
	fmt.Printf("Progress: %d/%d files processed\n", processedCount, total)
}

func main() {
	demoDirectory := "D:\\SteamLibrary\\steamapps\\common\\Counter-Strike Global Offensive\\game\\csgo\\replays"
	playerID := uint64(76561199121731119)
	filterDate := "2024-11-17"

	parsedFilterDate, err := time.Parse("2006-01-02", filterDate)
	if err != nil {
		log.Fatalf("invalid filter date format: %v", err)
	}

	var demoFiles []string
	var fileDates []time.Time
	err = filepath.Walk(demoDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("error accessing path %s: %v", path, err)
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ".dem" {
			fileModTime := info.ModTime()
			if fileModTime.After(parsedFilterDate) || fileModTime.Equal(parsedFilterDate) {
				demoFiles = append(demoFiles, path)
				fileDates = append(fileDates, fileModTime)
			}
		}
		return nil
	})

	if err != nil {
		log.Panicf("error walking through demo directory: %v", err)
	}

	sort.SliceStable(demoFiles, func(i, j int) bool {
		return fileDates[i].Before(fileDates[j])
	})
	sort.SliceStable(fileDates, func(i, j int) bool {
		return fileDates[i].Before(fileDates[j])
	})

	totalFiles := len(demoFiles)
	if totalFiles == 0 {
		fmt.Println("No demo files found.")
		return
	}

	fmt.Printf("Found %d demo files after %s. Starting processing...\n", totalFiles, parsedFilterDate.Format("2006-01-02"))

	outputFile, err := os.Create(defuseOutputFile)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxParallelism)
	for i := 0; i < maxParallelism; i++ {
		semaphore <- struct{}{}
	}

	var processed int32

	for i, filePath := range demoFiles {
		wg.Add(1)
		go processDemoFile(filePath, playerID, fileDates[i], &wg, semaphore, &processed, totalFiles, outputFile)
	}

	wg.Wait()
	fmt.Println("Processing complete.")
}
