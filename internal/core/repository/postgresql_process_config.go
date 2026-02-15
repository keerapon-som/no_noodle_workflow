package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/keerapon-som/no_noodle_workflow/internal/core/entitites"
)

func (p *PostgreSQLNoNoodleWorkflow) InsertProcessConfig(tx *sql.Tx, config *entitites.ProcessConfig) error {

	// Marshal maps to JSON
	mapStageTaskJSON, err := json.Marshal(config.MapStageTask)
	if err != nil {
		return err
	}
	mapStageReadyJSON, err := json.Marshal(config.MapStageReady)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO process_config (process_id, map_stage_task, map_stage_ready) VALUES ($1, $2, $3)", config.ProcessID, mapStageTaskJSON, mapStageReadyJSON)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgreSQLNoNoodleWorkflow) GetProcessConfigByProcessID(tx *sql.Tx, ProcessID string) (entitites.ProcessConfig, error) {
	var config entitites.ProcessConfig
	var mapStageTaskJSON []byte
	var mapStageReadyJSON []byte

	err := tx.QueryRow("SELECT process_id, map_stage_task, map_stage_ready FROM process_config WHERE process_id = $1", ProcessID).Scan(&config.ProcessID, &mapStageTaskJSON, &mapStageReadyJSON)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(mapStageTaskJSON, &config.MapStageTask)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(mapStageReadyJSON, &config.MapStageReady)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (p *PostgreSQLNoNoodleWorkflow) GetMapStageTaskByProcessID(tx *sql.Tx, ProcessID string) (map[string][]string, error) {
	var mapStageTaskJSON []byte
	err := tx.QueryRow("SELECT map_stage_task FROM process_config WHERE process_id = $1", ProcessID).Scan(&mapStageTaskJSON)
	if err != nil {
		return nil, err
	}

	var mapStageTask map[string][]string
	err = json.Unmarshal(mapStageTaskJSON, &mapStageTask)
	if err != nil {
		return nil, err
	}
	return mapStageTask, nil
}
