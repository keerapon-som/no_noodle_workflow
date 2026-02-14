package entitites

type ProcessConfig struct {
	ProcessID     string              `json:"process_id"`
	MapStageTask  map[string][]string `json:"map_stage_task"`
	MapStageReady map[string][]string `json:"map_stage_ready"`
}
