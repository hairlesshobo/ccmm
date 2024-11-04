// =================================================================================
//
//		ccmm - https://www.foxhollow.cc/projects/ccmm/
//
//	 go-import-media, aka gim, is a tool for automatically importing media
//	 from removable disks into a predefined folder structure automatically.
//
//		Copyright (c) 2024 Steve Cross <flip@foxhollow.cc>
//
//		Licensed under the Apache License, Version 2.0 (the "License");
//		you may not use this file except in compliance with the License.
//		You may obtain a copy of the License at
//
//		     http://www.apache.org/licenses/LICENSE-2.0
//
//		Unless required by applicable law or agreed to in writing, software
//		distributed under the License is distributed on an "AS IS" BASIS,
//		WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//		See the License for the specific language governing permissions and
//		limitations under the License.
//
// =================================================================================
package action

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"ccmm/importer/processor"
	"ccmm/model"
	"ccmm/util"
)

type ImportStatus int8

const (
	Pending ImportStatus = iota
	Scanning
	Importing
	Completed
	Failed
)

type ImportQueueItem struct {
	Params           model.ImportVolume
	Processors       []processor.Processor
	Files            []model.SourceFile
	Status           ImportStatus
	FinishedCallback func(queueItem *ImportQueueItem)
	processCallback  func(queueItem *ImportQueueItem)
}

var (
	queueIndex  int
	importQueue = map[int]*ImportQueueItem{}
	importMutex sync.Mutex
)

func getFirstQueueIndex(importQueue *map[int]*ImportQueueItem) int {
	keys := make([]int, 0)
	for k := range *importQueue {
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return -1
	}

	sort.Ints(keys)
	for _, k := range keys {
		return k
	}

	return -1
}

func ImportWorker() {
	// TODO: add cancel channel
	for {
		importMutex.Lock()
		firstQueueIndex := getFirstQueueIndex(&importQueue)

		// nothing pending, so we wait a bit and check again.
		if firstQueueIndex < 0 {
			importMutex.Unlock()
			time.Sleep(500 * time.Millisecond)
			continue
		}
		queueItem := importQueue[firstQueueIndex]
		importMutex.Unlock()

		slog.Info(fmt.Sprintf("Processing import queue item '%d', volume path: '%s'", firstQueueIndex, queueItem.Params.VolumePath))

		queueItem.processCallback(queueItem)

		importMutex.Lock()
		slog.Info(fmt.Sprintf("Finished processing import queue item '%d', volume path: '%s'", firstQueueIndex, queueItem.Params.VolumePath))
		delete(importQueue, firstQueueIndex)
		importMutex.Unlock()
	}
}

func Import(params model.ImportVolume, finishedCallback func(queueItem *ImportQueueItem)) {
	if !util.DirectoryExists(params.VolumePath) {
		slog.Error(fmt.Sprintf("Cannot import because directory not found: %s", params.VolumePath))
		return
	}

	queueIndex++
	slog.Info(fmt.Sprintf("Queueing import #%d for volume '%s'", queueIndex, params.VolumePath))

	importMutex.Lock()
	importQueue[queueIndex] = &ImportQueueItem{
		Params:           params,
		Status:           Scanning,
		Processors:       make([]processor.Processor, 0),
		Files:            make([]model.SourceFile, 0),
		FinishedCallback: finishedCallback,
	}
	importMutex.Unlock()

	processors := processor.FindProcessors(params.VolumePath)

	importMutex.Lock()
	importQueue[queueIndex].Status = Pending
	importQueue[queueIndex].Processors = processors
	importQueue[queueIndex].processCallback = func(queueItem *ImportQueueItem) {
		importMutex.Lock()
		queueItem.Status = Scanning
		importMutex.Unlock()

		files := processor.EnumerateSources(processors, params.Dump)

		importMutex.Lock()
		queueItem.Files = files
		queueItem.Status = Importing
		importMutex.Unlock()

		// TODO: add some sort of status callback here
		processor.ImportFiles(files, params.DryRun)

		slog.Info(fmt.Sprintf("Finished import for volume '%s'", params.VolumePath))
		queueItem.FinishedCallback(queueItem)
	}
	importMutex.Unlock()
}
