// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package ovnnb

const ConnectionTable = "Connection"

// Connection defines an object in Connection table
type Connection struct {
	UUID            string            `ovsdb:"_uuid"`
	ExternalIDs     map[string]string `ovsdb:"external_ids"`
	InactivityProbe *int              `ovsdb:"inactivity_probe"`
	IsConnected     bool              `ovsdb:"is_connected"`
	MaxBackoff      *int              `ovsdb:"max_backoff"`
	OtherConfig     map[string]string `ovsdb:"other_config"`
	Status          map[string]string `ovsdb:"status"`
	Target          string            `ovsdb:"target"`
}
