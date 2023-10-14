package hephaestus

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-kafka/connect"
)

func (s *suite) addTableToCapturedTableListInSourceConnector(ctx context.Context, tableName string) (context.Context, error) {
	stepstate := StepStateFromContext(ctx)

	fileName := s.SourceConnectorFileName
	path := filepath.Join(s.SourceConnectorDir, fileName)

	b, err := os.ReadFile(path)
	if err != nil {
		return ctx, err
	}

	sourceCfg := connect.Connector{}
	err = json.Unmarshal(b, &sourceCfg)
	if err != nil {
		return ctx, err
	}
	tableList := sourceCfg.Config["table.include.list"]
	tableList += fmt.Sprintf(",public.%s", tableName)
	sourceCfg.Config["table.include.list"] = tableList

	wb, err := json.MarshalIndent(sourceCfg, "", " ")
	if err != nil {
		return ctx, err
	}

	err = os.WriteFile(path, wb, fs.FileMode(0666))
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepstate), nil
}
