package catalog

import (
	"strings"

	"github.com/plexusone/systemspec-architecture/sas"
)

// httpOperations maps HTTP methods to the generic verb they specialize.
var httpOperations = map[string]sas.Operation{
	"GET":     sas.OperationRead,
	"HEAD":    sas.OperationRead,
	"OPTIONS": sas.OperationDiscover,
	"POST":    sas.OperationCreate,
	"PUT":     sas.OperationUpdate,
	"PATCH":   sas.OperationUpdate,
	"DELETE":  sas.OperationDelete,
}

// HTTPOperation maps an HTTP method (case-insensitive) to the generic
// verb it specializes, e.g. "GET" -> OperationRead. The bool result is
// false for methods with no natural generic-verb mapping.
func HTTPOperation(method string) (sas.Operation, bool) {
	op, ok := httpOperations[strings.ToUpper(method)]
	return op, ok
}

// sqlOperations maps SQL statement keywords to the generic verb they
// specialize.
var sqlOperations = map[string]sas.Operation{
	"SELECT": sas.OperationRead,
	"INSERT": sas.OperationCreate,
	"UPDATE": sas.OperationUpdate,
	"DELETE": sas.OperationDelete,
	"MERGE":  sas.OperationUpdate,
	"CREATE": sas.OperationAdminister,
	"ALTER":  sas.OperationAdminister,
	"DROP":   sas.OperationAdminister,
	"GRANT":  sas.OperationAdminister,
	"REVOKE": sas.OperationAdminister,
}

// SQLOperation maps a SQL statement keyword (case-insensitive, e.g.
// "SELECT" or the first token of a statement) to the generic verb it
// specializes. The bool result is false for keywords with no natural
// generic-verb mapping.
func SQLOperation(statement string) (sas.Operation, bool) {
	op, ok := sqlOperations[strings.ToUpper(statement)]
	return op, ok
}

// mcpOperations maps MCP action names to the generic verb they
// specialize.
var mcpOperations = map[string]sas.Operation{
	"tools/list":     sas.OperationDiscover,
	"tools/call":     sas.OperationExecute,
	"invoke_tool":    sas.OperationExecute,
	"resources/list": sas.OperationDiscover,
	"resources/read": sas.OperationRead,
}

// MCPOperation maps an MCP action name to the generic verb it
// specializes, e.g. "tools/call" -> OperationExecute. The bool result is
// false for actions with no natural generic-verb mapping.
func MCPOperation(action string) (sas.Operation, bool) {
	op, ok := mcpOperations[action]
	return op, ok
}
