package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/IsaacDSC/auditory/internal/audit"
	"github.com/IsaacDSC/auditory/pkg/clock"
	"github.com/IsaacDSC/auditory/pkg/mu"
)

type FilePath string

func NewFilePath(key Key) FilePath {
	return FilePath(fmt.Sprintf("tmp/%s.json", key))
}

func (fp FilePath) String() string {
	return string(fp)
}

type DataFileStore struct {
	mu mu.MutexByKey
}

func NewDataFileStore() *DataFileStore {
	return &DataFileStore{
		mu: make(mu.MutexByKey),
	}
}

type Key string

type Date string

func NewDate(t time.Time) Date {
	return Date(fmt.Sprintf("%d-%d-%d", t.Year(), t.Month(), t.Day()))
}

type Data map[Date][]audit.DataAudit

func (dfs *DataFileStore) Upsert(ctx context.Context, input audit.DataAudit) error {
	key := Key(input.Metadata.Key)
	mu := dfs.mu.GetOrCreate(string(key))
	mu.Lock()
	defer mu.Unlock()

	fileData, err := dfs.getInternal(key)
	if err != nil {
		return fmt.Errorf("failed to get data: %w", err)
	}

	dateKey := NewDate(clock.Now())
	data, exist := fileData[dateKey]
	if !exist {
		data = []audit.DataAudit{input}
	} else {
		data = append(data, input)
		// sort by event at desc
		sort.Slice(data, func(i, j int) bool {
			return data[i].Metadata.EventAt.After(data[j].Metadata.EventAt)
		})
	}

	// Update the map with the modified slice
	fileData[dateKey] = data

	payload, err := json.Marshal(fileData)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = os.WriteFile(NewFilePath(key).String(), payload, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// getInternal reads data without acquiring lock (for internal use when lock is already held)
func (dfs *DataFileStore) getInternal(key Key) (Data, error) {
	filePath := NewFilePath(key)
	file, err := os.OpenFile(filePath.String(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return Data{}, err
	}
	defer file.Close()

	var data Data
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		// If file is empty (newly created), return empty Data map
		if errors.Is(err, io.EOF) {
			return make(Data), nil
		}
		return Data{}, err
	}

	return data, nil
}

func (dfs *DataFileStore) Get(ctx context.Context, key Key) (Data, error) {
	mu := dfs.mu.GetOrCreate(string(key))
	mu.RLock()
	defer mu.RUnlock()

	filePath := NewFilePath(key)
	file, err := os.OpenFile(filePath.String(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return Data{}, err
	}
	defer file.Close()

	var data Data
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		// If file is empty (newly created), return empty Data map
		if errors.Is(err, io.EOF) {
			return make(Data), nil
		}
		return Data{}, err
	}

	return data, nil
}

func (dfs *DataFileStore) GetAll(ctx context.Context) (map[string]Data, error) {
	output := make(map[string]Data)
	files, err := os.ReadDir("tmp")
	if err != nil {
		return map[string]Data{}, err
	}

	for _, file := range files {
		// Strip .json extension to get the key
		key := strings.TrimSuffix(file.Name(), ".json")
		data, err := dfs.Get(ctx, Key(key))
		if err != nil {
			return map[string]Data{}, err
		}
		output[key] = data
	}

	return output, nil
}

func (dfs *DataFileStore) DeleteAfterDay(ctx context.Context, timeNow time.Time) error {
	files, err := os.ReadDir("tmp")
	if err != nil {
		return err
	}

	for _, dirEntry := range files {
		filePath := fmt.Sprintf("tmp/%s", dirEntry.Name())
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
		if err != nil {
			return err
		}

		var data Data
		err = json.NewDecoder(file).Decode(&data)
		file.Close()
		if err != nil {
			// If file is empty, skip it
			if errors.Is(err, io.EOF) {
				continue
			}
			return err
		}

		if data[NewDate(timeNow)] == nil {
			err = os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
