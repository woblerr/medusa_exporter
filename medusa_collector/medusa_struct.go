package medusa_collector

//	[{
//	  "backup_type": "string",
//	  "completed_nodes": number,
//	  "finished": number,
//	  "incomplete_nodes": number,
//	  "incomplete_nodes_list": [{}],
//	  "missing_nodes": number,
//	  "missing_nodes_list": [],
//	  "name": "string",
//	  "nodes": [{}],
//	  "num_objects": number,
//	  "size": number,
//	  "started": number,
//	}]
type backup struct {
	BackupType          string   `json:"backup_type"`
	CompletedNodes      int      `json:"completed_nodes"`
	Finished            int64    `json:"finished"`
	IncompleteNodes     int      `json:"incomplete_nodes"`
	IncompleteNodesList []node   `json:"incomplete_nodes_list"`
	MissingNodes        int      `json:"missing_nodes"`
	MissingNodesList    []string `json:"missing_nodes_list"`
	Name                string   `json:"name"`
	Nodes               []node   `json:"nodes"`
	NumObjects          int64    `json:"num_objects"`
	Size                int64    `json:"size"`
	Started             int64    `json:"started"`
}

//	"nodes": [{
//	  "finished": number,
//	  "fqdn": "string",
//	  "num_objects": number,
//	  "release_version": "string",
//	  "server_type": "string",
//	  "size": number,
//	  "started": number
//	}]
type node struct {
	Finished       int64  `json:"finished"`
	FQDN           string `json:"fqdn"`
	NumObjects     int64  `json:"num_objects"`
	ReleaseVersion string `json:"release_version"`
	ServerType     string `json:"server_type"`
	Size           int64  `json:"size"`
	Started        int64  `json:"started"`
}
